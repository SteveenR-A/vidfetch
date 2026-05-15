# 🎬 VidFetch

**VidFetch** es una herramienta de visualización de información del sistema que reproduce animaciones de video directamente en tu terminal, mostrándolas junto con tus datos de `fastfetch`. A diferencia de los fetch tradicionales estáticos, VidFetch le da un toque dinámico a tu terminal soportando gráficos en alta calidad (mediante el protocolo Kitty) y alternativas con caracteres Unicode.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.16-00ADD8.svg)

---

## ✨ Características

- 🎥 **Reproducción de videos en la terminal** con tres modos de renderizado.
- 🎨 **Modo Kitty Graphics:** Reproduce video real en alta calidad para terminales compatibles.
- 🔲 **Modo Block:** Animaciones usando caracteres Unicode de medio bloque de color.
- ⠿ **Modo Braille:** Alternativa ligera para terminales más básicos.
- 📦 **Videos embebidos** directamente en el ejecutable (¡funciona de inmediato!).
- 🎲 **Selección aleatoria** o específica de videos.
- 📁 **Videos personalizados** desde tu carpeta de `Videos/VidFetch/` del sistema.
- 🔧 **Dimensiones automáticas** basadas en el tamaño del terminal
- ⚡ **Optimizado** para bajo uso de CPU y rendering fluido
- 🎯 **Adaptación dinámica** al ancho del texto de información

---

## 📦 Dependencias

### Obligatorias

Estas dependencias son **necesarias** para que VidFetch funcione correctamente:

| Dependencia | Propósito | Instalación (Arch/Manjaro) |
|-------------|-----------|----------------------------|
| **Go** | Compilación del proyecto | `sudo pacman -S go` |
| **FFmpeg** | Procesamiento de video | `sudo pacman -S ffmpeg` |
| **FFprobe** | Lectura de metadatos de video | Incluido con `ffmpeg` |
| **fastfetch** | Generación de información del sistema | `sudo pacman -S fastfetch` |

### Opcionales (según modo de renderizado)

| Dependencia | Requerida para | Instalación |
|-------------|----------------|-------------|
| **Kitty Terminal** | Modo `kitty` (HD) | `sudo pacman -S kitty` |

---

## 🚀 Instalación

### 1. Clonar el repositorio

```bash
git clone https://github.com/tu-usuario/vidfetch.git
cd vidfetch
```

### 2. Ejecutar el Instalador (Recomendado)

VidFetch incluye un instalador interactivo para facilitar el proceso en Arch Linux:

```bash
./install.sh
```

El instalador te permitirá:
- Instalar automáticamente las dependencias necesarias (`go`, `ffmpeg`, `fastfetch`, `kitty`).
- Compilar y mover el ejecutable a `~/.local/bin/vidfetch`.
- Crear un alias `vidfetch` por comodidad.
- Configurar tu directorio local de videos en tu carpeta nativa del sistema (ej. `~/Videos/VidFetch`).
- Desinstalar el programa fácilmente.

### 3. Ejecución Manual

Si prefieres hacerlo de forma manual:

```bash
go build .
cp vidfetch ~/.local/bin/vidfetch
mkdir -p ~/.config/vidfetch
cp config.jsonc ~/.config/vidfetch/
```
---

## 🎮 Uso

### Sintaxis básica

```bash
vidfetch [opciones]
```

### Opciones disponibles

| Flag | Valores | Predeterminado | Descripción |
|------|---------|----------------|-------------|
| `-mode` | `kitty`, `block`, `braille` | `kitty` | Modo de renderizado |
| `-w` | Entero (0 = auto) | `0` | Ancho en caracteres |
| `-h` | Entero (0 = auto) | `0` | Alto en líneas |
| `-n` | Entero (0 = aleatorio) | `0` | Número de video específico |
| `-list` | - | - | Listar todos los videos disponibles |

### Ejemplos

```bash
# Ejecutar con configuración predeterminada (modo kitty, video aleatorio)
vidfetch

# Listar todos los videos disponibles
vidfetch -list

# Reproducir el video número 3
vidfetch -n 3

# Usar modo braille (compatible con cualquier terminal)
vidfetch -mode braille

# Usar modo block (bloques Unicode de color)
vidfetch -mode block

# Tamaño personalizado: 80 caracteres de ancho, 30 líneas de alto
vidfetch -w 80 -h 30

# Combinación: modo braille, video #2, tamaño específico
vidfetch -mode braille -n 2 -w 100 -h 40
```

---

## 📂 Estructura del Proyecto

```
vidfetch/
├── main.go              # Punto de entrada, manejo de flags y señales
├── video.go             # Lógica de reproducción y streaming de FFmpeg
├── render.go            # Funciones de renderizado (Kitty, Block, Braille)
├── utils.go             # Utilidades (fastfetch, terminal size, ffprobe)
├── config.jsonc         # Configuración de fastfetch
├── videos/              # Videos embebidos en el ejecutable
│   ├── video1.mp4
│   ├── video2.mp4
│   └── video3.mp4
└── README.md            # Este archivo
```

---

## ⚙️ Configuración

