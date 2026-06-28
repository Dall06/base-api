# BRO - Sistema de Alertas y Gestión de Pagos - Resumen de Implementación

**Fecha:** 2026-03-09
**Proyecto:** BRO (Sistema de Gestión de Gimnasios)

## Resumen Ejecutivo

Se ha implementado exitosamente el sistema de alertas en tiempo real y gestión de pagos recurrentes para el proyecto BRO. La implementación incluye:

1. ✅ Nuevos campos en Subscription para tracking de pagos
2. ✅ Enum PaymentStatus (paid, pending, overdue)
3. ✅ Rol "god" para el primer staff member
4. ✅ Use Case MarkPaid para marcar pagos como completados
5. ✅ Sistema de Alertas en tiempo real (sin tabla de alertas)
6. ✅ Nuevos métodos en repositories
7. ✅ Endpoints REST para alertas y mark-paid
8. ✅ Middlewares de autorización por rol
9. ✅ Migración SQL con índices optimizados

## Cambios Implementados

### 1. Domain Layer (`srv/gym/domain/`)

#### `entities.go`
- **Nuevo enum `PaymentStatus`:**
  ```go
  const (
      PaymentStatusPaid    PaymentStatus = "paid"
      PaymentStatusPending PaymentStatus = "pending"
      PaymentStatusOverdue PaymentStatus = "overdue"
  )
  ```

- **Nuevos campos en `Subscription`:**
  ```go
  NextPaymentDate *time.Time    `json:"next_payment_date,omitempty"`
  PaymentStatus   PaymentStatus `json:"payment_status"`
  ```

- **Nuevo método helper:**
  ```go
  func (s *Subscription) IsPaymentOverdue() bool
  ```

#### `models.go`
- **Nuevos tipos para Alertas:**
  ```go
  type AlertType string
  const (
      AlertTypePaymentOverdue       AlertType = "payment_overdue"
      AlertTypeInactiveMember       AlertType = "inactive_member"
      AlertTypeExpiringSubscription AlertType = "expiring_subscription"
  )

  type Alert struct {
      Type        AlertType
      MemberID    string
      MemberName  string
      Message     string
      Details     string
      DaysSince   *int
      DaysUntil   *int
  }

  type AlertsResponse struct {
      Alerts []Alert
      Total  int
  }
  ```

### 2. Company Domain (`srv/companies/domain/`)

#### `entities.go`
- **Nuevo rol God:**
  ```go
  const (
      StaffRoleGod   StaffRole = "god"   // Primer usuario con permisos totales
      StaffRoleOwner StaffRole = "owner"
      StaffRoleAdmin StaffRole = "admin"
      StaffRoleStaff StaffRole = "staff"
  )
  ```

### 3. Repositories (`srv/gym/repositories/`)

#### `subscription_repository.go`
- Actualizado `subscriptionDB` con nuevos campos
- Actualizado mapping functions (`toSubscriptionDB`, `toDomainSubscription`)
- **Nuevos métodos:**
  ```go
  GetOverdueSubscriptions(ctx context.Context) ([]domain.Subscription, error)
  GetExpiringSubscriptions(ctx context.Context, days int) ([]domain.Subscription, error)
  ```

#### `member_repository.go`
- **Nuevo método:**
  ```go
  GetInactiveMembers(ctx context.Context, days int) ([]domain.Member, error)
  ```

### 4. Ports (`srv/gym/ports/`)

#### `repositories.go`
- Actualizado `MemberRepository` interface
- Actualizado `SubscriptionRepository` interface

### 5. Use Cases

#### `srv/gym/usecases/subscription/mark_paid.go` (NUEVO)
```go
type MarkPaidUseCase struct {
    subRepo  ports.SubscriptionRepository
    planRepo ports.PlanRepository
}

func (uc *MarkPaidUseCase) Execute(ctx context.Context, id string) (*domain.SubscriptionResponse, error)
```

**Lógica:**
- Valida que la suscripción esté activa
- Marca `PaymentStatus = "paid"`
- Calcula `NextPaymentDate = NOW() + plan.DurationDays`
- Actualiza la suscripción

#### `srv/gym/usecases/alerts/get_recent.go` (NUEVO)
```go
type GetRecentUseCase struct {
    subRepo    ports.SubscriptionRepository
    memberRepo ports.MemberRepository
    planRepo   ports.PlanRepository
}

func (uc *GetRecentUseCase) Execute(ctx context.Context) (*domain.AlertsResponse, error)
```

