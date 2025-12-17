package helpers

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Valida structs de acordo com a tag "validate". Validações customizadas devem ser passadas em "extraTags"
func StructValidate(strc any, extraTags ...string) error {
	v := validator.New()

	translator, _ := NewTranslator()
	ConfigTranslator(v, translator, strc)

	if len(extraTags) > 0 {
		registerCustomValidation(v, extraTags...)
	}

	if errs := v.Struct(strc); errs != nil {
		errFields := NewErrValidation()

		for _, err := range errs.(validator.ValidationErrors) {
			tag := ExtractStructTag(err.Field(), "json", strc, true)
			errorTranslated := err.Translate(translator)
			errFields.AddField(tag, errorTranslated)
		}

		return errFields

	}
	return nil
}

// Valida structs e retorna os erros no padrão JRPC.
func StructValidateJRPC(strc any, errCache Cache, id any, extraTags ...string) error {
	v := validator.New()

	translator, _ := NewTranslator()
	ConfigTranslator(v, translator, strc)

	if len(extraTags) > 0 {
		registerCustomValidation(v, extraTags...)
	}

	if errs := v.Struct(strc); errs != nil {
		responses := make(JRPCResponses, 0)

		for _, err := range errs.(validator.ValidationErrors) {
			tag := ExtractStructTag(err.Field(), "json", strc, true)
			errorTranslated := err.Translate(translator)
			response := NewJRPCResponse(errCache, tag, errorTranslated, id)
			response.Err.Field = tag
			responses = append(responses, response)

		}

		return responses

	}
	return nil
}

// Extrai a tag especificada em "tag" do campo "field" na struct "model".
// Caso "onlyFirst" seja true, retorna apenas a primeira palavra da tag(o nome do campo).
//É realizada uma verificação se a struct passada em "model" é um ponteiro, pois assim pode se pegar o valor por ele apontado e evitar erros.

func ExtractStructTag(fieldName, tagName string, model any, onlyFirst bool) string {
	var structType reflect.Type
	val := reflect.ValueOf(model)

	if val.Kind() == reflect.Ptr {
		structType = val.Elem().Type()
	} else {
		structType = reflect.TypeOf(model)
	}

	field, found := structType.FieldByName(fieldName)
	if !found {
		panic(fmt.Sprintf("campo %s nao encontrado!", fieldName))
	}

	tag := field.Tag.Get(tagName)
	if onlyFirst {
		tags := strings.Split(tag, ",")
		if len(tags) > 1 {
			tag = tags[0]
		}
	}
	return tag
}

/*
Retorna um mapa tendo como chave o nome dos campos e como valor o valor armazenado nos campos da struct passada em strct.
São considerados apenas campos importados.
Se optional key não for vazia, será extraída a tag passada por ela nas chaves do mapa ao invés do nome do campo.
Se fieldToIgnore for passado, os campos com o nome dentro do slice serão ignorados.
*/
func StructToMap[V any](strct any, optionalKey string, fieldsToIgnore ...string) map[string]V {
	dict := make(map[string]V)
	typ := reflect.TypeOf(strct)
	val := reflect.ValueOf(strct)
	var key string
	var value V

	for i := 0; i < typ.NumField(); i++ {

		field := typ.Field(i)
		if slices.Contains(fieldsToIgnore, field.Name) {
			continue
		}
		if optionalKey != "" {
			key = ExtractStructTag(field.Name, "json", strct, true)
		} else {
			key = field.Name

		}
		value = val.Field(i).Interface().(V)
		dict[key] = value

	}
	return dict
}
