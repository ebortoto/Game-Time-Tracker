package detector

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)

// getForegroundPID retorna o ID do processo (PID) da janela que está em foco no momento
func getForegroundPID() (uint32, error) {
	// 1. Pega o "Handle" (identificador) da janela ativa
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return 0, nil // Nenhuma janela em foco
	}

	// 2. Pede ao Windows o PID dono desse Handle
	var pid uint32
	// A função retorna o ThreadID, mas preenche a variável 'pid' que passamos via ponteiro
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	return pid, nil
}

// DebugGetForegroundPID é apenas um wrapper público para testes no main
func DebugGetForegroundPID() (uint32, error) {
	return getForegroundPID()
}

// --- Adicione isso ao final do arquivo windows.go ---

var (
	procSetWindowPos = user32.NewProc("SetWindowPos")
)

// Constantes mágicas do Windows para controlar janelas
const (
	HWND_TOPMOST uintptr = ^uintptr(0)
	SWP_NOSIZE           = 0x0001
	SWP_NOMOVE           = 0x0002
)

// SetAlwaysOnTop força uma janela (pelo Título) a ficar sobre todas as outras
func SetAlwaysOnTop(tituloJanela string) {
	// 1. Encontrar a janela pelo nome (título)
	// Nota: O Fyne cria janelas com o título que definirmos.
	// Precisamos converter string Go para string C (ptr)
	ptrTitulo, _ := windows.UTF16PtrFromString(tituloJanela)

	hwnd, _, _ := user32.NewProc("FindWindowW").Call(
		0,
		uintptr(unsafe.Pointer(ptrTitulo)),
	)

	if hwnd == 0 {
		return // Janela não encontrada ainda
	}

	// 2. Aplicar a flag "TopMost"
	procSetWindowPos.Call(
		hwnd,
		uintptr(HWND_TOPMOST),
		0, 0, 0, 0,
		uintptr(SWP_NOMOVE|SWP_NOSIZE), // Não mudar tamanho nem posição, só a ordem
	)
}
