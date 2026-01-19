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

	processos, err := process.Processes()
	if err != nil {
		return false, "", false
	}

	for _, p := range processos {
		nome, err := p.Name()
		if err != nil {
			continue
		}

		nomeMinusculo := strings.ToLower(nome)

		for _, jogo := range c.TargetGames {
			if nomeMinusculo == jogo {
				// Jogo Encontrado!

				// Verifica se o PID do jogo é igual ao PID da janela em foco
				estaEmFoco := false
				if uint32(p.Pid) == focoPID {
					estaEmFoco = true
				}

				return true, nome, estaEmFoco
			}
		}
	}

	return false, "", false
}
