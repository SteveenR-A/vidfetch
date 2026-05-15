package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// getFastfetchData ejecuta fastfetch y devuelve las líneas
func getFastfetchData(configPath string) []string {
	var cmd *exec.Cmd

	// Si configPath está vacío o no existe, usar fastfetch con config por defecto
	if configPath == "" {
		cmd = exec.Command("fastfetch", "--pipe", "false")
	} else {
		cmd = exec.Command("fastfetch", "--config", configPath, "--pipe", "false")
	}

	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return []string{"Fastfetch Error"}
	}

	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if strings.Contains(line, "Terminal") && (strings.Contains(line, "go") || strings.Contains(line, "electron")) {
			lines[i] = strings.Replace(line, "go", "kitty", 1)
			lines[i] = strings.Replace(lines[i], "electron", "kitty", 1)
		}
	}
	return lines
}

// getTerminalSize obtiene ancho y alto del terminal
func getTerminalSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Fields(string(out))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("err")
	}
	r, _ := strconv.Atoi(parts[0])
	c, _ := strconv.Atoi(parts[1])
	return c, r, nil
}

// getVideoSize usa ffprobe para obtener dimensiones del video
func getVideoSize(videoData []byte) (int, int, error) {
	tmpFile, err := os.CreateTemp("", "vid-*.mp4")
	if err != nil {
		return 0, 0, err
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(videoData)
	tmpFile.Close()

	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=width,height", "-of", "csv=s=x:p=0", tmpFile.Name())
	out, _ := cmd.Output()
	parts := strings.Split(strings.TrimSpace(string(out)), "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("err")
	}
	w, _ := strconv.Atoi(parts[0])
	h, _ := strconv.Atoi(parts[1])
	return w, h, nil
}

// verifyVideo verifica si el archivo es un video válido
func verifyVideo(data []byte) error {
	tmp, _ := os.CreateTemp("", "check_*.mp4")
	defer os.Remove(tmp.Name())
	tmp.Write(data)
	tmp.Close()

	cmd := exec.Command("ffprobe", "-v", "error", tmp.Name())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("formato corrupto o desconocido")
	}
	return nil
}
