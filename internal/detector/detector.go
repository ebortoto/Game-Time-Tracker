package detector

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type Config struct {
	TargetGames []string
}

func New(games []string) *Config {
	for i, g := range games {
		games[i] = strings.ToLower(g)
	}
	sort.Strings(games)
	return &Config{TargetGames: games}
}

// Scan agora retorna: (encontrado bool, nome string, emFoco bool)
func (c *Config) Scan() (bool, string, bool) {
	// 1. Descobrir qual o PID da janela que está em foco AGORA
	focoPID, err := getForegroundPID()
	if err != nil {
		fmt.Println("Erro ao obter janela em foco:", err)
	}

	// 1.1 Primeiro, resolve diretamente o processo em foco para evitar flapping
	// quando há múltiplas instâncias/processos auxiliares do mesmo jogo.
	if focoPID != 0 {
		procFoco, err := process.NewProcess(int32(focoPID))
		if err == nil {
			nomeFoco, err := procFoco.Name()
			if err == nil {
				nomeFocoLower := strings.ToLower(nomeFoco)
				for _, jogo := range c.TargetGames {
					if nomeFocoLower == jogo {
						return true, nomeFoco, true
					}
				}
			}
		}
	}

	processos, err := process.Processes()
	if err != nil {
		return false, "", false
	}

	// Como o processo em foco já foi validado acima, aqui basta descobrir
	// se algum jogo alvo está em execução (estado pausado).
	encontrouJogo := false
	nomeEncontrado := ""

	for _, p := range processos {
		nome, err := p.Name()
		if err != nil {
			continue
		}

		nomeMinusculo := strings.ToLower(nome)

		for _, jogo := range c.TargetGames {
			if nomeMinusculo == jogo {
				encontrouJogo = true
				if nomeEncontrado == "" {
					nomeEncontrado = nome
				}
				break
			}
		}
	}

	if encontrouJogo {
		return true, nomeEncontrado, false
	}

	return false, "", false
}
