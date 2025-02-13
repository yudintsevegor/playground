package main

import (
	"log/slog"
)

func main() {

	f()
	return
	scientists, err := readDefinitionFiles()
	if err != nil {
		slog.Error("error reading definition files", slog.Any("error", err))
	}

	slog.Info("Scientists", slog.Any("scientists", scientists))
}
