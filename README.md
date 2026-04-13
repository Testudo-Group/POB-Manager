# POB Management System

**Personnel on Board (POB) Management System**
Built for Testudo Nigeria Limited — Offshore Oil & Gas Operations

A backend API for managing personnel presence, room allocation, activity planning, compliance tracking, travel logistics, minimum manning operations, and reporting across offshore vessels and installations.

---

## Tech Stack

| Layer        | Technology                                       |
| ------------ | ------------------------------------------------ |
| Backend      | Go (Golang) — Clean Architecture                 |
| API Style    | RESTful JSON                                     |
| Database     | MongoDB                                          |
| Cache        | Redis                                            |
| Auth         | JWT + Refresh Tokens + RBAC                      |
| File Storage | AWS S3 (certificate PDFs)                        |
| Hosting      | AWS / Azure                                      |
| Containers   | Docker                                           |
| Docs         | Postman Collection                               |

---

## Project Scope

- This repository is for the backend only.
- The primary deliverable is a production-ready REST API.
- Frontend, mobile apps, and third-party integrations are out of scope for the initial build.
- Cost optimization and AI-assisted planning are deferred until the core operational flows are stable.

---

## API Documentation

API documentation is provided as a **Postman Collection** located at:

```
docs/POB_Management_API.postman_collection.json
```

### Features
- All 35+ endpoints organized into 8 folders (Auth, Users, Positions, Personnel, Certificates, Notifications, Vessels, Rooms)
- Collection-level Bearer auth with automatic token management
- Auto-save test scripts — Login/Register responses automatically populate tokens and entity IDs
- Example request bodies with realistic offshore industry data
- Per-endpoint descriptions including RBAC permissions and enforcement rules

### Import into Postman
1. Open Postman
2. Click **File → Import**
3. Select `docs/POB_Management_API.postman_collection.json`
4. Update the `base_url` collection variable if your server is not running on `http://localhost:8080`

---

## User Roles

| Role                      | Description                                                                           |
| ------------------------- | ------------------------------------------------------------------------------------- |
| Activity Owner / Engineer | Creates and submits work activities; monitors personnel compliance                    |
| Planner                   | Reviews, approves, and optimizes all activities; manages room distribution and travel |
| Safety Admin              | Manages all certification records; enforces compliance before travel                  |
| OIM / Site Manager        | Read-only oversight; can trigger Minimum Manning Mode                                 |
| Personnel / Crew Member   | Views own profile, schedule, room assignment, and certification status                |
| System Admin              | Full platform configuration — vessel setup, user management, role definitions         |

---

## Role Permissions Matrix

| Feature             | Activity Owner | Planner | Safety Admin | OIM       | Personnel | Sys Admin |
| ------------------- | -------------- | ------- | ------------ | --------- | --------- | --------- |
| Create Activity     | ✅             | ✅      | ❌           | ❌        | ❌        | ✅        |
| Approve Activity    | ❌             | ✅      | ❌           | ❌        | ❌        | ❌        |
| Manage Gantt        | View Only      | Full    | ❌           | View Only | ❌        | ❌        |
| Manage Certificates | ❌             | ❌      | ✅           | ❌        | Own Only  | ✅        |
| Approve Travel      | ❌             | ✅      | ❌           | ❌        | ❌        | ❌        |
| View Cost Dashboard | ❌             | ✅      | ❌           | ✅        | ❌        | ✅        |
| Trigger Min Manning | ❌             | ❌      | ❌           | ✅        | ❌        | ✅        |
| Manage Roles        | ❌             | ❌      | ❌           | ❌        | ❌        | ✅        |
| Configure Vessel    | ❌             | ❌      | ❌           | ❌        | ❌        | ✅        |
| View Own Profile    | ✅             | ✅      | ✅           | ✅        | ✅        | ✅        |

---

## Development Phases

### ✅ Phase 1 — Foundation (Complete)

