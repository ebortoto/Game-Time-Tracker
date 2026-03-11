package ui

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	// Windows API constant for read/write access to shared memory
	fileMapAllAccess = 0xF001F
	// The specific name RTSS uses for its shared memory block
	rtssMemoryName = "RTSSSharedMemoryV2"
)

var (
	osdArrOffsetPtr *uint32
	osdFramePtr     *uint32
	osdEntryAddr    uintptr
)

func InitOverlay() {
	// Load the Windows kernel32 DLL to access memory mapping functions
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procOpenFileMappingW := kernel32.NewProc("OpenFileMappingW")
	procMapViewOfFile := kernel32.NewProc("MapViewOfFile")
	procUnmapViewOfFile := kernel32.NewProc("UnmapViewOfFile")
	procCloseHandle := kernel32.NewProc("CloseHandle")

	namePtr, _ := syscall.UTF16PtrFromString(rtssMemoryName)

	// 1. Open the RTSS Shared Memory block
	handle, _, err := procOpenFileMappingW.Call(
		uintptr(fileMapAllAccess),
		0,
		uintptr(unsafe.Pointer(namePtr)),
	)

	if handle == 0 {
		fmt.Printf("Could not open RTSS shared memory. Is RTSS running? Error: %v\n", err)
		return
	}
	defer procCloseHandle.Call(handle)

	// 2. Map the memory into our Go program's address space
	addr, _, err := procMapViewOfFile.Call(
		handle,
		uintptr(fileMapAllAccess),
		0, 0, 0,
	)

	if addr == 0 {
		fmt.Printf("Could not map view of file: %v\n", err)
		return
	}
	defer procUnmapViewOfFile.Call(addr)

	fmt.Println("Successfully connected to RTSS Shared Memory!")

	// 3. Navigate the RTSS C-Struct in memory
	// The OSD Array Offset is 24 bytes into the struct.
	// The OSD Frame Counter is 32 bytes into the struct.
	osdArrOffsetPtr = (*uint32)(unsafe.Pointer(addr + 24))
	osdFramePtr = (*uint32)(unsafe.Pointer(addr + 32))

	// Calculate the exact memory address where the text needs to go
	osdEntryAddr = addr + uintptr(*osdArrOffsetPtr)
}

func UpdateText(texto string) {
	// 4. Write our timer string to the memory address
	textBytes := append([]byte(texto), 0) // It must be a null-terminated C-string

	// Create a Go slice pointing directly to that block of shared memory and copy our text in
	dest := unsafe.Slice((*byte)(unsafe.Pointer(osdEntryAddr)), len(textBytes))
	copy(dest, textBytes)

	// Increment the frame counter. This tells the RTSS engine "Hey, the text changed, redraw it!"
	*osdFramePtr++

	fmt.Printf("Sent to RTSS: %s\n", texto)
}
