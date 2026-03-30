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

---

## Project Scope

- This repository is for the backend only.
- The primary deliverable is a production-ready REST API.
- Frontend, mobile apps, and third-party integrations are out of scope for the initial build.
- Cost optimization and AI-assisted planning are deferred until the core operational flows are stable.

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

### Phase 1 — Foundation

- Go project scaffold (Clean Architecture)
- MongoDB + Redis connection
- Environment config (.env, config loader)
- JWT + Refresh Token auth system
- RBAC middleware (6 roles)
- User management (CRUD)
- MongoDB indexes and bootstrap setup

### Phase 2 — Personnel & Compliance

- Personnel profile management
- Certificate management (BOSIET, DPR Offshore Permit, Scaffolding, role-specific)
- Compliance auto-validation engine
- Expiry reminder system (6 months, 4 months, 1 month before expiry)
- Travel blocking for non-compliant personnel
- Email + in-app alert dispatch

### Phase 3 — Vessel & Room Management

- Vessel setup (primary + secondary vessel)
- POB cap configuration per vessel
- Room management (create, assign, track)
- Real-time POB count via Redis
- Overshoot detection and alerting

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

1. Phase 1 — Foundation
2. Phase 2 — Personnel & Compliance
3. Phase 3 — Vessel & Room Management
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
| POST   | `/register`        | Register a new user                   | Public        |
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
| GET    | `/`         | List all users             | Sys Admin |
| GET    | `/:id`      | Get user by ID             | Sys Admin |
| PATCH  | `/:id`      | Update user                | Sys Admin |
| DELETE | `/:id`      | Deactivate user            | Sys Admin |
| PATCH  | `/:id/role` | Assign or change user role | Sys Admin |

---

### Vessels — `/api/v1/vessels`

| Method | Endpoint               | Description                     | Access                  |
| ------ | ---------------------- | ------------------------------- | ----------------------- |
| POST   | `/`                    | Create vessel or installation   | Sys Admin               |
| GET    | `/`                    | List all vessels                | Authenticated           |
| GET    | `/:id`                 | Get vessel details              | Authenticated           |
| PATCH  | `/:id`                 | Update vessel info              | Sys Admin               |
| DELETE | `/:id`                 | Remove vessel                   | Sys Admin               |
| GET    | `/:id/pob`             | Real-time POB count (Redis)     | Authenticated           |
| GET    | `/:id/capacity`        | POB capacity status             | Authenticated           |
| PATCH  | `/:id/pob-cap`         | Set or update POB cap           | Sys Admin               |
| GET    | `/:id/manifest`        | Full POB manifest snapshot      | Planner, OIM, Sys Admin |

---

### Rooms — `/api/v1/vessels/:vesselId/rooms`

| Method | Endpoint         | Description                | Access             |
| ------ | ---------------- | -------------------------- | ------------------ |
| POST   | `/`              | Create a room              | Sys Admin          |
| GET    | `/`              | List all rooms on vessel   | Authenticated      |
| GET    | `/:id`           | Get room details           | Authenticated      |
| PATCH  | `/:id`           | Update room info           | Sys Admin          |
| DELETE | `/:id`           | Delete room                | Sys Admin          |
| GET    | `/:id/occupants` | Get current room occupants | Planner, Sys Admin |

---

### Offshore Roles — `/api/v1/roles`

| Method | Endpoint            | Description                     | Access             |
| ------ | ------------------- | ------------------------------- | ------------------ |
| POST   | `/`                 | Create offshore role            | Sys Admin          |
| GET    | `/`                 | List all roles                  | Authenticated      |
| GET    | `/:id`              | Get role details                | Authenticated      |
| PATCH  | `/:id`              | Update role                     | Sys Admin          |
| DELETE | `/:id`              | Deactivate role                 | Sys Admin          |
| POST   | `/:id/assign`       | Assign personnel to role        | Sys Admin          |
| GET    | `/:id/personnel`    | List personnel assigned to role | Planner, Sys Admin |
| POST   | `/:id/back-to-back` | Set back-to-back personnel pair | Sys Admin          |

---

### Personnel — `/api/v1/personnel`

| Method | Endpoint                    | Description                   | Access                           |
| ------ | --------------------------- | ----------------------------- | -------------------------------- |
| POST   | `/`                         | Create personnel record       | Sys Admin                        |
| GET    | `/`                         | List all personnel            | Planner, Safety Admin, Sys Admin |
| GET    | `/:id`                      | Get personnel details         | Authenticated                    |
| PATCH  | `/:id`                      | Update personnel info         | Sys Admin                        |
| DELETE | `/:id`                      | Remove personnel              | Sys Admin                        |
| GET    | `/:id/certificates`         | Get all certificates          | Safety Admin, Sys Admin, Own     |
| POST   | `/:id/certificates`         | Upload certificate (PDF)      | Safety Admin, Sys Admin          |
| PATCH  | `/:id/certificates/:certId` | Update certificate record     | Safety Admin, Sys Admin          |
| DELETE | `/:id/certificates/:certId` | Delete certificate            | Safety Admin, Sys Admin          |
| GET    | `/:id/compliance`           | Get compliance status summary | Safety Admin, Planner, Sys Admin |