### Videos personalizados

Puedes agregar tus propios videos (formato `.mp4`) en la carpeta generada en tu directorio nativo de Videos:

```bash
~/Videos/VidFetch/
```
*(Nota: la ruta exacta se adapta al idioma de tu sistema gracias a `xdg-user-dir VIDEOS`)*

Los videos personalizados aparecerán en la lista junto con los embebidos:

```bash
vidfetch -list
```
Si un video personalizado tiene el mismo nombre que uno interno, el programa le dará prioridad al tuyo (evitando duplicados).

### Configuración de fastfetch

El archivo `config.jsonc` en el directorio del proyecto controla qué información se muestra. Puedes editarlo para personalizar:

- Logo del sistema
- Información mostrada (OS, kernel, terminal, etc.)
- Colores y formato

**Ubicación principal:** `~/.config/vidfetch/config.jsonc`

---

## 🎨 Modos de Renderizado

### 🖼️ Modo Kitty (`-mode kitty`)

- **Calidad:** Alta definición (HD)
- **Requisito:** Terminal Kitty
- **Ventaja:** Mejor calidad visual con imágenes reales
- **Desventaja:** Solo funciona en Kitty terminal

### 🔲 Modo Block (`-mode block`)

- **Calidad:** Media
- **Requisito:** Terminal con soporte Unicode
- **Ventaja:** Buen balance entre calidad y compatibilidad
- **Desventaja:** Menor detalle que Kitty

### ⠿ Modo Braille (`-mode braille`)

- **Calidad:** Básica
- **Requisito:** Cualquier terminal
- **Ventaja:** Máxima compatibilidad
- **Desventaja:** Menor calidad visual

---

## 🔧 Cómo Funciona

### Pipeline de procesamiento

```
Video (MP4) 
    ↓
FFmpeg (procesamiento y conversión)
    ↓
Frames en formato raw (RGBA/RGB24)
    ↓
Renderizado según modo seleccionado:
    ├── Kitty: Base64 → Protocolo Kitty Graphics
    ├── Block: RGB → Unicode half blocks (▀)
    └── Braille: Luminancia → Braille dots (⠿)
    ↓
Output en terminal con información de fastfetch al lado
```

### Características técnicas

- **Framerate:** ~30 FPS
- **Escalado dinámico:** Se ajusta al tamaño del terminal
- **Filtros FFmpeg:** Lanczos para mejor calidad
- **Gestión de señales:** Limpieza al salir (Ctrl+C, Ctrl+Z)
- **Cálculo de espacio:** Reserva automática de espacio para texto

---

## 🐛 Solución de Problemas

### El video cubre el texto de información

Si el texto se sobrepone con la imagen:

1. Verifica que el `gap` en `render.go` sea suficiente (actualmente: 12)
2. Reduce el ancho del video manualmente: `vidfetch -w 80`

### FFmpeg no encontrado

```bash
sudo pacman -S ffmpeg
```

### Fastfetch no encontrado

```bash
sudo pacman -S fastfetch
```

### El modo Kitty no funciona

Asegúrate de estar usando Kitty terminal:

```bash
echo $TERM
# Debería mostrar: xterm-kitty
```

Si no estás en Kitty, usa otro modo:

```bash
vidfetch -mode block
```

### Video corrupto o error al cargar

Verifica que el archivo sea un `.mp4` válido:

```bash
ffprobe tu_video.mp4
```

---

## 🛠️ Desarrollo

### Compilar

```bash
go build .
```

### Instalar localmente (para desarrollo)

```bash
go install .
```

### Agregar un nuevo video embebido

1. Coloca el archivo `.mp4` en `videos/`
2. Recompila el proyecto
3. El video se embebará automáticamente gracias a `//go:embed`

---

## 📝 Notas Técnicas

### Variables globales

- `width`, `height`: Dimensiones de renderizado (se calculan automáticamente)

### Funciones principales

| Función | Archivo | Propósito |
|---------|---------|-----------|
| `main()` | `main.go` | Punto de entrada |
| `playVideo()` | `video.go` | Loop de reproducción |
| `startFFmpeg()` | `video.go` | Inicializar FFmpeg |
| `processStream()` | `video.go` | Procesar frames |
| `renderFrameKitty()` | `render.go` | Renderizar en Kitty |
| `renderFrameBlock()` | `render.go` | Renderizar con blocks |
| `renderFrameBraille()` | `render.go` | Renderizar con Braille |
| `getFastfetchData()` | `utils.go` | Obtener info del sistema |
| `getTerminalSize()` | `utils.go` | Detectar tamaño del terminal |
| `calculateDimensions()` | `main.go` | Calcular dimensiones óptimas |
| `recalculateDimensions()` | `main.go` | Ajustar según espacio disponible |

---

## 📜 Licencia

Este proyecto está bajo la licencia MIT. Siéntete libre de usar, modificar y distribuir.

---

## 🙏 Créditos

- **FFmpeg** - Procesamiento de video
- **Fastfetch** - Información del sistema
- **Kitty Terminal** - Protocolo de gráficos
- **Go** - Lenguaje de programación

---



