#!/bin/bash

# Colores para salida en terminal
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Función para imprimir mensajes de éxito
print_success() {
    echo -e "${GREEN}[✔] $1${NC}"
}

# Función para imprimir mensajes de error
print_error() {
    echo -e "${RED}[✖] $1${NC}"
}

# Función para imprimir información
print_info() {
    echo -e "${YELLOW}[i] $1${NC}"
}

# Asegurar que el script se ejecute desde su propio directorio
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
cd "$SCRIPT_DIR" || exit 1

# Mostrar menú
echo "======================================"
echo -e "${GREEN}   Instalador de VidFetch${NC}"
echo "======================================"
echo "Por favor, selecciona una opción:"
echo "1) Instalar VidFetch"
echo "2) Desinstalar VidFetch"
echo "3) Salir"
echo "======================================"
read -p "Opción [1-3]: " opcion

case $opcion in
    1)
        print_info "Iniciando la instalación de VidFetch..."
        
        # Comprobar si go está instalado
        if ! command -v go &> /dev/null; then
            print_error "Go no está instalado. Por favor, instálalo con: sudo pacman -S go"
            exit 1
        fi

        # Comprobar si ffmpeg, fastfetch, kitty y xdg-user-dirs están instalados
        for pkg in ffmpeg fastfetch kitty xdg-user-dirs; do
            if ! command -v $pkg &> /dev/null; then
                print_info "$pkg no está instalado. Instalándolo..."
                sudo pacman -S --noconfirm $pkg
            fi
        done

        # Compilar el proyecto
        print_info "Compilando VidFetch..."
        if go build -o vidfetch .; then
            print_success "Compilación exitosa."
        else
            print_error "Error al compilar VidFetch."
            exit 1
        fi

        # Instalar el binario
        print_info "Instalando ejecutable en ~/.local/bin/..."
        
        # Eliminar versión antigua de /usr/bin si existía para evitar conflictos
        if [ -f /usr/bin/vidfetch ]; then
            print_info "Limpiando instalación antigua de /usr/bin/ (requiere permisos sudo)..."
            sudo rm -f /usr/bin/vidfetch 
        fi
        
        # Eliminar posible instalación local antigua
        rm -f ~/.local/bin/vidfetch 
        
        if cp vidfetch ~/.local/bin/vidfetch; then
            chmod +x ~/.local/bin/vidfetch
            print_success "Ejecutable instalado correctamente en ~/.local/bin/vidfetch."
        else
            print_error "Error al instalar el ejecutable."
            exit 1
        fi

        # Instalar configuración
        print_info "Configurando archivos de usuario..."
        mkdir -p ~/.config/vidfetch
        if [ -f config.jsonc ]; then
            cp config.jsonc ~/.config/vidfetch/
            print_success "Configuración copiada a ~/.config/vidfetch/config.jsonc"
        fi

        # Limpiar antigua carpeta de videos en .config si existe
        #rm -rf ~/.config/vidfetch/videos
        
        # Crear carpeta única VidFetch en la carpeta de videos del usuario (multi-idioma)
        VIDEOS_DIR=$(xdg-user-dir VIDEOS)
        if [ -n "$VIDEOS_DIR" ]; then
            if [ ! -d "$VIDEOS_DIR/VidFetch" ] || ! ls "$VIDEOS_DIR/VidFetch/"*.mp4 1> /dev/null 2>&1; then
                mkdir -p "$VIDEOS_DIR/VidFetch"
                print_success "Carpeta de videos preparada en $VIDEOS_DIR/VidFetch"
                
                # Mover/copiar videos predeterminados al sistema del usuario si existen
                if ls videos/*.mp4 1> /dev/null 2>&1; then
                    cp -n videos/*.mp4 "$VIDEOS_DIR/VidFetch/" 2>/dev/null || true
                    print_success "Videos predeterminados copiados a $VIDEOS_DIR/VidFetch"
                fi
            else
                print_info "La carpeta $VIDEOS_DIR/VidFetch ya contiene videos. Se conservarán los existentes."
            fi
        fi

        print_success "¡Instalación completada con éxito!"
        print_info "Puedes ejecutar el programa escribiendo 'vidfetch' en tu terminal."
        ;;
    2)
        print_info "Iniciando la desinstalación de VidFetch..."

        # Encontrar y eliminar el binario
        vidfetch_path=$(command -v vidfetch)
        
        # También verificamos ubicaciones comunes directamente por si no está en el PATH actual
        if [ -z "$vidfetch_path" ]; then
            if [ -f ~/.local/bin/vidfetch ]; then
                vidfetch_path=~/.local/bin/vidfetch
            elif [ -f /usr/bin/vidfetch ]; then
                vidfetch_path=/usr/bin/vidfetch
            fi
        fi

        if [ -n "$vidfetch_path" ]; then
            print_info "Ejecutable encontrado en $vidfetch_path"
            if [ -w "$(dirname "$vidfetch_path")" ]; then
                if rm "$vidfetch_path"; then
                    print_success "Ejecutable eliminado correctamente."
                else
                    print_error "Error al eliminar $vidfetch_path"
                fi
            else
                print_info "Se requieren permisos sudo para eliminar $vidfetch_path"
                if sudo rm "$vidfetch_path"; then
                    print_success "Ejecutable eliminado correctamente."
                else
                    print_error "Error al eliminar $vidfetch_path"
                fi
            fi
        else
            print_info "El ejecutable no se encontró en el sistema."
        fi

        # Preguntar sobre la configuración
        read -p "¿Deseas eliminar también los archivos de configuración y videos descargados en ~/.config/vidfetch/? [s/N]: " del_config
        if [[ "$del_config" =~ ^[sS]$ ]]; then
            rm -rf ~/.config/vidfetch
            
            VIDEOS_DIR=$(xdg-user-dir VIDEOS)
            if [ -n "$VIDEOS_DIR" ] && [ -d "$VIDEOS_DIR/VidFetch" ]; then
                rm -rf "$VIDEOS_DIR/VidFetch"
                print_success "Carpeta $VIDEOS_DIR/VidFetch eliminada."
            fi
            
            print_success "Archivos de configuración eliminados."
        else
            print_info "Archivos de configuración conservados."
        fi

        # Preguntar sobre las dependencias
        echo ""
        print_info "⚠️  PRECAUCIÓN: Las siguientes dependencias podrían estar siendo utilizadas por otras aplicaciones en tu sistema:"
        echo "   - ffmpeg"
        echo "   - fastfetch"
        echo "   - kitty"
        echo "   - xdg-user-dirs"
        read -p "¿Deseas desinstalar estas dependencias de todas formas? [s/N]: " del_deps
        if [[ "$del_deps" =~ ^[sS]$ ]]; then
            print_info "Desinstalando dependencias..."
            # Usamos pacman -R para removerlas de forma segura (si otro paquete depende rígidamente de ellas, pacman lo advertirá)
            sudo pacman -R --noconfirm ffmpeg fastfetch kitty xdg-user-dirs || print_error "Algunas dependencias no se pudieron desinstalar (posiblemente estén en uso)."
            print_success "Intento de desinstalación de dependencias finalizado."
        else
            print_info "Dependencias conservadas."
        fi

        print_success "¡Desinstalación completada!"
        ;;
    3)
        print_info "Saliendo..."
        exit 0
        ;;
    *)
        print_error "Opción no válida."
        exit 1
        ;;
esac