---

### Activities — `/api/v1/activities`

| Method | Endpoint          | Description                              | Access                    |
| ------ | ----------------- | ---------------------------------------- | ------------------------- |
| POST   | `/`               | Create activity                          | Activity Owner, Sys Admin |
| GET    | `/`               | List all activities                      | Authenticated             |
| GET    | `/:id`            | Get activity details                     | Authenticated             |
| PATCH  | `/:id`            | Update activity                          | Activity Owner, Sys Admin |
| DELETE | `/:id`            | Delete activity                          | Activity Owner, Sys Admin |
| POST   | `/:id/submit`     | Submit activity for planner review       | Activity Owner            |
| POST   | `/:id/approve`    | Approve activity                         | Planner                   |
| POST   | `/:id/reject`     | Reject activity                          | Planner                   |
| POST   | `/:id/reschedule` | Reschedule activity                      | Planner                   |
| GET    | `/gantt`          | Get all activities in Gantt-ready format | Planner, OIM, Sys Admin   |
| GET    | `/conflicts`      | Detect and list scheduling conflicts     | Planner                   |
| GET    | `/queue`          | Pending approval queue                   | Planner                   |

---

### Travel & Mobilization — `/api/v1/travel`

| Method | Endpoint                    | Description                           | Access             |
| ------ | --------------------------- | ------------------------------------- | ------------------ |
| POST   | `/transport`                | Create transport configuration        | Sys Admin          |
| GET    | `/transport`                | List all transport configurations     | Planner, Sys Admin |
| PATCH  | `/transport/:id`            | Update transport config               | Sys Admin          |
| DELETE | `/transport/:id`            | Remove transport                      | Sys Admin          |
| GET    | `/schedule`                 | View upcoming travel schedule         | Planner, Personnel |
| POST   | `/schedule/assign`          | Assign personnel to a trip            | Planner            |
| GET    | `/schedule/:id/utilization` | Get utilization status for a trip     | Planner            |
| GET    | `/alerts`                   | List low-utilization transport alerts | Planner            |

---

### Compliance — `/api/v1/compliance`

| Method | Endpoint                 | Description                                        | Access                       |
| ------ | ------------------------ | -------------------------------------------------- | ---------------------------- |
| GET    | `/expiring`              | Certificates expiring within N days                | Safety Admin, Planner        |
| GET    | `/expired`               | All expired certificates                           | Safety Admin                 |
| POST   | `/validate/:personnelId` | Validate personnel compliance for travel           | Safety Admin, Planner        |
| GET    | `/activity/:activityId`  | Compliance status for all personnel on an activity | Safety Admin, Activity Owner |

---

### Dashboard — `/api/v1/dashboard`

| Method | Endpoint                 | Description                           | Access                  |
| ------ | ------------------------ | ------------------------------------- | ----------------------- |
| GET    | `/`                      | Role-filtered full dashboard data     | Authenticated           |
| GET    | `/pob-today`             | Real-time POB vs Capacity             | Authenticated           |
| GET    | `/activities/upcoming`   | Activities in the next 7 days         | Authenticated           |
| GET    | `/certificates/expiring` | Certificates expiring in next 30 days | Safety Admin, Planner   |
| GET    | `/travel/upcoming`       | Travel schedule for next 7 days       | Planner, Personnel      |

---

### Reports — `/api/v1/reports`

| Method | Endpoint      | Description                             | Access                  |
| ------ | ------------- | --------------------------------------- | ----------------------- |
| GET    | `/daily`      | Daily POB report                        | Planner, OIM, Sys Admin |
| GET    | `/historical` | Historical POB data (date range filter) | Planner, OIM, Sys Admin |
| GET    | `/export/pdf` | Export report as PDF                    | Planner, Sys Admin      |
| GET    | `/export/csv` | Export report as CSV                    | Planner, Sys Admin      |

---

## Project Structure (Clean Architecture)

```
pob-backend/
├── cmd/
│   └── api/
│       └── main.go               # Entry point
├── internal/
│   ├── domain/                   # Entities and interfaces
│   │   ├── user.go
│   │   ├── vessel.go
│   │   ├── personnel.go
│   │   ├── activity.go
│   │   └── ...
│   ├── handlers/                 # HTTP layer (request/response)
│   │   ├── auth_handler.go
│   │   ├── vessel_handler.go
│   │   └── ...
│   ├── services/                 # Business logic
│   │   ├── auth_service.go
│   │   ├── compliance_service.go
│   │   └── ...
│   ├── repository/               # Database queries
│   │   ├── user_repo.go
│   │   ├── vessel_repo.go
│   │   └── ...
│   └── middleware/               # JWT, RBAC, logging
│       ├── auth.go
│       └── rbac.go
├── pkg/
│   ├── config/                   # Env config loader
│   ├── database/                 # PostgreSQL + Redis init
│   ├── jwt/                      # Token generation and parsing
│   └── utils/                    # Shared helpers
├── migrations/                   # SQL migration files
├── .env.example
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

_Version 1.0 — March 2026 | Testudo Nigeria Limited_