**Alertas generadas:**
1. **Pagos vencidos:** `NextPaymentDate < NOW() AND PaymentStatus != 'paid'`
2. **Miembros inactivos:** `Status = 'inactive' AND UpdatedAt <= NOW() - 30 días`
3. **Suscripciones por vencer:** `EndDate BETWEEN NOW() AND NOW() + 7 días`

### 6. Handlers

#### `srv/gym/handlers/subscription.go`
- Actualizado constructor para incluir `MarkPaidUseCase`
- **Nuevo endpoint:**
  ```go
  func (h *SubscriptionHandler) MarkPaid(ctx echo.Context) error
  // POST /api/v1/subscriptions/:id/mark-paid
  ```

#### `srv/gym/handlers/alerts.go` (NUEVO)
```go
type AlertsHandler struct {
    getRecentUC *alerts.GetRecentUseCase
}

func (h *AlertsHandler) GetRecent(ctx echo.Context) error
// GET /api/v1/alerts/recent
```

### 7. Middlewares (`opt/middlewares/`)

#### `auth.go`
- **Nuevas funciones:**
  ```go
  func RequireGod() echo.MiddlewareFunc
  // Requiere role = "god"

  func RequireAdminOrGod() echo.MiddlewareFunc
  // Requiere role = "admin" OR "god"
  ```

### 8. Migración SQL

#### `scripts/migrations/001_add_payment_tracking_fields.sql` (NUEVO)
```sql
ALTER TABLE subscriptions
ADD COLUMN IF NOT EXISTS next_payment_date DATE;

ALTER TABLE subscriptions
ADD COLUMN IF NOT EXISTS payment_status VARCHAR(20) NOT NULL DEFAULT 'pending'
CHECK (payment_status IN ('paid', 'pending', 'overdue'));

-- Índices optimizados
CREATE INDEX idx_subscriptions_next_payment_date ON subscriptions(next_payment_date);
CREATE INDEX idx_subscriptions_payment_status ON subscriptions(payment_status);
CREATE INDEX idx_subscriptions_payment_alerts
  ON subscriptions(next_payment_date, payment_status, status)
  WHERE status = 'active';
```

## Integración en Router

Para integrar los nuevos endpoints, actualizar `opt/router/gym.go`:

```go
// 1. Importar el nuevo use case
import (
    alertsUC "base-api/srv/gym/usecases/alerts"
)

// 2. Inicializar use cases
markPaidSubUC := subscriptionUC.NewMarkPaidUseCase(subRepo, planRepo)
getAlertsUC := alertsUC.NewGetRecentUseCase(subRepo, memberRepo, planRepo)

// 3. Actualizar handler de subscriptions
subscriptionHandler := handlers.NewSubscriptionHandler(
    createSubUC,
    getActiveSubUC,
    listByMemberSubUC,
    paySubUC,
    cancelSubUC,
    markPaidSubUC,  // NUEVO
)

// 4. Inicializar handler de alertas
alertsHandler := handlers.NewAlertsHandler(getAlertsUC)

// 5. Registrar rutas
protected.POST("/subscriptions/:id/mark-paid", subscriptionHandler.MarkPaid)
protected.GET("/alerts/recent", alertsHandler.GetRecent)
```

## Endpoints API

### 1. Marcar Pago como Completado
```http
POST /api/v1/subscriptions/:id/mark-paid
Authorization: Bearer {jwt_token}

Response 200:
{
  "id": "uuid",
  "member_id": "uuid",
  "plan_id": "uuid",
  "plan_name": "Plan Mensual",
  "start_date": "2026-03-01T00:00:00Z",
  "end_date": "2026-03-31T23:59:59Z",
  "total_amount": 500.00,
  "amount_paid": 500.00,
  "left_amount": 0.00,
  "payment_status": "paid",
  "next_payment_date": "2026-04-09T00:00:00Z",
  "status": "active"
}
```

### 2. Obtener Alertas Recientes
```http
GET /api/v1/alerts/recent
Authorization: Bearer {jwt_token}

Response 200:
{
  "alerts": [
    {
      "type": "payment_overdue",
      "member_id": "uuid",
      "member_name": "Juan Pérez",
      "message": "Pago vencido",
      "details": "Pago vencido hace 3 días",
      "days_since": 3
    },
    {
      "type": "inactive_member",
      "member_id": "uuid",
      "member_name": "María García",
      "message": "Cliente inactivo",
      "details": "Inactivo por 45 días",
      "days_since": 45
    },
    {
      "type": "expiring_subscription",
      "member_id": "uuid",
      "member_name": "Pedro López",
      "message": "Suscripción por vencer",
      "details": "Vence en 5 días",
      "days_until": 5
    }
  ],
  "total": 3
}
```

