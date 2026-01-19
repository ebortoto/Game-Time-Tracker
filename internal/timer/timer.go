package timer

import "time"

type Stopwatch struct {
	inicio    time.Time
	acumulado time.Duration
	Rodando   bool
}

// Start começa ou retoma a contagem
func (s *Stopwatch) Start() {
	if !s.Rodando {
		s.inicio = time.Now()
		s.Rodando = true
	}
}

// Pause para a contagem e salva o tempo decorrido
func (s *Stopwatch) Pause() {
	if s.Rodando {
		s.acumulado += time.Since(s.inicio)
		s.Rodando = false
	}
}

// Elapsed retorna o tempo total (acumulado + atual se estiver rodando)
func (s *Stopwatch) Elapsed() time.Duration {
	if s.Rodando {
		return s.acumulado + time.Since(s.inicio)
	}
	return s.acumulado
}
