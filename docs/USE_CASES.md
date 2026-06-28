# BRO Tool - Casos de Uso del Backend

## Arquitectura

```
┌─────────────────────────────────────────────────────────────┐
│                    srv/companies (Platform)                 │
│           DB Central - Control de todas las compañías       │
└─────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐
│   srv/gym         │ │   srv/gym         │ │   srv/gym         │
│   (Tenant A)      │ │   (Tenant B)      │ │   (Tenant C)      │
│   DB: gym_abc     │ │   DB: gym_xyz     │ │   DB: gym_123     │
└───────────────────┘ └───────────────────┘ └───────────────────┘
```

---

## 1. Platform (srv/companies)

### 1.1 Company (Gyms registrados)

| Caso de Uso | Descripción | Estado |
|-------------|-------------|--------|
| `CreateCompany` | Registrar un nuevo gym en la plataforma | ✅ Implementado |
| `GetCompany` | Obtener info de un gym por ID | ✅ Implementado |
| `GetCompanyBySlug` | Obtener gym por slug (para routing) | ✅ Repo listo |
| `ListCompanies` | Listar todos los gyms (filtrable por status) | ✅ Implementado |
| `UpdateCompany` | Actualizar datos de un gym | ✅ Repo listo |
| `UpdateCompanyStatus` | Cambiar status (pending→active→suspended) | ✅ Implementado |
| `DeleteCompany` | Eliminar gym (soft delete?) | ✅ Repo listo |

### 1.2 Subscription (Plan SaaS de cada gym)

| Caso de Uso | Descripción | Estado |
|-------------|-------------|--------|
| `CreateSubscription` | Crear suscripción SaaS para un gym | ⚪ Pendiente |
| `GetActiveSubscription` | Obtener suscripción activa de un gym | ✅ Repo listo |
| `ListSubscriptions` | Historial de suscripciones de un gym | ✅ Repo listo |
| `UpdateSubscription` | Modificar suscripción (upgrade/downgrade) | ⚪ Pendiente |
| `CancelSubscription` | Cancelar suscripción | ⚪ Pendiente |
| `CheckSubscriptionStatus` | Verificar si gym tiene acceso activo | ⚪ Pendiente |

### 1.3 Payment (Pagos de gyms a nosotros)

| Caso de Uso | Descripción | Estado |
|-------------|-------------|--------|
| `CreatePayment` | Registrar pago pendiente | ⚪ Pendiente |
| `MarkPaymentPaid` | Marcar pago como completado | ⚪ Pendiente |
| `ListPendingPayments` | Listar pagos pendientes (para cobranza) | ✅ Repo listo |
| `ListPaymentsByCompany` | Historial de pagos de un gym | ✅ Repo listo |

---

## 2. Tenant (srv/gym) - Por cada Gym

### 2.1 Staff (Usuarios admin del gym)

| Caso de Uso | Descripción | Estado |
|-------------|-------------|--------|
| `RegisterUser` | Registrar nuevo staff | ✅ Implementado |
| `Login` | Autenticar staff y generar JWT | ✅ Implementado |
| `GetUser` | Obtener info del usuario actual | ⚪ Pendiente |
| `UpdateProfile` | Actualizar perfil | ⚪ Pendiente |
| `UpdatePassword` | Cambiar contraseña | ⚪ Pendiente |

### 2.2 Plans (Membresías disponibles)

> **Nota:** En el código actual esto se llama `MembershipCatalog`

| Caso de Uso | Descripción | Estado |
|-------------|-------------|--------|
| `CreatePlan` | Crear plan (Mensual, Anual, etc.) | ⚪ Pendiente |
| `GetPlan` | Obtener un plan por ID | ⚪ Pendiente |
| `ListPlans` | Listar planes (activos o todos) | ⚪ Pendiente |
| `UpdatePlan` | Modificar plan | ⚪ Pendiente |
| `TogglePlanActive` | Activar/desactivar plan | ⚪ Pendiente |
| `DeletePlan` | Eliminar plan (soft delete) | ⚪ Pendiente |

### 2.3 Members (Clientes del gym)

> **Nota:** En el código actual esto se llama `Client`

| Caso de Uso | Descripción | Estado |
|-------------|-------------|--------|
| `CreateMember` | Registrar nuevo miembro | ⚪ Pendiente |
| `GetMember` | Obtener miembro por ID | ⚪ Pendiente |
| `GetMemberByQR` | Obtener miembro por código QR | ⚪ Pendiente |
| `ListMembers` | Listar miembros (filtrable) | ⚪ Pendiente |
| `UpdateMember` | Actualizar datos del miembro | ⚪ Pendiente |
| `SuspendMember` | Suspender miembro | ⚪ Pendiente |
| `DeleteMember` | Eliminar miembro (soft delete) | ⚪ Pendiente |

