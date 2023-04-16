package api

import (
	"fmt"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

type jaegerLogger struct {
}

func initLogger() {
	setLogLevel(util.GetEnvWithFallback("LOG_LEVEL", "debug"))
	log.AddHook(&jaegerLogger{})
}
func setLogLevel(logLevel string) {
	switch logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.DebugLevel)
		log.Errorf("Log level %s unknow", logLevel)
	}

}

func (jaeger *jaegerLogger) Levels() []log.Level {
	return []log.Level{log.TraceLevel, log.DebugLevel, log.InfoLevel, log.WarnLevel, log.ErrorLevel, log.FatalLevel, log.PanicLevel}
}

func (jaeger *jaegerLogger) Fire(entry *log.Entry) error {
	contextData := entry.Data["context"]
	if contextData == nil {
		return nil
	}
	context := contextData.(echo.Context)
	sp := jaegertracing.CreateChildSpan(context, fmt.Sprintf("%s log", entry.Level.String()))
	defer sp.Finish()
	sp.LogKV(
		"time", entry.Time,
		"level", entry.Level.String(),
		"message", entry.Message,
	)
	return nil
}
