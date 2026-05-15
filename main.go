package main

import (
	"embed"
	"encoding/binary"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	crand "crypto/rand"
)

//go:embed videos/*
var videosFS embed.FS

var (
	width  = 100
	height = 50
)

// Estructura auxiliar para saber de dónde viene cada video
type VideoEntry struct {
	Name       string
	IsExternal bool // true = disco duro, false = embed
}

func main() {
	var mode string
	defaultMode := "kitty"
	if os.Getenv("TERM") != "xterm-kitty" && os.Getenv("KITTY_PID") == "" {
		defaultMode = "block"
	}
	flag.StringVar(&mode, "mode", defaultMode, "Modo: 'braille', 'block', 'kitty'")
	flagWidth := flag.Int("w", 0, "Ancho (caracteres). 0 = auto")
	flagHeight := flag.Int("h", 0, "Alto (lineas). 0 = auto")
	flagVideoNum := flag.Int("n", 0, "Número de video a reproducir (0 = aleatorio)")
	flagList := flag.Bool("list", false, "Listar videos disponibles")
	flag.Parse()

	// 1. Selección de Video
	videoData, err := selectVideo(*flagVideoNum, *flagList)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// --- PROTECCIÓN NUEVA ---
	// Verificar que ffmpeg puede entender este archivo antes de intentar reproducirlo
	if err := verifyVideo(videoData); err != nil {
		fmt.Printf("\n❌ Error Crítico: El archivo seleccionado no es un video válido.\n")
		fmt.Printf("Detalle: %v\n", err)
		return
	}
	// ------------------------


	// --- CLEANUP & SIGNALS ---
	setupSignalHandling()

	fmt.Print("\033[?25l") // Ocultar cursor
	// Defer cleanup para salida normal (aunque el loop es infinito)
	defer cleanup()

	// Buscar config.jsonc en varias ubicaciones
	configPath := findConfigPath()
	infoLines := getFastfetchData(configPath)

	// 2. Detección automática de tamaño, considerando el texto
	width, height = calculateDimensions(mode, *flagWidth, *flagHeight, videoData, infoLines)

	for {
		if err := playVideo(videoData, infoLines, mode); err != nil {
			time.Sleep(time.Second)
		}
	}
}

func findConfigPath() string {
	// Ubicaciones posibles para config.jsonc (en orden de prioridad)
	homeDir, _ := os.UserHomeDir()

	possiblePaths := []string{
		filepath.Join(homeDir, ".config", "vidfetch", "config.jsonc"),
		"/etc/vidfetch/config.jsonc",
		"./config.jsonc",
		"config.jsonc",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Si no se encuentra config.jsonc, retornar vacío (fallback en getFastfetchData)
	return ""
}

func calculateDimensions(mode string, flagWidth, flagHeight int, videoData []byte, infoLines []string) (int, int) {
	if flagWidth > 0 && flagHeight > 0 {
		return flagWidth, flagHeight
	}

	termW, termH, err := getTerminalSize()
	if err != nil || termW <= 0 || termH <= 0 {
		termW, termH = 100, 30
	}

	maxTextWidth := 0
	for _, line := range infoLines {
		cleanLen := len(removeAnsiCodes(line))
		if cleanLen > maxTextWidth {
			maxTextWidth = cleanLen
		}
	}

	gap := 4
	textReserve := maxTextWidth + gap + 2

	maxCols := termW - textReserve
	if maxCols < 10 {
		maxCols = 10
	}

	maxRows := int(float64(termH) * 0.90)
	if maxRows < 5 {
		maxRows = 5
	}

	videoW, videoH, err := getVideoSize(videoData)
	if err != nil {
		videoW, videoH = 1920, 1080
	}

	var finalCols, finalRows int
	aspectRatio := float64(videoW) / float64(videoH)

	switch mode {
	case "kitty":
		cellAspect := aspectRatio * (18.0 / 9.0)
		fitRows := maxRows
		fitCols := int(float64(fitRows) * cellAspect)
		if fitCols > maxCols {
			fitCols = maxCols
			fitRows = int(float64(fitCols) / cellAspect)
		}
		finalCols = fitCols
		finalRows = fitRows
	case "block":
		cellAspect := aspectRatio * (2.0 / 1.0)
		fitRows := maxRows
		fitCols := int(float64(fitRows) * cellAspect)
		if fitCols > maxCols {
			fitCols = maxCols
			fitRows = int(float64(fitCols) / cellAspect)
		}
		finalCols = fitCols
		finalRows = fitRows
	default:
		cellAspect := aspectRatio * (4.0 / 2.0)
		fitRows := maxRows
		fitCols := int(float64(fitRows) * cellAspect)
		if fitCols > maxCols {
			fitCols = maxCols
			fitRows = int(float64(fitCols) / cellAspect)
		}
		finalCols = fitCols
		finalRows = fitRows
	}

	if finalCols%2 != 0 {
		finalCols--
	}
	if finalCols < 4 {
		finalCols = 4
	}
	if finalRows < 2 {
		finalRows = 2
	}

	return finalCols, finalRows
}

func removeAnsiCodes(s string) string {
	// Remover secuencias ANSI básicas para contar caracteres visibles
	result := ""
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			inEscape = true
		} else if inEscape {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEscape = false
			}
		} else {
			result += string(s[i])
		}
	}
	return result
}

