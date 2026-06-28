# Guía Completa de Arquitectura (Hexagonal & Clean)

Este proyecto sirve como una plantilla de referencia para APIs en Go (Golang) y sigue de forma estricta los principios de **Arquitectura Hexagonal (Ports & Adapters)** y **Clean Architecture**. Su propósito principal es desacoplar las reglas de negocio de los detalles de infraestructura (base de datos, framework HTTP, librerías de terceros) para garantizar un sistema altamente testeable, mantenible y escalable.

---

## 📐 El Flujo de Control y Dependencias

La regla dorada de Clean Architecture es que **las dependencias de código solo pueden apuntar hacia adentro**. 

Las capas externas de infraestructura conocen e interactúan con las capas internas (Casos de Uso, Dominio), pero las capas internas no conocen absolutamente nada sobre bases de datos, APIs web o protocolos de mensajería externos.

```mermaid
graph TD
    subgraph Capa de Infraestructura (Adaptadores Externos)
        HTTP[Handlers HTTP / Echo]
        DB[PostgreSQL Repository / Bun]
    end
    
    subgraph Capa de Aplicación (Interfaces y Contratos)
        ports_in[Ports de Entrada / Usecase Interfaces]
        ports_out[Ports de Salida / Repository Interfaces]
    end
    
    subgraph Capa de Lógica de Negocio (Core)
        usecase[Casos de Uso / Interactors]
        domain[Entidades de Dominio / Domain Entities]
    end

    HTTP -->|1. Llama a través de interfaz| ports_in
    usecase -->|2. Implementa| ports_in
    usecase -->|3. Llama a través de interfaz| ports_out
    DB -->|4. Implementa| ports_out
    usecase -->|5. Procesa y muta| domain
```

## 🛠️ El Punto de Composición (Composition Root) en `cmd/`

Una de las decisiones arquitectónicas clave es mantener la **inicialización y el acoplamiento concretos en la capa de `cmd/`**, mientras que el código de negocio en `srv/` se mantiene completamente agnóstico y pasivo.

### 1. El rol de `cmd/<servicio>/main.go` (Composición)
Esta carpeta actúa como el **Punto de Composición (Composition Root)** de cada servicio. Es el único lugar autorizado para:
* Leer variables de entorno y archivos `.env` (Configuración).
* Inicializar los pools de conexiones físicas de bases de datos (`database.Connect`).
* Instanciar los adaptadores concretos (ej. `repositories.NewUserRepository(db)`).
* Inyectar las dependencias en los casos de uso (`usecases.NewUserUsecase(repo)`).
* Levantar el servidor HTTP (Echo) e iniciar la escucha.

### 2. El rol de `srv/<servicio>/` (Lógica Pura)
La carpeta `srv/` contiene la definición lógica del servicio. Sus subcarpetas son pasivas:
* **No** inicializan bases de datos ni leen variables de entorno.
* Reciben todas las dependencias ya resueltas a través de sus constructores (`NewUserUsecase`, `NewUserRepository`, etc.) usando **Inyección de Dependencias (DI)**.
* Esto permite cambiar la infraestructura (por ejemplo, reemplazar el repositorio de Postgres por uno de MongoDB o en memoria) simplemente modificando el archivo `main.go` en `cmd/`, sin alterar una sola línea de código dentro de `srv/`.

## ⚖️ Reglas Estrictas de Programación

---

## 📂 Explicación Detallada de Capas e Inyección de Dependencias

Cada servicio modular dentro de `srv/` (por ejemplo, `srv/user/`) está autocontenido y se divide en 5 capas claramente diferenciadas:

### 1. Dominio (`srv/*/domain/`)
Es el corazón del software. Contiene las estructuras que representan los conceptos de negocio (entidades) y la lógica inherente a ellos.
* **Regla estricta:** No debe importar ningún framework web, ORM, o base de datos.
* **Mapeo de base de datos:** Las entidades de dominio **no** deben tener etiquetas de bases de datos (como `bun:"..."` o `gorm:"..."`). En su lugar, si es necesario, los repositorios usan estructuras de base de datos dedicadas y mapean los datos a entidades limpias del dominio.
* *Ejemplo:* `srv/user/domain/user.go` define el modelo `User`, y métodos lógicos puros como `.CheckPassword(password)`.