- [x] Go project scaffold (Clean Architecture)
- [x] MongoDB + Redis connection
- [x] Environment config (.env, config loader)
- [x] JWT + Refresh Token auth system
- [x] RBAC middleware (6 roles, 100+ permissions)
- [x] User management (CRUD)
- [x] MongoDB indexes and bootstrap setup

### ✅ Phase 2 — Personnel & Compliance (Complete)

- [x] Personnel profile management (CRUD)
- [x] Certificate management (BOSIET, DPR Offshore Permit, Scaffolding, role-specific)
- [x] Compliance auto-validation engine
- [x] Expiry reminder system (6 months, 4 months, 1 month before expiry)
- [x] Travel blocking for non-compliant personnel
- [x] In-app notification system + email dispatch stub

### ✅ Phase 3 — Vessel & Room Management (Complete)

- [x] Vessel setup (primary + secondary vessel)
- [x] POB cap configuration per vessel
- [x] Room management (create, assign, track)
- [x] Real-time POB count via Redis
- [x] Overshoot detection and hard blocking

### Phase 4 — Roles & Rotation

- Offshore role creation (Core / Flexible types)
- Rotation scheduling (e.g. 14/14, 28/28)
- Back-to-back personnel assignment (room stays with role)
- Dual role support (counted once in POB total)

### Phase 5 — Activity Management

- Activity creation (name, dates, duration, required roles, priority level)
- Pending review queue for Planner
- Approval / rejection / reschedule workflow
- Gantt chart data layer (activities, durations, personnel, priorities)
- Scheduling conflict detection and highlighting

### Phase 6 — Travel & Mobilization

- Transport configuration (Helicopter, Boat, Pickup, Hiace)
- Transport properties: capacity, cost per trip/seat, departure days, mobilization location
- Auto-match activity start dates to nearest transport schedule
- Personnel travel view (own assigned dates and transport)
- Utilization alerts (warn when trip is below 60% capacity)
- Trip consolidation suggestions

### Phase 7 — Minimum Manning Mode

- Activation by OIM or System Admin
- Reduce effective POB cap to pre-configured minimum safe level
- Auto-retain only Core roles
- Auto-suspend all non-Core activities
- Notify all affected personnel and activity owners
- Full audit trail (activated by, time, deactivated by, time)

### Phase 8 — Dashboard & Reporting

- Role-filtered dashboards per user type
- Real-time POB vs Capacity widget
- Upcoming activities (next 7 days)
- Expiring certificates (next 30 days)
- Upcoming travel schedule (next 7 days)
- Priority distribution of active activities
- Export reports as PDF and CSV

---

## Recommended Build Order

This is the lowest-rework backend delivery order based on feature dependencies:

1. ~~Phase 1 — Foundation~~ ✅
2. ~~Phase 2 — Personnel & Compliance~~ ✅
3. ~~Phase 3 — Vessel & Room Management~~ ✅
4. Phase 4 — Roles & Rotation
5. Phase 5 — Activity Management
6. Phase 6 — Travel & Mobilization
7. Phase 7 — Minimum Manning Mode
8. Phase 8 — Dashboard & Reporting

Dependency notes:

- Personnel and compliance should be implemented before travel and activity staffing rules.
- Vessel, room, and POB capacity rules should exist before minimum manning behavior is introduced.
- Roles and rotation should be stable before activity scheduling depends on them.
- Dashboards and reports should be built after the operational modules they summarize.

---

## API Endpoints

All routes are prefixed with `/api/v1`

---

### Auth — `/api/v1/auth`

| Method | Endpoint           | Description                           | Access        |
| ------ | ------------------ | ------------------------------------- | ------------- |
| POST   | `/register`        | Register a new organization + admin   | Public        |
| POST   | `/login`           | Login, returns access + refresh token | Public        |
| POST   | `/refresh`         | Refresh access token                  | Public        |
| POST   | `/logout`          | Invalidate refresh token              | Authenticated |
| GET    | `/me`              | Get current user profile              | Authenticated |
| PATCH  | `/me`              | Update current user profile           | Authenticated |
| POST   | `/change-password` | Change password                       | Authenticated |

