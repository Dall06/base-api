# Filosofía y Arquitectura del Proyecto

Este proyecto adopta **Arquitectura Hexagonal (Ports & Adapters)** y **Clean Architecture** para lograr una base de código desacoplada, mantenible y testeaable.

## 📐 Flujo de Dependencia

La regla fundamental de la Clean Architecture es que **las dependencias apuntan hacia adentro**. Las capas externas (como la base de datos o los controladores HTTP) conocen las capas internas (casos de uso, dominio), pero el dominio nunca sabe nada de la infraestructura exterior.

```mermaid
graph TD
    subgraph Adaptadores de Entrada (Externo)
        A[Controlador HTTP / Echo]
    end
    subgraph Puertos e Interfaces (Contratos)
        B[UserUsecase Interface]
        C[UserRepository Interface]
    end
    subgraph Casos de Uso (Lógica)
        D[UserUsecase Impl]
    end
    subgraph Dominio (Core Puro)
        E[User Entity]
    end
    subgraph Adaptadores de Salida (Externo)
        F[Repositorio SQL / Memoria]
    end

    A -->|llama| B
    D -->|implementa| B
    D -->|llama| C
    F -->|implementa| C
    D -->|manipula| E
    F -->|mapea a| E
```

## ⚖️ Reglas Strictas de Desarrollo

### 1. El Núcleo de pkg/
El directorio `pkg/` contiene paquetes genéricos de utilidad (logs, errores, validaciones, etc.). 
* **Regla estricta:** Un paquete dentro de `pkg/` **NO** puede importar a otro paquete dentro de `pkg/`. Si dos utilidades necesitan comunicarse, deben estar acopladas dentro de un solo paquete o inyectarse mediante interfaces independientes.

### 2. Conversión de Modelos
Las entidades definidas en `srv/*/domain/` son representaciones puras en Go de las reglas del negocio.
* No deben contener etiquetas de ORM (ej. `bun:"..."` o `gorm:"..."`).
* Los repositorios concretos son los encargados de traducir los modelos específicos de la base de datos (Ej: `UserDBStruct`) a las entidades puras del dominio antes de retornarlas al caso de uso.

### 3. Retornos Tempranos (No Else)
Para maximizar la legibilidad, evita usar bloques `else` cuando la condición inicial realiza un retorno inmediato.
```go
// Correcto
if err != nil {
    return nil, err
}
return data, nil
```
