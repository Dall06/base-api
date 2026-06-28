# WhatsApp Integration - BRO Platform

## Resumen de implementacion

### Arquitectura

```
pkg/whatsapp/            ← Cliente HTTP para Meta Cloud API (Graph API v21.0)
├── client.go            ← New(), SendTemplate(), SendText(), tipos de request/response
└── client_test.go       ← 6 tests unitarios

opt/whatsapp/            ← Servicio de mensajeria de alto nivel (depende de pkg/whatsapp)
└── messenger.go         ← SendMemberWelcome(), SendMerchantWelcome(), SendExpirationReminder()
```

`pkg/whatsapp` es un wrapper puro sobre la Graph API de Meta sin dependencias de negocio. `opt/whatsapp` depende de `pkg/whatsapp` y expone metodos con parametros de dominio de BRO.

### Archivos modificados

| Archivo | Cambio |
|---------|--------|
| `pkg/whatsapp/client.go` | Nuevo: cliente Meta Cloud API |
| `pkg/whatsapp/client_test.go` | Nuevo: 6 tests (config, send, error) |
| `opt/whatsapp/messenger.go` | Nuevo: servicio BRO (welcome member, welcome merchant, expiration) |
| `env/config/config.go` | Agregado: `WhatsAppPhoneNumberID`, `WhatsAppAccessToken` |
| `cmd/gym/main.go` | Inyeccion de WhatsApp messenger al scheduler |
| `opt/router/gym.go` | Inyeccion de WhatsApp messenger al crear miembros |
| `opt/router/companies.go` | Inyeccion de WhatsApp messenger al registrar merchants |
| `srv/gym/usecases/member/create.go` | Envia bienvenida + QR al crear miembro |
| `srv/companies/usecases/staff/register.go` | Envia bienvenida al registrar merchant |
| `srv/gym/usecases/scheduler/scheduler.go` | Integra recordatorios de expiracion en el ciclo horario |
| `srv/gym/usecases/scheduler/send_reminders.go` | Nuevo: query SQL + envio de recordatorios |
| `srv/gym/usecases/member/create_test.go` | Actualizado: nueva firma de constructor |
| `srv/gym/handlers/member_test.go` | Actualizado: nueva firma de constructor |
| `srv/companies/usecases/staff/register_test.go` | Actualizado: nueva firma de constructor |

### Puntos de integracion

#### 1. Bienvenida al miembro (tenant)

**Trigger:** `POST /api/v1/members` — al crear un miembro exitosamente.

**Comportamiento:** Fire-and-forget en goroutine. Si el miembro no tiene telefono o WhatsApp no esta configurado, se skipea silenciosamente.

**Template:** `bro_member_welcome`
- Parametros: nombre del miembro, nombre del gym, codigo QR
- Header opcional: imagen del QR

#### 2. Bienvenida al merchant (platform)

**Trigger:** `POST /api/v1/auth/register` — al registrar un nuevo gym.

**Comportamiento:** Fire-and-forget en goroutine, igual que la notificacion de Trello.

**Template:** `bro_merchant_welcome`
- Parametros: nombre del dueno, nombre del gym, URL de acceso

#### 3. Recordatorio de expiracion (scheduler)

**Trigger:** Scheduler automatico cada hora.

**Comportamiento:** Para cada tenant, ejecuta un query SQL que busca suscripciones activas con `end_date` entre hoy y 7 dias. Envia un mensaje por cada miembro con telefono.

**Template:** `bro_expiration_reminder`
- Parametros: nombre del miembro, nombre del gym, nombre del plan, dias restantes

**Query SQL:**
```sql
SELECT m.name, m.phone, p.name AS plan_name, s.end_date
FROM subscriptions s
JOIN members m ON m.id = s.member_id
JOIN plans p ON p.id = s.plan_id
WHERE s.status = 'active'
  AND s.end_date BETWEEN NOW() AND NOW() + 7 days
  AND m.phone IS NOT NULL AND m.phone != ''
  AND m.status = 'active'
```

### Variables de entorno

```bash
WHATSAPP_PHONE_NUMBER_ID=   # ID del numero de telefono en Meta (no es el numero)
WHATSAPP_ACCESS_TOKEN=       # Token permanente del system user
```

Si ambas estan vacias, todas las funciones de WhatsApp se desactivan silenciosamente (modo desarrollo).

### Costos estimados

