package main

import (
	"log/slog"
	"marat/fayz/kafka_reader_writer/internal/app"
	"os"
)

func main() {
	// config := &config.Config{
	// 	username:  "SFA",
	// 	password:  "SFADEV123",
	// 	broker:    "172.16.15.171:9093",
	// 	topic:     "SFA_MORTG_DEALCALL_TEST",
	// 	partition: 0,
	// 	// certificateFilePath: "./bmc.crt",
	// 	certificateFilePath: "./certificate.pem",
	// }

	a := app.New()

	if err := a.Run(); err != nil {
		slog.Error("ошибка приложения", "err", err)
		os.Exit(1)
	}
}