## Configuración de Roles

### God Role (Primer Staff)
El rol "god" debe asignarse al primer staff creado en cada company. Esto puede hacerse:

1. **Automáticamente** durante el provisioning del tenant
2. **Manualmente** en la creación del primer usuario

```go
// Ejemplo en use case de creación de staff
if isFirstStaff {
    staff.Role = domain.StaffRoleGod
} else {
    staff.Role = domain.StaffRoleAdmin // o el que corresponda
}
```

### Uso de Middlewares
```go
// Solo God puede acceder
protected.DELETE("/critical-action", handler.Action, middlewares.RequireGod())

// Admin o God pueden acceder
protected.POST("/members", handler.Create, middlewares.RequireAdminOrGod())
```

## Decisiones de Diseño

### 1. Alertas en Tiempo Real
- ✅ No se creó tabla de alertas
- ✅ Se calculan on-demand mediante queries
- ✅ Índices optimizados para performance

### 2. Cero Días de Gracia
- ✅ `NextPaymentDate < NOW()` → inmediatamente overdue
- ✅ Sin tolerancia de días

### 3. God Role
- ✅ Primer staff member de cada company
- ✅ Permisos totales
- ✅ No puede ser modificado por otros admins

### 4. Inscripción Configurable
- ✅ Ya existe `InscriptionFee` en Plan
- ✅ No requiere cambios adicionales

## Pasos para Deployment

### 1. Ejecutar Migración
```bash
psql -h localhost -U postgres -d gym_tenant_001 -f scripts/migrations/001_add_payment_tracking_fields.sql
```

### 2. Actualizar Router
Modificar `opt/router/gym.go` según la sección "Integración en Router"

### 3. Rebuild y Deploy
```bash
go build -tags gym -o bin/gym-srv cmd/gym/main.go
./bin/gym-srv
```

### 4. Verificar Endpoints
```bash
# Health check
curl http://localhost:8082/api/v1/health

# Alertas (requiere JWT)
curl -H "Authorization: Bearer {token}" \
     http://localhost:8082/api/v1/alerts/recent
```

## Testing Sugerido

### 1. Unit Tests
- [ ] `MarkPaidUseCase.Execute()`
- [ ] `GetRecentUseCase.Execute()`
- [ ] `RequireGod()` middleware
- [ ] `RequireAdminOrGod()` middleware

### 2. Integration Tests
- [ ] POST /subscriptions/:id/mark-paid
- [ ] GET /alerts/recent
- [ ] Verificar cálculo de NextPaymentDate
- [ ] Verificar detección de pagos vencidos

### 3. Performance Tests
- [ ] Query de alertas con 10k+ subscriptions
- [ ] Índices funcionando correctamente

## Notas Adicionales

### Índices de Base de Datos
Los índices creados optimizan:
- Búsqueda por `next_payment_date`
- Filtrado por `payment_status`
- Query compuesto para alertas (partial index con `WHERE status = 'active'`)

### Extensibilidad
El sistema de alertas puede extenderse fácilmente:
```go
// Agregar nueva alerta
alerts = append(alerts, domain.Alert{
    Type:       "new_alert_type",
    MemberID:   member.ID,
    MemberName: member.Name,
    Message:    "Nuevo tipo de alerta",
})
```

### Rollback
Si necesitas revertir la migración:
```sql
DROP INDEX IF EXISTS idx_subscriptions_payment_alerts;
DROP INDEX IF EXISTS idx_subscriptions_payment_status;
DROP INDEX IF EXISTS idx_subscriptions_next_payment_date;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS payment_status;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS next_payment_date;
```

## Archivos Modificados/Creados

### Modificados
1. `/srv/gym/domain/entities.go`
2. `/srv/gym/domain/models.go`
3. `/srv/companies/domain/entities.go`
4. `/srv/gym/repositories/subscription_repository.go`
5. `/srv/gym/repositories/member_repository.go`
6. `/srv/gym/ports/repositories.go`
7. `/srv/gym/handlers/subscription.go`
8. `/opt/middlewares/auth.go`

### Creados
1. `/srv/gym/usecases/subscription/mark_paid.go`
2. `/srv/gym/usecases/alerts/get_recent.go`
3. `/srv/gym/handlers/alerts.go`
4. `/scripts/migrations/001_add_payment_tracking_fields.sql`
5. `/IMPLEMENTATION_SUMMARY.md` (este archivo)

---

**Implementación completada exitosamente** ✅

Para cualquier duda o modificación, revisar los archivos mencionados o consultar este documento.
