package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

type Overlay struct {
	App    fyne.App
	Window fyne.Window
	Label  *canvas.Text // Usamos canvas.Text para ter mais controle de cor/tamanho
}

func NewOverlay() *Overlay {
	a := app.New()
	w := a.NewWindow("GameTimerOverlay") // Esse título é usado pelo windows.go para achar a janela

	// Configuração Visual
	w.SetDecored(false) // Remove barra de título, botão fechar, bordas (fica flutuante)
	w.Resize(fyne.NewSize(200, 50))

	// Criando o texto
	texto := canvas.NewText("Aguardando...", color.White)
	texto.TextSize = 20 // Tamanho da fonte
	texto.TextStyle = fyne.TextStyle{Bold: true}
	texto.Alignment = fyne.TextAlignCenter

	// Fundo preto semi-transparente para ler melhor
	fundo := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 200})

	// Organiza layout (Fundo atrás, Texto na frente)
	conteudo := container.New(layout.NewMaxLayout(), fundo, container.NewCenter(texto))
	w.SetContent(conteudo)

	// Posicionar no canto superior direito (Gambiarra simples)
	// O ideal seria pegar a resolução da tela, mas fixaremos uma posição inicial
	w.Move(fyne.NewPos(1500, 50)) // Ajuste esses valores conforme sua tela (X, Y)

	return &Overlay{
		App:    a,
		Window: w,
		Label:  texto,
	}
}

// UpdateText atualiza o texto da interface de forma segura
func (o *Overlay) UpdateText(texto string, cor color.Color) {
	o.Label.Text = texto
	o.Label.Color = cor
	o.Label.Refresh()
}
