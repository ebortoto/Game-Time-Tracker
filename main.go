package main

import (
	"fmt"
	"time"

	"game-time-tracker/internal/detector"
	"game-time-tracker/internal/timer"
	"game-time-tracker/internal/ui"
)

func main() {
	// 1. Configuração
	// DICA: Adicione "notepad.exe" ou "calc.exe" (se for a antiga) para testar fácil
	listaDeJogos := []string{"notepad.exe", "cs2.exe", "valorant.exe", "calculatorapp.exe", "PapersPlease.exe"}
	monitor := detector.New(listaDeJogos)
	cronometros := make(map[string]*timer.Stopwatch)

	// 2. Inicializa a Interface
	ui.InitOverlay()

	// 3. Loop principal: mantém o processo vivo e atualiza o monitor a cada segundo.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		encontrado, nomeJogo, emFoco := monitor.Scan()
		detector.SetAlwaysOnTop("GameTimerOverlay")

		if encontrado {
			if _, existe := cronometros[nomeJogo]; !existe {
				cronometros[nomeJogo] = &timer.Stopwatch{}
			}
			relogio := cronometros[nomeJogo]

			if emFoco {
				relogio.Start()
				texto := fmt.Sprintf("[JOGANDO]\n%s\n%s", nomeJogo, formatarTempo(relogio.Elapsed()))
				ui.UpdateText(texto)
			} else {
				relogio.Pause()
				texto := fmt.Sprintf("[PAUSA]\n%s\n%s", nomeJogo, formatarTempo(relogio.Elapsed()))
				ui.UpdateText(texto)
			}
		} else {
			for _, t := range cronometros {
				if t.Rodando {
					t.Pause()
				}
			}
			ui.UpdateText("Aguardando Jogo...")
		}
	}
}

func formatarTempo(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
