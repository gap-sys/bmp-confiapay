package helpers

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var LogMetrics = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "bmp_log_metrics",
	Help: "Logs generated",
}, []string{"level", "status"})

// customCallerHandler envolve um slog.Handler e fornece a funcionalidade de pular frames do chamador.
type customCallerHandler struct {
	slog.Handler
	callerSkip int // Frames adicionais para pular a partir da chamada direta do método slog.Logger
}

// newCustomCallerHandler cria um novo CustomCallerHandler.
// O parâmetro 'callerSkip' especifica quantos frames devem ser pulados *além* da chamada
// direta do método slog.Logger para alcançar o ponto de iniciação do log no código do usuário.
// Por exemplo:
//   - Se você chamar logger.Info() diretamente, callerSkip deve ser 0.
//   - Se você tiver uma função wrapper MyLog.Info() que chama logger.Info(), e você quer que a
//     origem aponte para o chamador de MyLog.Info(), então callerSkip deve ser 1.
func newCustomCallerHandler(h slog.Handler, callerSkip int) *customCallerHandler {
	return &customCallerHandler{Handler: h, callerSkip: callerSkip}
}

// Handle implementa a interface slog.Handler.
func (h *customCallerHandler) Handle(ctx context.Context, r slog.Record) error {

	// A 'r.PC' (Program Counter) geralmente é definida pelo slog se AddSource for true.
	// Como definimos AddSource: false, r.PC será 0.
	// Precisamos obter manualmente as informações do chamador.

	// Rastreamento da pilha a partir de runtime.Caller(0) quando chamado dentro deste método Handle:
	// 0: runtime.Callers
	// 1: CustomCallerHandler.Handle
	// 2: slog.Logger.log (método interno do slog)
	// 3: slog.Logger.Info/Debug/Warn/Error (o método slog real chamado pelo usuário ou wrapper)
	// 4: Código do usuário chamando slog.Logger.Info/Debug/Warn/Error (ou uma função wrapper)
	// 5: Ponto de iniciação do log do usuário (se o frame 4 for um wrapper)

	// Queremos chegar ao frame 4 + h.callerSkip.
	// Portanto, o skip total para runtime.Caller deve ser 4 + h.callerSkip.
	const slogInternalFrames = 4 // Frames de runtime.Callers até a chamada direta do método slog.Logger pelo usuário

	pc, file, line, ok := runtime.Caller(slogInternalFrames + h.callerSkip)
	if ok {
		f := runtime.FuncForPC(pc)
		// Adiciona o atributo de origem ao registro.
		// Usamos slog.SourceKey para que genAttrReplaceFunc possa processá-lo.
		r.AddAttrs(slog.Any(slog.SourceKey, &slog.Source{
			Function: f.Name(),
			File:     file,
			Line:     line,
		}))
	}

	return h.Handler.Handle(ctx, r)
}

func (h *customCallerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newCustomCallerHandler(h.Handler.WithAttrs(attrs), h.callerSkip)

}

func (h *customCallerHandler) WithGroup(group string) slog.Handler {
	return newCustomCallerHandler(h.Handler.WithGroup(group), h.callerSkip)

}

// NewECSJSONHandler cria um slog.Handler que formata logs em JSON com campos básicos do ECS.
func NewECSJSONHandler(level slog.Level, loc *time.Location) slog.Handler {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: genAttrReplaceFunc(loc),
	}).WithAttrs([]slog.Attr{{
		Key:   "app",
		Value: slog.StringValue("cobranca-bmp"),
	},
	})

	return newCustomCallerHandler(h, 0)
}

func genAttrReplaceFunc(loc *time.Location) func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		// Renomeia os campos padrão do slog para os equivalentes do ECS
		switch a.Key {
		case slog.TimeKey:
			// O ECS usa "@timestamp". Formato RFC3339Nano é o padrão.
			return slog.String("@timestamp", a.Value.Time().In(loc).Format(time.RFC3339Nano))
		case slog.LevelKey:
			// O ECS usa "log.level"
			return slog.String("log.level", a.Value.String())
		case slog.MessageKey:
			// O ECS usa "message"
			return slog.String("message", a.Value.String())
		case slog.SourceKey:
			// O ECS tem campos como "log.origin.file.name", "log.origin.file.line", "log.origin.function"
			// Para simplificar, podemos deixar o slog.SourceKey como está ou criar um grupo "log.origin"
			// Aqui, vamos apenas renomear para "log.origin.source" para manter a informação.
			// Para uma conformidade mais estrita, você precisaria de um handler customizado.
			if a.Value.Kind() == slog.KindAny {

				if fs, ok := a.Value.Any().(*slog.Source); ok {
					return slog.Group("log.origin",
						slog.String("file.name", fs.File),
						slog.Int("file.line", fs.Line),
						slog.String("function", fs.Function),
					)
				}
			}
			return a // Retorna o atributo original se não for um *slog.Source

		default:
			// Para outros atributos, você pode ter regras específicas.
			// Por exemplo, mapear "requestID" para "http.request.id"
			// case "requestID":
			// 	return slog.String("http.request.id", a.Value.String())
			return a
		}
	}
}

func LogDebug(ctx context.Context, logger *slog.Logger, _ *time.Location, contexto, status, mensagem string, payload any) {
	if !logger.Handler().Enabled(ctx, slog.LevelDebug) {
		return
	}

	logger.LogAttrs(
		ctx,
		slog.LevelDebug,
		mensagem,
		//slog.Time("time", time.Now().In(loc)),
		slog.String("contexto", contexto),
		slog.String("status", status),
		slog.Any("payload", payload),
	)

}

func LogInfo(ctx context.Context, logger *slog.Logger, _ *time.Location, contexto, status, mensagem string, payload any) {
	if !logger.Handler().Enabled(ctx, slog.LevelInfo) {
		return
	}

	logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		mensagem,
		slog.String("contexto", contexto),
		slog.String("status", status),
		slog.Any("payload", payload),
	)

}

func LogInfoCurl(ctx context.Context, logger *slog.Logger, _ *time.Location, contexto, status, mensagem string, payload string) {
	if !logger.Handler().Enabled(ctx, slog.LevelInfo) {
		return
	}

	logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		mensagem,
		slog.String("contexto", contexto),
		slog.String("status", status),
		slog.String("payload", payload),
	)

}

func LogWarn(ctx context.Context, logger *slog.Logger, _ *time.Location, contexto, status, mensagem string, erro, payload any) {
	if !logger.Handler().Enabled(ctx, slog.LevelWarn) {
		return
	}

	logger.LogAttrs(
		ctx,
		slog.LevelWarn,
		mensagem,
		slog.String("contexto", contexto),
		slog.String("status", status),
		slog.Any("erro", erro),
		slog.Any("payload", payload),
	)
}

func LogError(ctx context.Context, logger *slog.Logger, _ *time.Location, contexto, status, mensagem string, erro, payload any) {
	if !logger.Handler().Enabled(ctx, slog.LevelError) {
		return
	}

	logger.LogAttrs(
		ctx,
		slog.LevelError,
		mensagem,
		slog.String("contexto", contexto),
		slog.String("status", status),
		slog.Any("erro", erro),
		slog.Any("payload", payload),
	)

}
