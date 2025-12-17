package helpers

import (
	"fmt"
	"math"
	"strings"
)

/*
// GerarTaxas gera N taxas distribuídas uniformemente entre taxaMin e taxaMax
//Fazer a cada metade ex: tx minima é 20
func GerarTaxas(taxaMin, taxaMax float64, numCenarios int) []float64 {
	fmt.Println("Taxa Min", taxaMin, "Taxa Max", taxaMax, "Num Cenarios", numCenarios)
	if numCenarios == 1 {
		return []float64{taxaMin}
	}

	taxas := make([]float64, numCenarios)
	intervalo := (taxaMax - taxaMin) / float64(numCenarios-1)

	for i := range numCenarios {
		taxa := taxaMin + (float64(i) * intervalo)
		taxa = math.Round(taxa*100) / 100
		fmt.Println("Taxa Gerada", taxa)
		taxas[i] = taxa

	}

	return taxas
}
*/

func GerarTaxas(taxaMin, taxaMax, intervalo float64) []float64 {

	taxas := make([]float64, 0)
	taxas = append(taxas, taxaMax)
	//intervalo := (taxaMax - taxaMin) / float64(numCenarios-1)
	counter := taxaMax
	for {
		if counter <= taxaMin {
			break
		}
		counter -= intervalo
		counter = math.Round(counter*100) / 100
		taxas = append(taxas, counter)

	}
	return taxas

}

// CalcularPrazo calcula o número de parcelas usando a fórmula Price
// Intervalo representa o intervalo em dias para os casos em que ele não é mensal
// Ao passar 0 ou qualquer valor abaixo, a função considerará um intervalo mensal
func CalcularPrazoPrice(valorEmprestimo, valorParcela, taxaJurosMensal float64, intervalo int) (int, error) {
	fmt.Println("ValorEmprestimo", valorEmprestimo, "ValorParcela", valorParcela, "TaxaMensal", taxaJurosMensal)
	// Converter taxa de % para decimal

	i := taxaJurosMensal / 100.0

	//Para o caso de intervalos que não são mensais
	if intervalo > 0 {
		//	i = math.Pow(1+i/100, 1.0/float64(intervalo)) - 1
		i = math.Pow(1+i, 1.0/float64(intervalo)) - 1

	}

	// Verificar se parcela é suficiente
	jurosPrimeiroMes := valorEmprestimo * i
	if valorParcela <= jurosPrimeiroMes {
		return 0, fmt.Errorf("parcela insuficiente: R$ %.2f não cobre juros de R$ %.2f",
			valorParcela, jurosPrimeiroMes)
	}

	// Calcular prazo usando fórmula Price
	// n = log(PMT / (PMT - PV * i)) / log(1 + i)
	numerador := math.Log(valorParcela / (valorParcela - valorEmprestimo*i))
	denominador := math.Log(1 + i)
	n := numerador / denominador
	fmt.Println("Numerador:", numerador, "Denominador", denominador, "Numerador/Denominador:", n)

	// Arredondar
	prazo := int(math.Round(n))
	return prazo, nil
}