### 2. Puertos (`srv/*/ports/`)
Contiene los contratos (interfaces) que definen cómo el mundo exterior puede interactuar con el dominio (Puertos de Entrada) y cómo el dominio puede interactuar con el mundo exterior (Puertos de Salida).
* **Ports de Entrada (Usecases):** Interfaces que consumen los Handlers HTTP.
* **Ports de Salida (Repositories):** Interfaces que definen cómo se guardan/leen los datos y que implementan los repositorios concretos.
* *Ejemplo:* `UserRepository` y `UserUsecase` en `srv/user/ports/ports.go`.

### 3. Casos de Uso (`srv/*/usecases/`)
Implementa las operaciones de negocio del software. Coordina la obtención de datos a través de los puertos de salida, ejecuta la lógica del dominio, y retorna los resultados a los puertos de entrada.
* **Inyección de Dependencias:** El Usecase recibe sus dependencias de infraestructura (como el Repositorio de base de datos) a través de las interfaces definidas en `ports/`.
* *Ejemplo:* `srv/user/usecases/user.go` cifra la contraseña, guarda al usuario a través del puerto `UserRepository` y genera el JWT de sesión.

### 4. Handlers / Adaptadores de Entrada (`srv/*/handlers/`)
Son los controladores de transporte. En este caso, adaptadores HTTP construidos sobre el framework **Echo**.
* **Responsabilidades:** Leer la petición del cliente, enlazar y validar los datos de entrada (BIND), llamar al caso de uso correspondiente, y traducir los errores de negocio a respuestas HTTP con códigos de estado REST adecuados (mediante `pkg/errs`).
* *Ejemplo:* `srv/user/handlers/http.go` maneja las rutas de Echo para `/api/v1/auth/login`.

### 5. Repositorios / Adaptadores de Salida (`srv/*/repositories/`)
Implementan los puertos de salida y manejan los accesos a los sistemas de almacenamiento reales.
* **Responsabilidad:** Realizar queries SQL, persistir datos, y conectarse a servicios de caché o APIs externas.
* *Ejemplo:* `srv/user/repositories/postgres.go` implementa la interfaz `UserRepository` usando el ORM `bun` y conectándose a PostgreSQL.

---

## ⚖️ Reglas Estrictas de Desarrollo (El ADN del Código)

Para garantizar la mantenibilidad a largo plazo de la base de código, se deben cumplir sin excepción las siguientes directrices:

### 1. El Directorio de Utilidades `pkg/` es Altamente Desacoplado
El directorio `pkg/` contiene paquetes genéricos que pueden ser compartidos por cualquier parte del sistema (ej: `logs`, `jwt`, `errs`).
* **Regla estricta:** Un paquete dentro de `pkg/` **NO** puede importar a otro paquete dentro del mismo `pkg/`. 
* *Razón:* Esto previene de forma estricta el acoplamiento circular. Si `pkg/logs` dependiera de `pkg/errs`, y `pkg/errs` necesitara registrar logs usando `pkg/logs`, la compilación fallaría. Las utilidades deben ser piezas de Lego completamente independientes.

### 2. Infraestructura y Middlewares en `opt/`
El directorio `opt/` aloja lógica de infraestructura específica y middlewares (ej: base de datos, autenticación de firmas Sigil, RBAC de Guardian, migraciones, integraciones con R2/Trello).
* A diferencia de `pkg/`, los paquetes de `opt/` pueden tener interacciones controladas entre sí para orquestar la infraestructura compartida.

### 3. Principio de No Else (Early Returns)
Para maximizar la legibilidad y evitar niveles innecesarios de indentación del código, queda prohibido el uso de bloques `else` cuando la condición inicial realiza un retorno inmediato.
```go
// ❌ INCORRECTO
if err != nil {
    return nil, err
} else {
    return data, nil
}

//  CORRECTO
if err != nil {
    return nil, err
}
return data, nil
```

### 4. Table-Driven Tests (Pruebas Basadas en Tablas)
Toda la lógica de negocio en `usecases/` debe estar probada exhaustivamente mediante pruebas unitarias estructuradas en tablas de casos de prueba (*test cases*), facilitando la validación de múltiples escenarios de éxito y fallo rápidamente.
