package main

import (
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/app"
	"os"
)

func main() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	handler := slog.NewTextHandler(file, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	a := app.New()

	if err := a.Run(); err != nil {
		slog.Error("ошибка приложения", "err", err)
		os.Exit(1)
	}
}
