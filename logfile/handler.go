package logfile

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"time"
)

type PropostaService interface {
	UpdateAssync(data models.UpdateDbData) (bool, error)
}

// Representa um objeto que processa dados escritos em arquivos.
type LogFileHander struct {
	ctx             context.Context
	files           chan models.DbLogfileData
	logger          *slog.Logger
	location        *time.Location
	propostaService PropostaService
}

func NewLogFileHandler(ctx context.Context, logger *slog.Logger, chFiles chan models.DbLogfileData, location *time.Location, PropostaService PropostaService) *LogFileHander {
	fileHandler := &LogFileHander{
		ctx:             ctx,
		files:           chFiles,
		logger:          logger,
		location:        location,
		propostaService: PropostaService,
	}
	return fileHandler
}

func (l LogFileHander) Produce(path string, data any) error {
	file, err := os.Create(path)
	if err != nil {
		helpers.LogError(l.ctx, l.logger, l.location, "file log", "", "Erro ao criar arquivo", err.Error(), data)
		return err
	}

	fileBytes, err := json.Marshal(data)
	if err != nil {
		helpers.LogError(l.ctx, l.logger, l.location, "file log", "", "Erro ao serializar dados para salvar em arquivo", err.Error(), data)
		return err
	}
	_, err = file.Write(fileBytes)
	if err != nil {
		helpers.LogError(l.ctx, l.logger, l.location, "file log", "", "Erro ao escrever em arquivo", err.Error(), data)
		return err
	}
	l.files <- models.DbLogfileData{FilePath: path, File: file}
	return nil
}

func (l *LogFileHander) Consume() {
	for f := range l.files {
		f.File.Seek(0, io.SeekStart)
		fileBytes, err := io.ReadAll(f.File)
		f.File.Close()
		if err != nil {
			helpers.LogError(l.ctx, l.logger, l.location, "file log", "", "Erro ao abrir arquivo de logs", err.Error(), nil)
			continue
		}

		var payload models.UpdateDbData
		if err := json.Unmarshal(fileBytes, &payload); err != nil {
			helpers.LogError(l.ctx, l.logger, l.location, "file log", "", "Erro ao desserializar arquivo de logs", err.Error(), nil)
			continue
		}
		helpers.LogInfo(l.ctx, l.logger, l.location, "file log", "", "entrou em file log", payload)

		for {
			if noConn, err := l.propostaService.UpdateAssync(payload); err == nil || !noConn {
				break
			}
			time.Sleep(config.DB_QUEUE_DELAY)
		}
		if err := os.Remove(f.FilePath); err != nil {
			helpers.LogError(l.ctx, l.logger, l.location, "file log", "", "Erro ao remover arquivo de logs já processado", err.Error(), f.FilePath)
		}

	}

}

func (l *LogFileHander) Close() {
	close(l.files)

}