### 2.4 Subscriptions (Membresía activa de cada miembro)

> **Nota:** En el código actual esto se llama `Membership`

| Caso de Uso | Descripción | Estado |
|-------------|-------------|--------|
| `CreateSubscription` | Asignar membresía a miembro | ⚪ Pendiente |
| `GetActiveSubscription` | Obtener membresía activa | ⚪ Pendiente |
| `ListSubscriptionsByMember` | Historial de membresías | ⚪ Pendiente |
| `RenewSubscription` | Renovar membresía | ⚪ Pendiente |
| `CancelSubscription` | Cancelar membresía | ⚪ Pendiente |
| `CheckExpiringSoon` | Listar membresías por vencer (7 días) | ⚪ Pendiente |

### 2.5 Payments (Pagos de miembros al gym)

| Caso de Uso | Descripción | Estado |
|-------------|-------------|--------|
| `CreatePayment` | Registrar pago (completo o cuota) | ⚪ Pendiente |
| `MarkPaymentPaid` | Marcar cuota como pagada | ⚪ Pendiente |
| `ListPaymentsByMember` | Historial de pagos de un miembro | ⚪ Pendiente |
| `ListPendingPayments` | Pagos pendientes (para cobranza) | ⚪ Pendiente |
| `ListOverduePayments` | Pagos vencidos | ⚪ Pendiente |

### 2.6 Verification (Verificación de acceso)

| Caso de Uso | Descripción | Estado |
|-------------|-------------|--------|
| `VerifyMember` | Verificar acceso por QR (allow/warn/deny) | ⚪ Pendiente |
| `GetMemberStatus` | Estado completo de un miembro | ⚪ Pendiente |

---

## 3. Flujos Principales

### 3.1 Registro de Gym (Onboarding)

```
1. Admin de Bro → CreateCompany (status: pending)
2. Admin de Bro → CreateSubscription (plan: trial/basic/pro)
3. Sistema → Provisionar DB para el tenant
4. Admin de Bro → UpdateCompanyStatus (pending → active)
5. Owner del Gym → RegisterUser (role: owner)
6. Owner del Gym → Login
```

### 3.2 Registro de Miembro

```
1. Staff → CreateMember (genera QR automático)
2. Staff → CreateSubscription (asigna plan)
3. Staff → CreatePayment (registra pago o cuotas)
```

### 3.3 Verificación de Acceso (Escaneo QR)

```
1. Miembro escanea QR
2. Sistema → VerifyMember
   - Busca miembro por QR
   - Verifica status del miembro
   - Verifica suscripción activa
   - Verifica pagos pendientes
3. Sistema retorna:
   - ALLOW: Todo OK
   - WARN: Tiene pagos pendientes pero puede pasar
   - DENY: Sin membresía activa o suspendido
```

### 3.4 Cobranza (Nuestros pagos)

```
1. Sistema → ListPendingPayments (gyms que deben)
2. Admin de Bro → Contactar gym
3. Gym paga (transferencia)
4. Admin de Bro → MarkPaymentPaid
5. Si no paga → UpdateCompanyStatus (active → suspended)
```

---

## 4. Mapeo de Nomenclatura

| Código Actual | Nuevo Esquema | Descripción |
|---------------|---------------|-------------|
| `Client` | `Member` | Clientes del gym |
| `MembershipCatalog` | `Plan` | Tipos de membresía disponibles |
| `Membership` | `Subscription` | Membresía asignada a un miembro |
| `User` | `Staff` | Usuarios admin del gym |
| `Organization` | (eliminado) | Se reemplaza por tenant DB |

---

## 5. Prioridad de Implementación

### Fase 1 - MVP (Core)
1. ✅ Staff: Login, Register
2. ⚪ Plans: CRUD básico
3. ⚪ Members: CRUD + QR
4. ⚪ Subscriptions: Crear, Obtener activa
5. ⚪ VerifyMember: Verificación por QR

### Fase 2 - Pagos
1. ⚪ Payments: CRUD miembros
2. ⚪ VerifyMember: Incluir warning de pagos
3. ⚪ Dashboard: Stats básicos

### Fase 3 - Platform
1. ⚪ Companies: CRUD completo
2. ⚪ Company Subscriptions: Gestión SaaS
3. ⚪ Company Payments: Cobranza
4. ⚪ Tenant provisioning automático
