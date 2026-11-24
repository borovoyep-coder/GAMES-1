//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"github.com/jchv/go-webview2/webviewloader"
	"golang.org/x/sys/windows"
)

const windowTitle = "Каменный рубеж — PvP Line Tower Defense"

var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	messageBox       = user32.NewProc("MessageBoxW")
	findWindow       = user32.NewProc("FindWindowW")
	showWindow       = user32.NewProc("ShowWindow")
	foregroundWindow = user32.NewProc("SetForegroundWindow")
	getSystemMetrics = user32.NewProc("GetSystemMetrics")
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")
	createMutex      = kernel32.NewProc("CreateMutexW")
	shell32          = windows.NewLazySystemDLL("shell32.dll")
)

func wide(value string) *uint16 {
	return windows.StringToUTF16Ptr(value)
}

func notify(text string, flags uintptr) uintptr {
	result, _, _ := messageBox.Call(0, uintptr(unsafe.Pointer(wide(text))), uintptr(unsafe.Pointer(wide("Каменный рубеж"))), flags)
	return result
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := run(); err != nil {
		notify("Не удалось запустить игру.\n\n"+err.Error(), 0x10)
	}
}

func run() error {
	mutex, _, mutexErr := createMutex.Call(0, 0, uintptr(unsafe.Pointer(wide(`Local\StoneLine.Desktop`))))
	if mutex == 0 {
		return fmt.Errorf("проверка запущенной копии: %v", mutexErr)
	}
	defer windows.CloseHandle(windows.Handle(mutex))
	if mutexErr == windows.ERROR_ALREADY_EXISTS {
		hwnd, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(wide(windowTitle))))
		if hwnd != 0 {
			showWindow.Call(hwnd, 9)
			foregroundWindow.Call(hwnd)
		}
		return nil
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("папка сохранений недоступна: %w", err)
	}
	dataDir := filepath.Join(cacheDir, "StoneLine")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("не удалось создать папку сохранений: %w", err)
	}
	logFile, err := os.OpenFile(filepath.Join(dataDir, "launcher.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	gameURL, err := prepareGame(dataDir)
	if err != nil {
		return fmt.Errorf("подготовка игры: %w", err)
	}
	// Load Microsoft's bundled loader from an explicit path. The app does not
	// download executable code or need separate DLLs next to the downloaded EXE.
	loaderPath := filepath.Join(dataDir, "WebView2Loader.dll")
	if err := writeIfChanged(loaderPath, webviewloader.WebView2Loader); err != nil {
		return fmt.Errorf("подготовка WebView2: %w", err)
	}
	loader, err := windows.LoadLibraryEx(loaderPath, 0, windows.LOAD_LIBRARY_SEARCH_DLL_LOAD_DIR|windows.LOAD_LIBRARY_SEARCH_SYSTEM32)
	if err != nil {
		return fmt.Errorf("загрузка WebView2Loader: %w", err)
	}
	defer windows.FreeLibrary(loader)
	version, err := webviewloader.GetInstalledVersion()
	if err != nil {
		return fmt.Errorf("проверка Microsoft Edge WebView2: %w", err)
	}
	if version == "" {
		answer := notify("Для окна игры требуется Microsoft Edge WebView2 Runtime.\n\nОткрыть официальную страницу Microsoft для его установки? После установки запустите игру снова.", 0x24)
		if answer == 6 {
			shell32.NewProc("ShellExecuteW").Call(0, uintptr(unsafe.Pointer(wide("open"))), uintptr(unsafe.Pointer(wide("https://developer.microsoft.com/microsoft-edge/webview2/"))), 0, 0, 1)
		}
		return nil
	}
	shell32.NewProc("SetCurrentProcessExplicitAppUserModelID").Call(uintptr(unsafe.Pointer(wide("StoneLine.Desktop"))))
	if result, _, _ := windows.NewLazySystemDLL("ole32.dll").NewProc("CoInitializeEx").Call(0, 2); result == 0 || result == 1 {
		defer windows.NewLazySystemDLL("ole32.dll").NewProc("CoUninitialize").Call()
	}
	screenW, _, _ := getSystemMetrics.Call(0)
	screenH, _, _ := getSystemMetrics.Call(1)
	width, height := uint(1440), uint(900)
	if screenW > 100 && width > uint(screenW)-80 {
		width = uint(screenW) - 80
	}
	if screenH > 100 && height > uint(screenH)-80 {
		height = uint(screenH) - 80
	}
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:         false,
		AutoFocus:     true,
		DataPath:      filepath.Join(dataDir, "WebView2"),
		WindowOptions: webview2.WindowOptions{Title: windowTitle, Width: width, Height: height, IconId: 2, Center: true},
	})
	if w == nil {
		return fmt.Errorf("не удалось создать окно WebView2; подробности: %s", filepath.Join(dataDir, "launcher.log"))
	}
	defer w.Destroy()
	minW, minH := 900, 560
	if int(width) < minW {
		minW = int(width)
	}
	if int(height) < minH {
		minH = int(height)
	}
	w.SetSize(minW, minH, webview2.HintMin)
	w.Navigate(gameURL)
	w.Run()
	return nil
}
