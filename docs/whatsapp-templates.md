# WhatsApp Templates — Meta Business Manager

Todos los templates deben crearse en [Meta Business Manager](https://business.facebook.com/wa/manage/message-templates/) con idioma **Español (es)**.

## Templates existentes en el código

Estos templates ya están implementados en `opt/whatsapp/messenger.go`. Deben existir en Meta para que funcionen.

### 1. `bro_member_welcome`
**Trigger:** Al crear un nuevo miembro desde la app.
**Categoría:** UTILITY
**Parámetros:**
| # | Variable | Ejemplo |
|---|----------|---------|
| 1 | `{{1}}` — Nombre del miembro | Juan Pérez |
| 2 | `{{2}}` — Nombre del gym | Iron Gym |
| 3 | `{{3}}` — Código QR (BRO-xxx) | BRO-47697c3b61d1 |

**Texto sugerido:**
```
¡Hola {{1}}! 👋

Bienvenido a *{{2}}*. Tu código de acceso es:

🔑 *{{3}}*

Presenta este código o tu QR en la entrada para ingresar. ¡Nos vemos en el gym! 💪
```

---

### 2. `bro_merchant_welcome`
**Trigger:** Al crear una nueva cuenta de gym (tenant).
**Categoría:** UTILITY
**Parámetros:**
| # | Variable | Ejemplo |
|---|----------|---------|
| 1 | `{{1}}` — Nombre del dueño | Carlos López |
| 2 | `{{2}}` — Nombre del gym | Iron Gym |
| 3 | `{{3}}` — URL del panel | https://demo.brogym.app |

**Texto sugerido:**
```
¡Hola {{1}}! 🎉

Tu gimnasio *{{2}}* ya está activo en BRO.

Accede a tu panel de administración:
🔗 {{3}}

Desde ahí puedes agregar clientes, crear planes y gestionar todo tu gym. ¿Necesitas ayuda? Escríbenos.
```

---

### 3. `bro_expiration_reminder`
**Trigger:** Automático (cron cada 30 min) cuando la membresía expira en N días.
**Categoría:** UTILITY
**Parámetros:**
| # | Variable | Ejemplo |
|---|----------|---------|
| 1 | `{{1}}` — Nombre del miembro | Juan Pérez |
| 2 | `{{2}}` — Nombre del gym | Iron Gym |
| 3 | `{{3}}` — Nombre del plan | Mensualidad Básica |
| 4 | `{{4}}` — Días restantes | 3 |

**Texto sugerido:**
```
Hola {{1}}, te recordamos que tu membresía *{{3}}* en *{{2}}* vence en *{{4}} día(s)*.

Renueva antes de que expire para no perder acceso. Acércate a recepción o contacta al gym.
```

---

### 4. `bro_payment_received`
**Trigger:** Cuando el gym recibe un comprobante de pago (flujo Stripe/manual).
**Categoría:** UTILITY
**Parámetros:**
| # | Variable | Ejemplo |
|---|----------|---------|
| 1 | `{{1}}` — Nombre del dueño/gym | Carlos López |
| 2 | `{{2}}` — Nombre del gym | Iron Gym |
| 3 | `{{3}}` — Monto | $599 MXN |
| 4 | `{{4}}` — Nombre del plan | Mensualidad Premium |
| 5 | `{{5}}` — Ciclo de facturación | Mensual |

**Texto sugerido:**
```
Hola {{1}}, hemos recibido un comprobante de pago para *{{2}}*:

💰 Monto: *{{3}}*
📋 Plan: *{{4}}*
🔄 Ciclo: *{{5}}*

Tu pago está siendo validado. Te confirmaremos cuando esté procesado.
```

---

### 5. `bro_payment_confirmed`
**Trigger:** Cuando el admin marca un pago como confirmado.
**Categoría:** UTILITY
**Parámetros:**
| # | Variable | Ejemplo |
|---|----------|---------|
| 1 | `{{1}}` — Nombre del miembro | Juan Pérez |
| 2 | `{{2}}` — Nombre del gym | Iron Gym |
| 3 | `{{3}}` — Nombre del plan | Mensualidad Básica |
| 4 | `{{4}}` — Monto | $500 MXN |

**Texto sugerido:**
```
¡Hola {{1}}! ✅

Tu pago de *{{4}}* para el plan *{{3}}* en *{{2}}* ha sido confirmado.

Tu membresía está activa. ¡Nos vemos en el gym! 💪
```

---

### 6. `bro_payment_overdue`
**Trigger:** Automático (cron) cuando un pago pasa de su fecha límite.
**Categoría:** UTILITY
**Parámetros:**
| # | Variable | Ejemplo |
|---|----------|---------|
| 1 | `{{1}}` — Nombre del miembro | Juan Pérez |
| 2 | `{{2}}` — Nombre del gym | Iron Gym |
| 3 | `{{3}}` — Nombre del plan | Mensualidad Básica |

**Texto sugerido:**
```
Hola {{1}}, tu pago del plan *{{3}}* en *{{2}}* está vencido.

Por favor acércate a recepción para ponerte al día y no perder acceso al gym.
```

---

### 7. `bro_subscription_expired`
**Trigger:** Automático (cron) cuando la membresía pasa su fecha de vencimiento.
**Categoría:** UTILITY
**Parámetros:**
| # | Variable | Ejemplo |
|---|----------|---------|
| 1 | `{{1}}` — Nombre del miembro | Juan Pérez |
| 2 | `{{2}}` — Nombre del gym | Iron Gym |
| 3 | `{{3}}` — Nombre del plan | Mensualidad Básica |

**Texto sugerido:**
```
Hola {{1}}, tu membresía *{{3}}* en *{{2}}* ha expirado.

Renueva tu membresía para seguir entrenando. Acércate a recepción o contacta al gym para renovar.
```

---

### 8. `bro_company_suspended`
**Trigger:** Cuando la suscripción SaaS del gym es suspendida por falta de pago.
**Categoría:** UTILITY
**Parámetros:**
| # | Variable | Ejemplo |
|---|----------|---------|
| 1 | `{{1}}` — Nombre del dueño | Carlos López |
| 2 | `{{2}}` — Nombre del gym | Iron Gym |

**Texto sugerido:**
```
Hola {{1}}, la cuenta de *{{2}}* en BRO ha sido suspendida por falta de pago.

Tu panel y las funciones de verificación están temporalmente desactivados. Renueva tu suscripción para reactivar el servicio.
```

---

## Templates para Broadcast (texto libre, NO requieren aprobación de Meta)

Estos se envían como texto libre via `SendText()`. Solo funcionan dentro de la **ventana de 24 horas** (si el miembro te escribió en las últimas 24h). Si no hay ventana activa, el mensaje no se entrega.

Los textos están hardcodeados en `srv/gym/usecases/notification/broadcast.go`:

| ID | Nombre | Texto |
|---|--------|-------|
| `gym_closure` | Cierre temporal | "Estimado miembro, le informamos que el gimnasio estará cerrado temporalmente. Estaremos de vuelta pronto." |
| `schedule_change` | Cambio de horario | "Le informamos que ha habido un cambio de horario en nuestras clases. Consulte el calendario actualizado." |
| `promotion` | Promoción | "¡Aprovecha nuestra promoción especial! Contactanos para más información." |
| `general` | Aviso general | (texto libre del admin) |

> **Nota:** Para broadcasts confiables sin restricción de ventana de 24h, estos deberían convertirse en templates aprobados por Meta. Los nombres serían `bro_gym_closure`, `bro_schedule_change`, `bro_promotion`, `bro_general_announcement`.

---

## Resumen — Templates a crear en Meta Business Manager

| # | Template Name | Categoría | Params | Estado |
|---|--------------|-----------|--------|--------|
| 1 | `bro_member_welcome` | UTILITY | 3 (name, gym, qr_code) | **Requerido** |
| 2 | `bro_merchant_welcome` | UTILITY | 3 (owner, gym, url) | **Requerido** |
| 3 | `bro_expiration_reminder` | UTILITY | 4 (name, gym, plan, days) | **Requerido** |
| 4 | `bro_payment_received` | UTILITY | 5 (owner, gym, amount, plan, cycle) | **Requerido** |
| 5 | `bro_payment_confirmed` | UTILITY | 4 (name, gym, plan, amount) | **Requerido** |
| 6 | `bro_payment_overdue` | UTILITY | 3 (name, gym, plan) | **Requerido** |
| 7 | `bro_subscription_expired` | UTILITY | 3 (name, gym, plan) | **Requerido** |
| 8 | `bro_company_suspended` | UTILITY | 2 (owner, gym) | **Requerido** |
| 9 | `bro_gym_closure` | MARKETING | 0 | Opcional (mejora broadcast) |
| 10 | `bro_schedule_change` | MARKETING | 0 | Opcional (mejora broadcast) |
| 11 | `bro_promotion` | MARKETING | 0 | Opcional (mejora broadcast) |
| 12 | `bro_general_announcement` | MARKETING | 1 (message) | Opcional (mejora broadcast) |

Los templates 1-8 son **obligatorios** para que las notificaciones automáticas funcionen.
Los templates 9-12 son **opcionales** — mejoran la entrega de broadcasts al no depender de la ventana de 24h.
