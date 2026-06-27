# Base API: Template & CLI Installer

Este repositorio contiene un generador y plantilla de arquitectura limpia en Go (Golang) diseñado para acelerar la inicialización de proyectos y asegurar el cumplimiento de principios de **Arquitectura Hexagonal (Ports & Adapters)** y **Clean Architecture**.

Incluye un instalador CLI interactivo escrito en Go que automatiza el andamiaje (*scaffolding*) de proyectos y servicios.

---

## 🛠️ Instalación y Uso del CLI (`hexcli`)

El instalador está ubicado en `cmd/hexcli`. Puedes compilarlo localmente con:

```bash
go build -o hexcli cmd/hexcli/main.go
```

### 1. Inicializar un nuevo proyecto (`hexcli init`)
Copia la estructura del template, inyectando el path de tu módulo de Go automáticamente y configurando los imports correctos.

```bash
./hexcli init <nombre-del-proyecto> --module github.com/tu-usuario/nombre-del-proyecto
```

### 2. Agregar un nuevo servicio hexagonal (`hexcli add service`)
Crea de manera automática la estructura de carpetas (`domain`, `ports`, `usecases`, `handlers`, `repositories`) para un servicio dentro de tu proyecto.

> [!NOTE]
> Debes ejecutar este comando desde la raíz de tu proyecto inicializado (donde exista el directorio `srv/`).

```bash
./hexcli add service <nombre-del-servicio>
```

---

## 🏗️ Estructura del Template

El proyecto generado incluye:
* Un entrypoint en `cmd/api/main.go` listo para levantar el servidor y manejar apagados graciosos (*graceful shutdown*).
* Paquetes comunes de utilidad en `pkg/` desacoplados (logging estructurado con sanitización de credenciales, respuestas HTTP estándar y errores personalizados).
* Un servicio completo de ejemplo (`srv/user/`) que expone endpoints de Registro, Inicio de Sesión y Perfil (`/api/v1/auth/signup`, `/login`, `/me`) utilizando persistencia en memoria y seguridad con contraseñas cifradas y JWT.
* Suite de pruebas basadas en tablas (*Table-Driven Tests*).
* Documentación interna de arquitectura.

---

## 🧪 Validar Cambios en el Template

Para validar que la plantilla compile y pase todos los tests correctamente:

```bash
cd cmd/hexcli/template
go test ./...
```