func CalcularPrazo(valorDesejado, parcelaAlvo, taxaMin, taxaMax, tacPercentual, tacValorFixo, IOF float64, prazo, diasEntreParcelas, diasAtePagamento int) (float64, float64) {
	/*	valorDesejado := 1000.00
		parcelaAlvo := 312.00
		taxaMin := 0.15 // 1% ao mês
		taxaMax := 0.40 // 20% ao mês
		prazo := 5      // Prazo fixo em dias
	*/
	diasCarencia := float64(diasAtePagamento)
	//fmt.Println("Valor", valorDesejado, "TAC", tacValorFixo, "Parcela", parcelaAlvo, "Taxa Minima", taxaMin, "Taxa Maxima", taxaMax, "Prazo", prazo, "IOF", IOF, "DiasEntreParcelas", diasEntreParcelas, "DiasCarencia", diasCarencia)

	//tacPercentual = tacPercentual / 100
	taxaMin = taxaMin / 100
	taxaMax = taxaMax / 100
	//fmt.Println("TaxaMInDec", taxaMin, "TaxaMaxDec", taxaMax)

	duracaoTotalDias := float64(diasEntreParcelas * prazo)
	iof := IOF + (0.000082 * duracaoTotalDias)
	numParcelas := duracaoTotalDias / float64(diasEntreParcelas)
	periodosPorMes := 30.0 / float64(diasEntreParcelas)
	periodosCarencia := diasCarencia / float64(diasEntreParcelas)

	melhorDiferenca := math.MaxFloat64
	var melhorTaxa float64
	var melhorParcela float64

	for taxa := taxaMin; taxa <= taxaMax; taxa += 0.0001 {
		jurosPorPeriodo := math.Pow(1+taxa, 1/periodosPorMes) - 1
		fatorPrice := jurosPorPeriodo / (1 - math.Pow(1+jurosPorPeriodo, -numParcelas))

		valorCorrigido := valorDesejado * math.Pow(1+jurosPorPeriodo, periodosCarencia)
		parcela := fatorPrice * valorCorrigido * (1 + iof + tacPercentual)

		if tacValorFixo > 0 {
			parcela = fatorPrice * valorCorrigido * (1 + iof + (tacValorFixo / valorCorrigido))
		}

		diferenca := math.Abs(parcela - parcelaAlvo)
		if diferenca < melhorDiferenca {
			melhorDiferenca = diferenca
			melhorTaxa = taxa
			melhorParcela = parcela
		}
	}

	//fmt.Printf("Melhor combinação encontrada:\n")
	//fmt.Printf("Prazo fixo: %d dias\n", prazo)
	//fmt.Printf("Taxa mensal: %.4f\n", melhorTaxa)
	//fmt.Printf("Parcela estimada: R$ %.2f\n", melhorParcela)
	melhorTaxa *= 100
	melhorTaxa = math.Round(melhorTaxa*100) / 100

	return melhorTaxa, melhorParcela
}

/*func TestCalcularPrazoVariandoTaxa() {
	valorEmprestimo := 8900.00 + 10
	valorParcela := 2500.00
	intervalo := 0 // mensal

	// Testar várias taxas de juros
	taxas := GerarTaxas(15, 30, 0.5)
	fmt.Println("\n Gerando Prazos \n")
	fmt.Println("Taxa\tPrazo")
	for _, taxa := range taxas {
		prazo, err := CalcularPrazoPrice(valorEmprestimo, valorParcela, taxa, intervalo)
		if err != nil {
			fmt.Printf("%.3f\tErro: %v\n", taxa, err)
		} else {
			fmt.Printf("%.3f\t%d\n", taxa, prazo)
		}
	}
}*/

func GerarValorSolicitado(parcela, jurosMensal, tacPercentual, tacValorFixo, vlrIOF float64, diasIntervaloPrazo, numeroDiasAcreascimo, prazo int) float64 {
	fmt.Println("VLR PARCELA", parcela, "jurosMensal", jurosMensal, "IOF", vlrIOF, "PERC TAC", tacPercentual, "VLR TAC", tacValorFixo, "DIAS INTERVALO", diasIntervaloPrazo, "DIAS ACRESCIMO", numeroDiasAcreascimo, "Prazo", prazo)

	tacPercentual = tacPercentual / 100

	jurosMensal = jurosMensal / 100

	// Frequência da parcela

	duracaoTotalDias := float64(diasIntervaloPrazo * prazo)

	// Carência
	diasCarencia := float64(numeroDiasAcreascimo)

	// IOF
	iof := vlrIOF + (0.000082 * duracaoTotalDias)

	// Conversão da taxa mensal para taxa por período
	periodosPorMes := 30.0 / float64(diasIntervaloPrazo)
	jurosPorPeriodo := math.Pow(1+jurosMensal, 1/periodosPorMes) - 1

	// Número de parcelas
	numParcelas := duracaoTotalDias / float64(diasIntervaloPrazo)

	// Fator PRICE
	fatorPrice := jurosPorPeriodo / (1 - math.Pow(1+jurosPorPeriodo, -numParcelas))

	// Se TAC for valor fixo, recalcular com base nele
	// Estimar valor solicitado base (sem TAC fixo ainda)
	valorSolicitadoBase := parcela / (fatorPrice * (1 + iof + tacPercentual))

	if tacValorFixo > 0 {
		valorSolicitadoBase = parcela / (fatorPrice * (1 + iof + (tacValorFixo / valorSolicitadoBase)))
	}

	// Aplicar carência
	periodosCarencia := diasCarencia / float64(diasIntervaloPrazo)
	valorSolicitadoCorrigido := valorSolicitadoBase / math.Pow(1+jurosPorPeriodo, periodosCarencia)
	fmt.Println("VALOR CORRIGIDO", valorSolicitadoCorrigido)
	return math.Round(valorSolicitadoCorrigido*100) / 100
}

