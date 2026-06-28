# Base API: Go Modular Monorepo Boilerplate (Cleaned from bro-api)

Este repositorio contiene la arquitectura limpia, modular y robusta extraída de `bro-api`. Sirve como una plantilla (boilerplate) lista para producción con soporte para microservicios/monorepos modulares en Go.

Incluye la inyección de dependencias, base de datos con **Bun (PostgreSQL)**, mensajería con **NATS**, registros estructurados y sanitizados con **Slog**, seguridad de tokens con **JWT**, y un instalador CLI integrado.

---

## 🏗️ Arquitectura y Componentes

### 1. Servicios (`srv/` y `cmd/`)
* **`gateway` (API Gateway):** Corre en el puerto `8080`. Se encarga de la recuperación ante pánicos (*panic recover*), auditoría de operaciones, control de CORS, análisis de inyección SQL/XSS, y proxy de peticiones hacia el servicio correspondiente.
* **`user` (Servicio de Usuario):** Corre en el puerto `8081`. Conecta a PostgreSQL usando Bun ORM, crea la tabla de usuarios de manera automática si no existe, y proporciona registro (`POST /auth/signup`), inicio de sesión (`POST /auth/login`) y perfil de usuario autenticado (`GET /users/me`).

### 2. Módulos Auxiliares (`opt/`)
* **`db/`:** Manejador de pool de conexiones PostgreSQL con soporte de reintentos.
* **`middlewares/`:** Middlewares compartidos para Echo (IDs, loggers, verificación de firmas de Sigil).
* **`migrations/`:** Motor de migraciones de base de datos multitenant o plataforma.
* **`notify/`:** Motor de notificaciones.
* **`guardian/`:** Validador y gestor de permisos.

### 3. Utilidades Compartidas (`pkg/`)
* **`logs/`:** Logger estructurado JSON `slog` que enmascara de manera automática credenciales y contraseñas.
* **`errs/`:** Respuestas de error estandarizadas que mapean errores de dominio a códigos REST.
* **`jwt/`:** Generador y validador de tokens JSON Web Tokens.
* **`validation/`:** Lógica de validación de entradas.

---

## 🚀 Inicio Rápido

### Requisitos previos
* Go 1.25+ instalado.
* Docker y Docker Compose.

### Paso 1: Levantar infraestructura local
Levanta PostgreSQL y NATS usando:
```bash
docker-compose up -d
```

### Paso 2: Configurar entorno
Crea tu `.env` a partir de la plantilla:
```bash
cp .env.example .env
```

### Paso 3: Iniciar el Servicio de Usuario y el Gateway
En dos terminales separadas (o en segundo plano):
```bash
# Iniciar servicio de usuario (puerto 8081)
go run cmd/user/main.go

# Iniciar gateway (puerto 8080)
go run cmd/gateway/main.go
```

El API Gateway estará expuesto en `http://localhost:8080`.

---

## 🛠️ Uso del CLI de Instalación (`hexcli`)

El archivo `main.go` en la raíz actúa como el instalador de la plantilla. Compílalo localmente:

```bash
go build -o hexcli main.go
```

### Inicializar un nuevo proyecto
Copia de forma recursiva toda la arquitectura limpia de `github.com/diegoaleon/test-app` a un nuevo directorio, renombrando todos los imports del módulo de Go al path de tu elección:

```bash
./hexcli init <nombre-directorio> --module github.com/tu-usuario/nombre-proyecto
```

### Agregar un nuevo servicio
Genera el andamiaje hexagonal de un nuevo servicio (`domain`, `ports`, `usecases`, `handlers`, `repositories`):

```bash
# Ejecutar desde la raíz del nuevo proyecto generado
./hexcli add service billing
```

---

## 🧪 Pruebas unitarias
Para validar que todo compile y pase los tests de casos de uso del servicio de usuario:
```bash
go test ./...
```
