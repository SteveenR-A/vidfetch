package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// playVideo ejecuta el bucle de reproducción para un video dado
func playVideo(videoData []byte, infoLines []string, mode string) error {
	cmd, stdout, err := startFFmpeg(videoData, mode)
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	processStream(stdout, infoLines, mode)

	return cmd.Wait()
}

func startFFmpeg(videoData []byte, mode string) (*exec.Cmd, io.ReadCloser, error) {
	var pixelWidth, pixelHeight int
	var pixFmt string
	var filter string

	// Width y Height son variables globales definidas en main.go
	if mode == "kitty" {
		pixelWidth = width * 9
		pixelHeight = height * 18

		if pixelWidth%2 != 0 {
			pixelWidth--
		}
		if pixelHeight%2 != 0 {
			pixelHeight--
		}

		pixFmt = "rgba"
		// Align to TOP: y=0 instead of (oh-ih)/2
		filter = fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease:flags=lanczos,pad=%d:%d:(ow-iw)/2:0:color=0x00000000", pixelWidth, pixelHeight, pixelWidth, pixelHeight)

	} else if mode == "block" {
		pixelWidth = width
		pixelHeight = height * 2
		pixFmt = "rgb24"
		filter = fmt.Sprintf("eq=contrast=1.1:saturation=1.2,scale=%d:%d:force_original_aspect_ratio=decrease:flags=lanczos,pad=%d:%d:(ow-iw)/2:(oh-ih)/2", pixelWidth, pixelHeight, pixelWidth, pixelHeight)
	} else {
		// Braille
		pixelWidth = width * 2
		pixelHeight = height * 4
		pixFmt = "rgb24"
		filter = fmt.Sprintf("eq=contrast=1.1:saturation=1.2,scale=%d:%d:force_original_aspect_ratio=decrease:flags=lanczos,unsharp=3:3:1.0,pad=%d:%d:(ow-iw)/2:(oh-ih)/2", pixelWidth, pixelHeight, pixelWidth, pixelHeight)
	}

	cmd := exec.Command("ffmpeg",
		"-re",
		"-i", "pipe:0",
		"-vf", filter,
		"-f", "image2pipe",
		"-vcodec", "rawvideo",
		"-pix_fmt", pixFmt,
		"-",
		"-loglevel", "quiet",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	go func() {
		defer stdin.Close()
		reader := bytes.NewReader(videoData)
		io.Copy(stdin, reader)
	}()

	stdout, err := cmd.StdoutPipe()
	return cmd, stdout, err
}

func processStream(stdout io.ReadCloser, infoLines []string, mode string) {
	pixelWidth, pixelHeight, bytesPerPixel := getVideoDimensions(mode)

	frameSize := pixelWidth * pixelHeight * bytesPerPixel

	fmt.Print("\033[2J") // Clear screen

	reader := bufio.NewReader(stdout)
	outputBuf := make([]byte, 0, width*height*100)
	frameBuf := make([]byte, frameSize)

	brailleMap := [4][2]int{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}}

	// Pre-select render function
	var renderOp func(in, out []byte) []byte

	switch mode {
	case "kitty":
		renderOp = func(in, out []byte) []byte {
			return renderFrameKitty(in, pixelWidth, pixelHeight, infoLines, out)
		}
	case "block":
		renderOp = func(in, out []byte) []byte {
			return renderFrameBlock(in, width, height, infoLines, out)
		}
	default:
		renderOp = func(in, out []byte) []byte {
			return renderFrameBraille(in, width, height, pixelWidth, brailleMap, infoLines, out)
		}
	}

	lastTime := time.Now()

	for {
		_, err := io.ReadFull(reader, frameBuf)
		if err != nil {
			break
		}

		outputBuf = outputBuf[:0]
		outputBuf = renderOp(frameBuf, outputBuf)

		os.Stdout.Write(outputBuf)

		elapsed := time.Since(lastTime)
		if elapsed < 33*time.Millisecond {
			time.Sleep(33*time.Millisecond - elapsed)
		}
		lastTime = time.Now()
	}
}

func getVideoDimensions(mode string) (int, int, int) {
	switch mode {
	case "kitty":
		pw := width * 9
		ph := height * 18
		if pw%2 != 0 {
			pw--
		}
		if ph%2 != 0 {
			ph--
		}
		return pw, ph, 4 // RGBA
	case "block":
		return width, height * 2, 3
	default:
		return width * 2, height * 4, 3
	}
}
