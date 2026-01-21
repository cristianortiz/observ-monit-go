# =============================================================================
# STAGE 1: Builder - Compilación del binario Go optimizado
# =============================================================================
FROM golang:1.23-alpine AS builder

# Instalar dependencias de compilación (git para go modules, ca-certificates para HTTPS)
RUN apk add --no-cache git ca-certificates tzdata

# Configurar directorio de trabajo
WORKDIR /build

# Copiar archivos de dependencias primero (mejor uso de cache de Docker)
COPY go.mod go.sum ./

# Configurar GOTOOLCHAIN para permitir descarga automática de la versión requerida
ENV GOTOOLCHAIN=auto

# Descargar dependencias (se cachea si go.mod/go.sum no cambian)
RUN go mod download
RUN go mod verify

# Copiar el código fuente
COPY . .

# Compilar el binario con optimizaciones agresivas
# -ldflags: Flags del linker para reducir tamaño
#   -s: Eliminar tabla de símbolos y debug info
#   -w: Eliminar información DWARF debug
# CGO_ENABLED=0: Compilación estática sin dependencias de C (binario portable)
# -trimpath: Remueve paths absolutos del binario (seguridad y tamaño)
# -tags netgo: Usa implementación Go pura de networking (sin CGO)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-s -w -extldflags "-static"' \
    -trimpath \
    -tags netgo \
    -o factorit \
    ./cmd/factorit/main.go

# Verificar que el binario se creó correctamente
RUN ls -lh /build/factorit

# =============================================================================
# STAGE 2: Runtime - Imagen final minimalista
# =============================================================================
FROM alpine:3.19

# Instalar certificados CA para HTTPS y timezone data
RUN apk --no-cache add ca-certificates tzdata

# Crear usuario no-root para seguridad
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Configurar directorio de trabajo
WORKDIR /app

# Copiar binario desde el builder
COPY --from=builder /build/factorit /app/factorit

# Copiar archivos de configuración si existen (opcional)
# COPY --from=builder /build/config /app/config

# Cambiar permisos del binario
RUN chmod +x /app/factorit

# Cambiar a usuario no-root
USER appuser

# Exponer puerto de la API (ajustar según tu configuración)
EXPOSE 8080

# Health check (opcional pero recomendado)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Comando para ejecutar la aplicación
ENTRYPOINT ["/app/factorit"]

# =============================================================================
# Instrucciones de build:
# 
# Build de la imagen:
#   docker build -t factorit:latest .
#
# Build con optimización adicional (multi-plataforma):
#   docker buildx build --platform linux/amd64,linux/arm64 -t factorit:latest .
#
# Ver tamaño de la imagen:
#   docker images factorit:latest
#
# Tamaño esperado: ~15-25MB (depende de dependencias)
# =============================================================================