---

### Users — `/api/v1/users`

| Method | Endpoint    | Description                | Access    |
| ------ | ----------- | -------------------------- | --------- |
| POST   | `/`         | Create user in org         | Sys Admin |
| GET    | `/`         | List all users             | Sys Admin |
| GET    | `/:id`      | Get user by ID             | Sys Admin |
| PATCH  | `/:id`      | Update user                | Sys Admin |
| DELETE | `/:id`      | Deactivate user            | Sys Admin |
| PATCH  | `/:id/role` | Assign or change user role | Sys Admin |

---

### Positions (Offshore Roles) — `/api/v1/positions`

| Method | Endpoint | Description                  | Access        |
| ------ | -------- | ---------------------------- | ------------- |
| POST   | `/`      | Create offshore role         | Authenticated |
| GET    | `/`      | List all roles               | Authenticated |

---

### Personnel — `/api/v1/personnel`

| Method | Endpoint                    | Description                   | Access        |
| ------ | --------------------------- | ----------------------------- | ------------- |
| POST   | `/`                         | Create personnel record       | Authenticated |
| GET    | `/`                         | List all personnel            | Authenticated |
| PATCH  | `/:id`                      | Update personnel info         | Authenticated |
| DELETE | `/:id`                      | Remove personnel              | Authenticated |
| GET    | `/:id/compliance`           | Get compliance status summary | Authenticated |
| POST   | `/:id/certificates`         | Add certificate               | Authenticated |
| GET    | `/:id/certificates`         | List certificates             | Authenticated |
| PATCH  | `/:id/certificates/:certId` | Update certificate record     | Authenticated |
| DELETE | `/:id/certificates/:certId` | Delete certificate            | Authenticated |

---

### Notifications — `/api/v1/notifications`

| Method | Endpoint       | Description                | Access        |
| ------ | -------------- | -------------------------- | ------------- |
| GET    | `/`            | Get user's notifications   | Authenticated |
| PATCH  | `/:id/read`    | Mark notification as read  | Authenticated |

---

### Vessels — `/api/v1/vessels`

| Method | Endpoint               | Description                     | Access        |
| ------ | ---------------------- | ------------------------------- | ------------- |
| POST   | `/`                    | Create vessel or installation   | Authenticated |
| GET    | `/`                    | List all vessels                | Authenticated |
| GET    | `/:id`                 | Get vessel details              | Authenticated |
| PATCH  | `/:id`                 | Update vessel info              | Authenticated |
| DELETE | `/:id`                 | Remove vessel                   | Authenticated |
| GET    | `/:id/pob`             | Real-time POB count (Redis)     | Authenticated |
| GET    | `/:id/manifest`        | Full POB manifest snapshot      | Authenticated |

---

### Rooms

**Under Vessel context — `/api/v1/vessels/:vesselId/rooms`**

| Method | Endpoint         | Description                | Access        |
| ------ | ---------------- | -------------------------- | ------------- |
| POST   | `/`              | Create a room              | Authenticated |
| GET    | `/`              | List all rooms on vessel   | Authenticated |
| POST   | `/assign`        | Assign personnel to room   | Authenticated |

**Direct access — `/api/v1/rooms`**

| Method | Endpoint         | Description                | Access        |
| ------ | ---------------- | -------------------------- | ------------- |
| GET    | `/:id`           | Get room details           | Authenticated |
| PATCH  | `/:id`           | Update room info           | Authenticated |
| DELETE | `/:id`           | Delete room                | Authenticated |
| GET    | `/:id/occupants` | Get current room occupants | Authenticated |

---

### Future Endpoints (Phase 4+)

