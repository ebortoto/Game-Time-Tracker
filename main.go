package main

import (
	"fmt"
	"image/color"
	"time"

	"game-time-tracker/internal/detector"
	"game-time-tracker/internal/timer"
	"game-time-tracker/internal/ui"
)

func main() {
	// 1. Configuração
	listaDeJogos := []string{"notepad.exe", "cs2.exe", "valorant.exe"}
	monitor := detector.New(listaDeJogos)
	cronometros := make(map[string]*timer.Stopwatch)

	// 2. Inicializa a Interface (UI)
	overlay := ui.NewOverlay()

	// 3. Inicia o Loop de Detecção em PARALELO (Goroutine)
	go func() {
		// Loop infinito de monitoramento
		ticker := time.NewTicker(1 * time.Second)

		for range ticker.C {
			encontrado, nomeJogo, emFoco := monitor.Scan()

			// Tenta forçar a janela a ficar no topo a cada ciclo (garantia)
			detector.SetAlwaysOnTop("GameTimerOverlay")

			if encontrado {
				// Gerencia Cronômetro
				if _, existe := cronometros[nomeJogo]; !existe {
					cronometros[nomeJogo] = &timer.Stopwatch{}
				}
				relogio := cronometros[nomeJogo]

				if emFoco {
					relogio.Start()
					texto := fmt.Sprintf("%s\n%s", nomeJogo, formatarTempo(relogio.Elapsed()))
					// Atualiza UI com cor VERDE
					overlay.UpdateText(texto, color.RGBA{R: 0, G: 255, B: 0, A: 255})
				} else {
					relogio.Pause()
					texto := fmt.Sprintf("%s (Pausa)\n%s", nomeJogo, formatarTempo(relogio.Elapsed()))
					// Atualiza UI com cor AMARELA
					overlay.UpdateText(texto, color.RGBA{R: 255, G: 255, B: 0, A: 255})
				}
			} else {
				// Pausa tudo se nada encontrado
				for _, t := range cronometros {
					if t.Rodando {
						t.Pause()
					}
				}
				overlay.UpdateText("Buscando Jogo...", color.White)
			}
		}
	}()

	// 4. Mostra a Janela e Roda o App (Bloqueia o main até fechar)
	fmt.Println("Overlay Iniciado. Verifique o canto da tela.")
	overlay.Window.ShowAndRun()
}

func formatarTempo(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