func setupSignalHandling() {
	// Capturar Ctrl+C para limpiar la terminal antes de salir
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()

	// Capturar Ctrl+Z para limpiar antes de suspender
	z := make(chan os.Signal, 1)
	signal.Notify(z, syscall.SIGTSTP)
	go func() {
		for {
			<-z
			cleanup()
			// Re-enviar la señal para que se suspenda realmente
			syscall.Kill(syscall.Getpid(), syscall.SIGSTOP)
		}
	}()
}

func cleanup() {
	fmt.Print("\033[?25h") // Mostrar cursor
	// Limpiar cualquier imagen residual de Kitty
	// 'a=d' delete all placement IDs
	fmt.Print("\x1b_Ga=d\x1b\\")
	fmt.Println() // Salto de línea final
}

func selectVideo(index int, listOnly bool) ([]byte, error) {
	allVideos, userVideoPath := loadVideos()

	if len(allVideos) == 0 {
		return nil, fmt.Errorf("no se encontraron videos (ni internos ni en %s)", userVideoPath)
	}

	// --- MODO LISTAR (-list) ---
	if listOnly {
		printVideoList(allVideos, userVideoPath)
		os.Exit(0)
	}

	// --- SELECCIÓN ---
	var selected VideoEntry

	if index == 0 {
		seedRandom()
		selected = allVideos[rand.Intn(len(allVideos))]
	} else {
		if index < 1 || index > len(allVideos) {
			return nil, fmt.Errorf("video número %d no existe (hay %d disponibles)", index, len(allVideos))
		}
		selected = allVideos[index-1]
	}

	return readVideoContent(selected, userVideoPath)
}

func loadVideos() ([]VideoEntry, string) {
	var allVideos []VideoEntry

	// 1. Cargar videos EMBEBIDOS (Internos)
	embeddedFiles, err := videosFS.ReadDir("videos")
	if err == nil {
		for _, f := range embeddedFiles {
			allVideos = append(allVideos, VideoEntry{Name: f.Name(), IsExternal: false})
		}
	}

	// 2. Cargar videos EXTERNOS (Usuario)
	homeDir, _ := os.UserHomeDir()
	userVideoPath := filepath.Join(homeDir, "Videos", "VidFetch") // Fallback por defecto
	
	cmd := exec.Command("xdg-user-dir", "VIDEOS")
	out, err := cmd.Output()
	if err == nil {
		xdgVideos := strings.TrimSpace(string(out))
		if xdgVideos != "" {
			userVideoPath = filepath.Join(xdgVideos, "VidFetch")
		}
	}

	os.MkdirAll(userVideoPath, 0755)

	userFiles, err := os.ReadDir(userVideoPath)
	if err == nil {
		for _, f := range userFiles {
			if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".mp4") {
				allVideos = append(allVideos, VideoEntry{Name: f.Name(), IsExternal: true})
			}
		}
	}

	// Filtrar duplicados: el video externo sobreescribe al interno si tienen el mismo nombre
	uniqueVideos := make(map[string]VideoEntry)
	for _, v := range allVideos {
		if existing, ok := uniqueVideos[v.Name]; ok {
			if v.IsExternal {
				uniqueVideos[v.Name] = v // externo tiene prioridad
			} else {
				uniqueVideos[v.Name] = existing
			}
		} else {
			uniqueVideos[v.Name] = v
		}
	}

	var finalVideos []VideoEntry
	for _, v := range uniqueVideos {
		finalVideos = append(finalVideos, v)
	}

	sort.Slice(finalVideos, func(i, j int) bool {
		return finalVideos[i].Name < finalVideos[j].Name
	})

	return finalVideos, userVideoPath
}

func printVideoList(allVideos []VideoEntry, userVideoPath string) {
	fmt.Printf("Videos disponibles (%d total):\n", len(allVideos))
	fmt.Printf("Ruta externa: %s\n", userVideoPath)
	fmt.Println("---------------------------------------------------")
	for i, v := range allVideos {
		origin := "[Interno]"
		if v.IsExternal {
			origin = "[Usuario]"
		}
		fmt.Printf("  [%d] %s %s\n", i+1, origin, v.Name)
	}
}

func seedRandom() {
	// Use crypto/rand for better seeding
	var b [8]byte
	_, err := crand.Read(b[:])
	if err == nil {
		seed := int64(binary.LittleEndian.Uint64(b[:]))
		rand.Seed(seed)
	} else {
		rand.Seed(time.Now().UnixNano())
	}
}

func readVideoContent(selected VideoEntry, userVideoPath string) ([]byte, error) {
	if selected.IsExternal {
		fullPath := filepath.Join(userVideoPath, selected.Name)
		return os.ReadFile(fullPath)
	} else {
		return videosFS.ReadFile("videos/" + selected.Name)
	}
}