The following endpoint groups are planned for future phases:

- **Activities** — `/api/v1/activities` (Phase 5)
- **Travel & Mobilization** — `/api/v1/travel` (Phase 6)
- **Compliance Reports** — `/api/v1/compliance` (Phase 6)
- **Dashboard** — `/api/v1/dashboard` (Phase 8)
- **Reports** — `/api/v1/reports` (Phase 8)

---

## Project Structure (Clean Architecture)

```
pob-management/
├── cmd/
│   └── api/
│       └── main.go                           # Entry point
├── config/
│   ├── config.go                             # Env config loader
│   └── role.config.go                        # RBAC roles & permissions matrix
├── internal/
│   ├── domain/                               # Entities
│   │   ├── user.go
│   │   ├── organization.go
│   │   ├── vessel.go
│   │   ├── room.go
│   │   ├── room_assignment.go
│   │   ├── personnel.go
│   │   ├── certificate.go
│   │   ├── certificate_type.go
│   │   ├── offshore_role.go
│   │   ├── notification.go
│   │   └── ...
│   ├── delivery/http/
│   │   ├── controllers/                      # HTTP handlers
│   │   │   ├── auth.controller.go
│   │   │   ├── user.controller.go
│   │   │   ├── vessel.controller.go
│   │   │   ├── room.controller.go
│   │   │   ├── personnel.controller.go
│   │   │   ├── certificate.controller.go
│   │   │   ├── offshore_role.controller.go
│   │   │   └── notification.controller.go
│   │   ├── middleware/                       # JWT, RBAC
│   │   │   ├── auth.go
│   │   │   └── rbac.go
│   │   └── routes/
│   │       └── routes.go                     # Route registration
│   ├── repository/                           # Database queries
│   │   ├── user.repository.go
│   │   ├── organization.repository.go
│   │   ├── vessel.repository.go
│   │   ├── room.repository.go
│   │   ├── room_assignment.repository.go
│   │   ├── personnel.repository.go
│   │   ├── certificate.repository.go
│   │   ├── certificate_type.repository.go
│   │   ├── offshore_role.repository.go
│   │   └── notification.repository.go
│   └── service/                              # Business logic
│       ├── auth.service.go
│       ├── token_manager.go
│       ├── user.service.go
│       ├── vessel.service.go
│       ├── room.service.go
│       ├── personnel.service.go
│       ├── certificate.service.go
│       ├── certificate_type.service.go
│       ├── offshore_role.service.go
│       ├── compliance.service.go
│       ├── reminder.service.go
│       └── notification.service.go
├── pkg/
│   ├── database/                             # MongoDB + Redis init
│   │   ├── mongo.go
│   │   └── redis.go
│   ├── logger/                               # Logging utilities
│   └── response/                             # Standardized API responses
├── docs/
│   ├── POB_Management_API.postman_collection.json   # Postman API docs
│   └── model-relationships.md                       # Entity relationship docs
├── .env
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

---

## Non-Functional Requirements

- Dashboard loads within 3 seconds under normal network
- API response times under 500ms for standard reads
- Gantt chart renders up to 50 simultaneous activities
- Supports up to 500 concurrent users
- 99.5% uptime target during business hours
- All data in transit encrypted via HTTPS/TLS
- Passwords hashed with bcrypt
- All user actions logged for audit
- File uploads validated for type and scanned
- Dates stored in UTC, displayed in user's local timezone
- Supports up to 10 vessels per organization

---

## System Constraints

- Backend API is the primary Phase 1 deliverable
- Certificate documents accepted in PDF format only
- AI optimization is out of scope for the initial backend phases
- No third-party ERP/HR integration in Phase 1 (API designed to support it later)
- All cost values stored and displayed in USD unless configured otherwise

---

_Version 1.1 — April 2026 | Testudo Nigeria Limited_
_Phases 1–3 complete. Phases 4–8 in progress._