| Volumen | Costo Meta (utility LATAM) | Markup | Total |
|---------|---------------------------|--------|-------|
| 1,000 msgs/mes | ~$0.80 | $0 | ~$0.80 |
| 4,000 msgs/mes | ~$3.20 | $0 | ~$3.20 |
| 10,000 msgs/mes | ~$8.00 | $0 | ~$8.00 |
| 40,000 msgs/mes | ~$32.00 | $0 | ~$32.00 |

Meta Cloud API no cobra markup. Solo pagas la tarifa de Meta por conversacion (utility ~$0.0008 USD en LATAM).

---

## Pasos para activar WhatsApp

### Requisitos previos

- Una cuenta de Facebook personal (la del admin de BRO)
- Un numero de telefono dedicado que NO este registrado en WhatsApp personal ni WhatsApp Business app (compra un chip prepago nuevo, cuesta ~$3-5 USD)
- El chip debe poder recibir SMS o llamadas para verificacion

### Paso 1: Crear cuenta de Meta Business

1. Ve a [business.facebook.com](https://business.facebook.com)
2. Click en **Crear cuenta**
3. Llena:
   - **Nombre del negocio:** `BRO Gym Management` (o el nombre legal de tu empresa)
   - **Tu nombre:** tu nombre real
   - **Email del negocio:** un email profesional (ej: `admin@brogym.app`)
4. Confirma el email que te llega
5. Meta te asigna un **Business Manager ID** — guardalo

### Paso 2: Crear una app en Meta Developers

1. Ve a [developers.facebook.com/apps](https://developers.facebook.com/apps)
2. Click **Crear app**
3. Selecciona **Otro** como tipo de app
4. Selecciona **Empresa** como tipo
5. Llena:
   - **Nombre de la app:** `BRO WhatsApp`
   - **Email de contacto:** tu email
   - **Business Manager:** selecciona el que creaste en paso 1
6. Click **Crear app**
7. En la pantalla de productos, busca **WhatsApp** y click **Configurar**

### Paso 3: Agregar tu numero de telefono

1. En el dashboard de la app, ve a **WhatsApp > Configuracion de la API** (o **API Setup** en ingles)
2. Veras un numero de prueba que Meta te da gratis — ignoralo, es temporal
3. Click en **Agregar numero de telefono** (Add phone number)
4. Llena:
   - **Nombre para mostrar:** `BRO` (este es el nombre que veran los clientes)
   - **Categoria:** `Servicios comerciales` o `Fitness`
5. Ingresa el numero del chip prepago nuevo (formato internacional, ej: `+52 55 1234 5678`)
6. Selecciona **SMS** como metodo de verificacion
7. Ingresa el codigo que llega al chip
8. Una vez verificado, Meta te asigna un **Phone Number ID** — este es el valor para `WHATSAPP_PHONE_NUMBER_ID`

> **Importante:** Una vez registrado aqui, ese numero ya no funciona con la app de WhatsApp personal. Es exclusivo para la API.

### Paso 4: Crear un System User y token permanente

El token de prueba que Meta te da dura 24 horas. Necesitas un token permanente.

1. Ve a [business.facebook.com/settings/system-users](https://business.facebook.com/settings/system-users)
2. Click **Agregar** (Add)
3. Llena:
   - **Nombre:** `base-api-server`
   - **Rol:** `Admin`
4. Click **Crear usuario del sistema**
5. Click en el usuario recien creado
6. Click **Agregar activos** (Add Assets)
7. Selecciona **Apps** > selecciona **BRO WhatsApp** > activa **Control total**
8. Click **Guardar cambios**
9. Click **Generar nuevo token** (Generate New Token)
10. Selecciona la app **BRO WhatsApp**
11. Marca estos permisos:
    - `whatsapp_business_messaging`
    - `whatsapp_business_management`
12. Click **Generar token**
13. **Copia el token inmediatamente** — no se vuelve a mostrar. Este es el valor para `WHATSAPP_ACCESS_TOKEN`

### Paso 5: Crear los templates de mensaje

Los templates deben ser aprobados por Meta antes de poder usarlos. Solo se envian una vez para aprobacion.

1. Ve a [business.facebook.com](https://business.facebook.com)
2. En el menu lateral: **WhatsApp Manager** > **Herramientas de cuenta** > **Plantillas de mensaje** (Message Templates)
3. Click **Crear plantilla**

#### Template 1: Bienvenida al miembro

| Campo | Valor |
|-------|-------|
| **Nombre** | `bro_member_welcome` |
| **Categoria** | `Utility` |
| **Idioma** | `Espanol (es)` |
| **Header** | Tipo: Imagen (opcional — para el QR) |
| **Body** | `Hola {{1}}, bienvenido a {{2}}! 🏋️ Tu codigo de acceso es: {{3}}. Presentalo en recepcion para ingresar.` |

- Click **Enviar** para revision

#### Template 2: Bienvenida al merchant

| Campo | Valor |
|-------|-------|
| **Nombre** | `bro_merchant_welcome` |
| **Categoria** | `Utility` |
| **Idioma** | `Espanol (es)` |
| **Body** | `Hola {{1}}, tu gimnasio {{2}} ya esta activo en BRO! 🎉 Accede a tu panel de administracion en {{3}} para comenzar a gestionar tus clientes.` |

- Click **Enviar** para revision

#### Template 3: Recordatorio de expiracion

| Campo | Valor |
|-------|-------|
| **Nombre** | `bro_expiration_reminder` |
| **Categoria** | `Utility` |
| **Idioma** | `Espanol (es)` |
| **Body** | `Hola {{1}}, tu membresia en {{2}} (plan {{3}}) vence en {{4}} dias. Acercate al gimnasio para renovar y seguir entrenando! 💪` |

- Click **Enviar** para revision

> **Nota sobre aprobacion:** Los templates de tipo `Utility` se aprueban en minutos a 48 horas. Si te rechazan alguno, revisa que no tenga lenguaje promocional agresivo. Meta es estricto con marketing pero flexible con utility/transaccional.

### Paso 6: Verificar el negocio (requerido para produccion)

Para enviar mensajes a numeros que no son de prueba, Meta requiere verificar tu negocio.

1. Ve a **Meta Business Suite** > **Configuracion** > **Centro de seguridad** (Security Center)
2. Click **Iniciar verificacion**
3. Sube uno de estos documentos:
   - Acta constitutiva de la empresa
   - Registro fiscal (RFC/RUC/NIT segun tu pais)
   - Recibo de servicios a nombre de la empresa
   - Extracto bancario empresarial
4. Llena la informacion legal del negocio
5. Meta revisa en 1-3 dias habiles
6. Una vez verificado, tu limite de mensajes sube automaticamente

> **Limites de envio (antes de verificacion):** 250 conversaciones/24h.
> **Despues de verificacion:** 1,000/24h inicialmente, sube automaticamente a 10K, 100K, ilimitado segun tu quality rating.

### Paso 7: Configurar las variables de entorno

Agrega a tu archivo `.env` o sistema de secretos:

```bash
# WhatsApp - Meta Cloud API
WHATSAPP_PHONE_NUMBER_ID=123456789012345
WHATSAPP_ACCESS_TOKEN=EAAxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### Paso 8: Deploy y verificacion

1. Deploya el servicio `gym` y `companies` con las nuevas variables
2. Verifica en los logs:
   - Si las credenciales estan configuradas, veras mensajes de envio exitoso
   - Si estan vacias, todo se skipea silenciosamente (modo dev)
3. Prueba creando un miembro con telefono — deberia llegarle el WhatsApp de bienvenida
4. Prueba registrando un gym nuevo — deberia llegarle al dueno el WhatsApp de bienvenida

### Monitoreo

- **Quality Rating:** Revisalo en WhatsApp Manager > Numeros de telefono. Si baja a `Low`, Meta restringe tu envio.
- **Logs:** Todos los envios se loguean via `slog` con nivel `Info` (exito) o `Error` (fallo).
- **Errores comunes:**
  - `131026` — El numero no tiene WhatsApp
  - `131047` — Template no aprobado o nombre incorrecto
  - `131049` — Limite de mensajes alcanzado
  - `131051` — Template con parametros incorrectos

### Anti-spam: buenas practicas

- Usa solo templates de tipo **Utility** (no Marketing) para recordatorios y bienvenidas
- Nunca envies mas de 2-3 mensajes por miembro por mes
- Incluye opcion de opt-out en el texto (ej: "Responde STOP para dejar de recibir")
- Si el quality rating baja, pausa envios y revisa los reportes en WhatsApp Manager
- Los miembros pueden bloquear el numero — eso afecta tu quality rating
