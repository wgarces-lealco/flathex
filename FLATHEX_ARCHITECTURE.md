# FLATHEX ARCHITECTURE: SYSTEM PROMPT & MANIFESTO

**[SYSTEM INSTRUCTION FOR AI AGENTS]**
You are strictly bound to the FlatHex architecture for Go. FlatHex combines Package-Oriented Design, Vertical Slicing (by Aggregates), and Hexagonal Architecture. Do not deviate from these rules, even if standard Go community patterns suggest otherwise.

---

## 0. STRICT PROHIBITIONS (DO NOT DO THIS)

* ❌ NO "God objects" or generic folders (`domain`, `application`, `utils`, `helpers`).
* ❌ NO framework tags (`json`, `gorm`, `bson`) inside `/internal/core/`.
* ❌ NO `log.Printf` or `fmt.Println` inside `/internal/core/`.
* ❌ NO imports between different aggregates inside `/internal/core/` (e.g., `core/invoices` CANNOT import `core/subscriptions`).

---

## 1. PHYSICAL STRUCTURE (DIRECTORY MAP)

Always generate files respecting this exact topology:

```text
/internal
├── platform/                   # 1. BASE TÉCNICA (Arranque)
│   ├── config/                 # Variables de entorno (.env)
│   └── secrets/                # Gestores de credenciales (Ej. AWS)
│
├── core/                       # 2. EL NÚCLEO (Reglas de negocio)
│   ├── invoices/               # -> AGREGADO 1 (Vertical Slice)
│   │   ├── entity.go           # Entidad Pura (Structs con comportamiento).
│   │   ├── errors.go           # Errores centinela del dominio (ErrNotFound).
│   │   ├── ports.go            # Interfaces de SALIDA requeridas por este agregado.
│   │   └── service.go          # Casos de uso (Lógica que orquesta la entidad y los puertos).
│   │
│   └── subscriptions/          # -> AGREGADO 2 (Totalmente aislado de invoices)
│
├── adapters/                   # 3. ADAPTADORES DE SALIDA (Infraestructura externa)
│   ├── postgres/               # Implementa los puertos de base de datos definidos en `core`.
│   └── stripe/                 # Implementa los puertos de APIs externas definidos en `core`.
│
└── presentation/               # 4. ADAPTADORES DE ENTRADA (Frontera exterior)
    └── rest/                   # Handlers, middlewares y rutas HTTP. Consume el `core`.
```

---

## 2. THE 5 GOLDEN RULES OF FLATHEX

### Rule 1: Consumer Defines the Interface (ISP)

The core NEVER exports interfaces for presentation to consume. The core exports concrete `*Service` structs. The consumer (presentation) defines the interface.

❌ **MAL (AI Hallucination):** Creating `type InvoiceService interface` in `core/invoices/service.go`.

✅ **BIEN (FlatHex Standard):**

```go
// internal/presentation/rest/invoice_handler.go
type InvoiceCreator interface {
    Create(ctx context.Context, inv *invoices.Invoice) error
}
type Handler struct { creator InvoiceCreator }
```

### Rule 2: Accept Interfaces, Return Structs in Core

✅ **BIEN:**

```go
// internal/core/invoices/service.go
func NewService(repo Repository) *Service { return &Service{repo: repo} }
```

### Rule 3: Anti-Corruption Layer (Strict Mapping)

- **Presentation:** DTOs with `json` tags live in `presentation`. The handler calls `dto.ToDomain()` before calling the core.
- **Adapters:** Models with `gorm` tags live in `adapters`. The adapter maps the core entity to the DB model before saving.

### Rule 4: Context Propagation

Every cross-boundary method MUST receive `ctx context.Context` as the first parameter.

✅ **BIEN:** `Save(ctx context.Context, inv *Invoice) error`

### Rule 5: Inter-Aggregate Isolation

Aggregates cannot talk to each other directly via imports.

❌ **MAL:** `core/subscriptions/service.go` importing `core/invoices`.

✅ **BIEN:** Orchestration happens in the presentation handler (calling both services), OR via Domain Events (e.g., publishing `InvoiceCreated` to a message broker, which `subscriptions` listens to via an adapter). Pass primitive IDs (`invoiceID string`), never structs across aggregate boundaries.

---

## 3. OBSERVABILITY AND ERROR HANDLING

Core **DOES NOT LOG**. The core's responsibility is to return rich, typed domain errors. The presentation layer is responsible for logging and mapping to HTTP status codes.

✅ **BIEN (Core):**

```go
// internal/core/invoices/errors.go
var ErrNotFound = errors.New("invoice not found")
```

✅ **BIEN (Presentation):**

```go
// internal/presentation/rest/handler.go
if err != nil {
    if errors.Is(err, invoices.ErrNotFound) {
        slog.Warn("invoice not found", "error", err)
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
}
```

---

## 4. TESTING CONVENTIONS

When generating tests, the AI must follow:

- **Location:** `_test.go` files live directly next to the file being tested (e.g., `service_test.go` alongside `service.go`).
- **Style:** Always use Table-Driven Tests (`[]struct{ name string ... }`).
- **Mocking Databases:** Use Fakes (in-memory maps) for repositories to test state. Do not use `mockery` for databases.
- **Mocking External APIs:** Use `mockery` for external I/O ports (e.g., Stripe, Email senders) to test interactions.

---

## 5. END-TO-END REFERENCE AGGREGATE (THE BLUEPRINT)

> **AI Agent:** Use this as your exact template when asked to create a new aggregate.

**`internal/core/payments/entity.go`**

```go
package payments

type Payment struct {
    id     string
    amount float64
}

func NewPayment(id string, amount float64) *Payment {
    return &Payment{id: id, amount: amount}
}
func (p *Payment) ID() string { return p.id }
```

**`internal/core/payments/errors.go`**

```go
package payments

import "errors"

var ErrInvalidAmount = errors.New("payment amount must be positive")
```

**`internal/core/payments/ports.go`**

```go
package payments

import "context"

type Repository interface {
    Save(ctx context.Context, p *Payment) error
}
```

**`internal/core/payments/service.go`**

```go
package payments

import "context"

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) Process(ctx context.Context, id string, amount float64) error {
    if amount <= 0 {
        return ErrInvalidAmount
    }
    payment := NewPayment(id, amount)
    return s.repo.Save(ctx, payment)
}
```