func GerarValorTac(valorSolicitado, PercTAC float64) float64 {
	valorTac := valorSolicitado * (PercTAC / 100)
	return valorTac
}

/*
func main() {
	parcela := 45.00
	jurosMensal := 0.15

	// Frequência da parcela
	diasEntreParcelas := 1
	prazo := 7
	duracaoTotalDias := float64(diasEntreParcelas * prazo)

	// Carência
	diasCarencia := 0.00

	// TAC: pode ser percentual ou valor fixo
	tacPercentual := 0.187 // ex: 18.7%
	tacValorFixo := 17.50  // ex: R$200.00

	// IOF
	iof := 0.0038 + (0.000082 * duracaoTotalDias)

	// Conversão da taxa mensal para taxa por período
	periodosPorMes := 30.0 / float64(diasEntreParcelas)
	jurosPorPeriodo := math.Pow(1+jurosMensal, 1/periodosPorMes) - 1

	// Número de parcelas
	numParcelas := duracaoTotalDias / float64(diasEntreParcelas)

	// Fator PRICE
	fatorPrice := jurosPorPeriodo / (1 - math.Pow(1+jurosPorPeriodo, -numParcelas))

	// Estimar valor solicitado base (sem TAC fixo ainda)
	valorSolicitadoBase := parcela / (fatorPrice * (1 + iof + tacPercentual))

	// Se TAC for valor fixo, recalcular com base nele
	if tacValorFixo > 0 {
		valorSolicitadoBase = parcela / (fatorPrice * (1 + iof + (tacValorFixo / valorSolicitadoBase)))
	}

	// Aplicar carência
	periodosCarencia := diasCarencia / float64(diasEntreParcelas)
	valorSolicitadoCorrigido := valorSolicitadoBase / math.Pow(1+jurosPorPeriodo, periodosCarencia)

	fmt.Printf("Valor Solicitado estimado com carência de %.0f dias: R$ %.2f\n", diasCarencia, valorSolicitadoCorrigido)
}
*/

/*

func GerarValorSolicitado2(parcela, jurosMensal, tacPercentual, tacValorFixo, vlrIOF float64, diasIntervaloPrazo, numeroDiasAcreascimo, prazo int) float64 {
	if tacPercentual == 0 {
		tacPercentual = 1
	}
	jurosMensal = jurosMensal / 100

	// Frequência da parcela

	duracaoTotalDias := float64(diasIntervaloPrazo * prazo)

	// Carência
	diasCarencia := float64(numeroDiasAcreascimo)

	// IOF
	iof := vlrIOF + (0.000082 * duracaoTotalDias)

	// Conversão da taxa mensal para taxa por período
	periodosPorMes := 30.0 / float64(diasIntervaloPrazo)
	jurosPorPeriodo := math.Pow(1+jurosMensal, 1/periodosPorMes) - 1

	// Número de parcelas
	numParcelas := duracaoTotalDias / float64(diasIntervaloPrazo)

	// Fator PRICE
	fatorPrice := jurosPorPeriodo / (1 - math.Pow(1+jurosPorPeriodo, -numParcelas))

	// Se TAC for valor fixo, recalcular com base nele
	// Estimar valor solicitado base (sem TAC fixo ainda)
	valorSolicitadoBase := parcela / (fatorPrice * (1 + iof + tacPercentual))

	if tacValorFixo > 0 {
		valorSolicitadoBase = parcela / (fatorPrice * (1 + iof + (tacValorFixo / valorSolicitadoBase)))
	}

	// Aplicar carência
	periodosCarencia := diasCarencia / float64(diasIntervaloPrazo)
	valorSolicitadoCorrigido := valorSolicitadoBase / math.Pow(1+jurosPorPeriodo, periodosCarencia)
	fmt.Println("VALOR CORRIGIDO", valorSolicitadoCorrigido)
	return math.Round(valorSolicitadoCorrigido*100) / 100
}

func GerarValorTac(valorSolicitado, PercTAC float64) float64 {
	valorTac := valorSolicitado * (PercTAC / 100)
	return valorTac
}


*/

func ArrayMetade(arr []float64) int {
	l := len(arr)
	if l == 1 {
		return 0
	}

	size := float64(l)
	if l%2 == 0 {
		return int((size / 2)) - 1
	} else {
		return int(math.Ceil(size/2)) - 1
	}

}

//

func IntSliceToSQLIn(slice []int) string {
	strs := make([]string, len(slice))
	for i, v := range slice {
		strs[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(strs, ",")
}
