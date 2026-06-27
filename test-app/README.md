# API Hexagonal Boilerplate

Este es un template base y modular para crear APIs web en Go (Golang) utilizando la biblioteca **Echo** y siguiendo los principios de la **Arquitectura Hexagonal (Puertos y Adaptadores)** y **Clean Architecture**.

## 🚀 Inicio Rápido

### Requisitos previos
* Go 1.25+ instalado.
* Docker y Docker Compose (opcional, para base de datos).

### Paso 1: Configurar variables de entorno
Crea un archivo `.env` en la raíz del proyecto copiando la plantilla:
```bash
cp .env.example .env
```

### Paso 2: Levantar dependencias locales
```bash
docker-compose up -d
```

### Paso 3: Correr el servidor
```bash
go run cmd/api/main.go
```

El servidor estará corriendo en `http://localhost:8080`.

---

## 🏗️ Estructura del Proyecto

* **`cmd/api/main.go`**: Punto de entrada de la aplicación. Configura la inyección de dependencias y levanta el servidor web.
* **`pkg/`**: Paquetes de utilidad e infraestructura compartidos y desacoplados (Ej: logger estructurado, formateador de errores, utilidades de JWT).
* **`srv/`**: Servicios modulares autocontenidos. Cada servicio se divide en:
  * **`domain/`**: Entidades del modelo de negocio y reglas de dominio puras (sin librerías externas).
  * **`ports/`**: Interfaces que definen los contratos para los casos de uso (puertos de entrada) y repositorios (puertos de salida).
  * **`usecases/`**: Implementación de la lógica de negocio.
  * **`handlers/`**: Controladores de entrada (HTTP / Echo).
  * **`repositories/`**: Acceso y persistencia de datos (SQL, memoria, NoSQL).

---

## 🧪 Ejecutar Pruebas

Para ejecutar la suite de pruebas unitarias basadas en tablas (*Table-Driven Tests*):
```bash
go test ./... -v
```
