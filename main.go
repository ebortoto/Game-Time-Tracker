package main

import (
	"fmt"
	"time"

	apptracking "game-time-tracker/internal/application/tracking"
	infraoverlay "game-time-tracker/internal/infrastructure/overlay"
	infraruntime "game-time-tracker/internal/infrastructure/runtime"
	infrascanner "game-time-tracker/internal/infrastructure/scanner"
)

func main() {
	releaseLock, alreadyRunning, err := infraruntime.AcquireSingleInstance()
	if err != nil {
		fmt.Println("Erro ao iniciar lock de instancia unica:", err)
		return
	}
	if alreadyRunning {
		fmt.Println("Ja existe outra instancia do Game Time Tracker em execucao.")
		return
	}
	defer releaseLock()

	// 1. Configuração
	// DICA: Adicione "notepad.exe" ou "calc.exe" (se for a antiga) para testar fácil
	listaDeJogos := []string{"PapersPlease.exe"}
	scanner := infrascanner.NewProcessScanner(listaDeJogos)
	overlay := infraoverlay.NewRTSSOverlay()
	service := apptracking.NewService(scanner, overlay)

	// 2. Inicializa a Interface
	overlay.Init()
	defer overlay.Close()

	// 3. Loop principal: mantém o processo vivo e atualiza o monitor a cada segundo.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		service.Tick()
	}
}
