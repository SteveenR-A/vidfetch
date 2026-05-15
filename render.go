package main

import (
	"encoding/base64"
	"fmt"
)

// --- RENDERIZADO KITTY (HD) ---

func renderFrameKitty(frameBuf []byte, pxW, pxH int, infoLines []string, buf []byte) []byte {
	// Limpiar pantalla al inicio de cada frame para evitar superposición
	buf = append(buf, []byte("\033[H")...)


	// Primero renderizar el texto (para que la imagen no lo cubra)
	imageCols := pxW / 9
	gap := 4

	for i, line := range infoLines {
		move := fmt.Sprintf("\033[%d;%dH", i+1, imageCols+gap)
		buf = append(buf, []byte(move)...)
		buf = append(buf, []byte(line)...)
		// Borrar hasta el final de la línea para evitar artefactos del frame anterior
		buf = append(buf, []byte("\033[K")...)
	}

	// Ahora enviar la imagen con posicionamiento absoluto
	// Volver a la posición inicial
	buf = append(buf, []byte("\033[H")...)

	encoded := base64.StdEncoding.EncodeToString(frameBuf)
	chunkSize := 4096
	totalLen := len(encoded)

	for i := 0; i < totalLen; i += chunkSize {
		end := i + chunkSize
		if end > totalLen {
			end = totalLen
		}
		chunk := encoded[i:end]

		m := "1"
		if end == totalLen {
			m = "0"
		}

		if i == 0 {
			// Usar c y r en lugar de X,Y para que Kitty escale la imagen a las celdas correctas
			// según la resolución real del monitor. z=-1 para que esté DETRÁS del texto
			cols := pxW / 9
			rows := pxH / 18
			header := fmt.Sprintf("a=T,f=32,s=%d,v=%d,c=%d,r=%d,z=-1,q=2,m=%s;", pxW, pxH, cols, rows, m)
			buf = append(buf, []byte("\x1b_G"+header+chunk+"\x1b\\")...)
		} else {
			buf = append(buf, []byte(fmt.Sprintf("\x1b_Gm=%s;%s\x1b\\", m, chunk))...)
		}
	}

	// Mover cursor debajo de la imagen
	linesHeight := pxH / 18
	if linesHeight < 1 {
		linesHeight = 1
	}
	buf = append(buf, []byte(fmt.Sprintf("\033[%d;1H", linesHeight+1))...)

	return buf
}

// --- RENDERIZADO BLOCK ---

func renderFrameBlock(frameBuf []byte, w, h int, infoLines []string, buf []byte) []byte {
	buf = append(buf, []byte("\033[H")...)
	pixelWidth := w

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idxTop := ((y*2)*pixelWidth + x) * 3
			idxBot := ((y*2+1)*pixelWidth + x) * 3

			r1, g1, b1 := frameBuf[idxTop], frameBuf[idxTop+1], frameBuf[idxTop+2]
			r2, g2, b2 := frameBuf[idxBot], frameBuf[idxBot+1], frameBuf[idxBot+2]

			s := fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm\u2580", r1, g1, b1, r2, g2, b2)
			buf = append(buf, []byte(s)...)
		}

		buf = append(buf, []byte("\033[0m   ")...)
		if y < len(infoLines) {
			buf = append(buf, []byte(infoLines[y])...)
		}
		buf = append(buf, '\n')
	}
	buf = append(buf, []byte("\033[0m")...)
	return buf
}

// --- RENDERIZADO BRAILLE ---

func renderFrameBraille(frameBuf []byte, w, h, pxW int, bMap [4][2]int, infoLines []string, buf []byte) []byte {
	// Logic based on standard braille encoding for 2x4 cell
	buf = append(buf, []byte("\033[H")...)

	// Braille offsets:
	// (0,0) (1,0) -> 0x01, 0x08
	// (0,1) (1,1) -> 0x02, 0x10
	// (0,2) (1,2) -> 0x04, 0x20
	// (0,3) (1,3) -> 0x40, 0x80

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var charCode int = 0x2800 // Braille base

			// Color averaging for the cell
			var rSum, gSum, bSum int
			count := 0

			for by := 0; by < 4; by++ {
				for bx := 0; bx < 2; bx++ {
					// Pixel coordinates
					py := y*4 + by
					px := x*2 + bx

					// Index in frameBuf
					idx := (py*pxW + px) * 3
					r := int(frameBuf[idx])
					g := int(frameBuf[idx+1])
					b := int(frameBuf[idx+2])

					// If brightness > threshold (simple check)
					// Using luminance formula approximation: 0.299R + 0.587G + 0.114B
					lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
					if lum > 60 { // Threshold
						charCode |= bMap[by][bx]
						rSum += r
						gSum += g
						bSum += b
						count++
					}
				}
			}

			if count > 0 {
				rSum /= count
				gSum /= count
				bSum /= count
				buf = append(buf, []byte(fmt.Sprintf("\033[38;2;%d;%d;%dm%c", rSum, gSum, bSum, rune(charCode)))...)
			} else {
				buf = append(buf, ' ')
			}
		}

		buf = append(buf, []byte("\033[0m   ")...)
		if y < len(infoLines) {
			buf = append(buf, []byte(infoLines[y])...)
		}
		buf = append(buf, '\n')
	}
	buf = append(buf, []byte("\033[0m")...)
	return buf
}
