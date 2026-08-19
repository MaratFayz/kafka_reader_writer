package debug

import (
	"log/slog"
)

func WriteDebugFile(text string) {
	slog.Info(text)
}
