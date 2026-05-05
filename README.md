# POB Management System
Personnel on Board (POB) Management System Built for Testudo Nigeria Limited — Offshore Oil & Gas Operations

A backend API for managing personnel presence, room allocation, activity planning, compliance tracking, travel logistics, minimum manning operations, and reporting across offshore vessels and installations.

## Tech Stack
| Layer          | Technology                          |
|----------------|-------------------------------------|
| Backend        | Go (Golang) — Clean Architecture    |
| API Style      | RESTful JSON                        |
| Database       | MongoDB                             |
| Cache          | Redis                               |
| Auth           | JWT + Refresh Tokens + RBAC         |
| File Storage   | AWS S3 (certificate PDFs)           |
| Hosting        | AWS / Azure                         |
| Containers     | Docker                              |
| Docs           | Postman Collection                  |

## Project Scope
This repository is for the backend only.  
The primary deliverable is a production-ready REST API.  
Frontend, mobile apps, and third-party integrations are out of scope for the initial build.  
Cost optimization and AI-assisted planning are deferred until the core operational flows are stable.

## API Documentation
API documentation is provided as a Postman Collection located at:

`docs/POB_Management_API.postman_collection.json`

### Features
- All 70+ endpoints organized into logical folders
- Collection-level Bearer auth with automatic token management
- Auto-save test scripts — Login/Register responses automatically populate tokens and entity IDs
- Example request bodies with realistic offshore industry data
- Per-endpoint descriptions including RBAC permissions and enforcement rules

### Import into Postman
1. Open Postman
2. Click **File → Import**
3. Select `docs/POB_Management_API.postman_collection.json`
4. Update the `base_url` collection variable if your server is not running on `http://localhost:8080`

## User Roles
| Role                        | Description                                                                              |
|-----------------------------|------------------------------------------------------------------------------------------|
| Activity Owner / Engineer   | Creates and submits work activities; monitors personnel compliance                        |
| Planner                     | Reviews, approves, and optimizes all activities; manages room distribution and travel     |
| Safety Admin                | Manages all certification records; enforces compliance before travel                      |
| OIM / Site Manager          | Read-only oversight; can trigger Minimum Manning Mode                                     |
| Personnel / Crew Member     | Views own profile, schedule, room assignment, and certification status                    |
| System Admin                | Full platform configuration — vessel setup, user management, role definitions             |

## Role Permissions Matrix
| Feature                 | Activity Owner | Planner | Safety Admin | OIM  | Personnel | Sys Admin |
|-------------------------|----------------|---------|--------------|------|-----------|-----------|
| Create Activity         | ✅             | ✅      | ❌           | ❌   | ❌        | ✅        |
| Approve Activity        | ❌             | ✅      | ❌           | ❌   | ❌        | ❌        |
| Manage Gantt            | View Only      | Full    | ❌           | View Only | ❌   | ❌        |
| Manage Certificates     | ❌             | ❌      | ✅           | ❌   | Own Only  | ✅        |
| Approve Travel          | ❌             | ✅      | ❌           | ❌   | ❌        | ❌        |
| View Cost Dashboard     | ❌             | ✅      | ❌           | ✅   | ❌        | ✅        |
| Trigger Min Manning     | ❌             | ❌      | ❌           | ✅   | ❌        | ✅        |
| Manage Roles            | ❌             | ❌      | ❌           | ❌   | ❌        | ✅        |
| Configure Vessel        | ❌             | ❌      | ❌           | ❌   | ❌        | ✅        |
| View Own Profile        | ✅             | ✅      | ✅           | ✅   | ✅        | ✅        |

## Development Phases
| Phase | Description | Status |
|-------|-------------|--------|
| 1     | Foundation (Auth, RBAC, DB, Redis) | ✅ Complete |
| 2     | Personnel & Compliance | ✅ Complete |
| 3     | Vessel & Room Management | ✅ Complete |
| 4     | Roles & Rotation Scheduling | ✅ Complete |
| 5     | Activity Management (Gantt, Approvals) | ✅ Complete |
| 6     | Travel & Mobilization | ✅ Complete |
| 7     | Minimum Manning Mode | ✅ Complete |
| 8     | Dashboard & Reporting (PDF/CSV) | ✅ Complete |

## API Endpoints
All routes are prefixed with `/api/v1`

### Auth — `/api/v1/auth`
| Method | Endpoint           | Description                         | Access        |
|--------|--------------------|-------------------------------------|---------------|
| POST   | `/register`        | Register a new organization + admin | Public        |
| POST   | `/login`           | Login, returns access + refresh token | Public      |
| POST   | `/refresh`         | Refresh access token                | Public        |
| POST   | `/logout`          | Invalidate refresh token            | Authenticated |
| GET    | `/me`              | Get current user profile            | Authenticated |
| PATCH  | `/me`              | Update current user profile         | Authenticated |
| POST   | `/change-password` | Change password                     | Authenticated |

### Users — `/api/v1/users`
| Method | Endpoint         | Description              | Access     |
|--------|------------------|--------------------------|------------|
| POST   | `/`              | Create user in org       | Sys Admin  |
| GET    | `/`              | List all users           | Sys Admin  |
| GET    | `/:id`           | Get user by ID           | Sys Admin  |
| PATCH  | `/:id`           | Update user              | Sys Admin  |
| DELETE | `/:id`           | Deactivate user          | Sys Admin  |
| PATCH  | `/:id/role`      | Assign or change role    | Sys Admin  |

### Positions (Offshore Roles) — `/api/v1/positions`
| Method | Endpoint | Description         | Access        |
|--------|----------|---------------------|---------------|
| POST   | `/`      | Create offshore role | Authenticated |
| GET    | `/`      | List all roles       | Authenticated |

### Personnel — `/api/v1/personnel`
| Method | Endpoint                         | Description                | Access                    |
|--------|----------------------------------|----------------------------|---------------------------|
| POST   | `/`                              | Create personnel record    | Authenticated             |
| GET    | `/`                              | List all personnel         | Authenticated             |
| PATCH  | `/:id`                           | Update personnel info      | Authenticated             |
| DELETE | `/:id`                           | Remove personnel           | Authenticated             |
| GET    | `/:id/compliance`                | Get compliance status      | Authenticated             |
| POST   | `/:id/certificates`              | Add certificate            | Authenticated             |
| GET    | `/:id/certificates`              | List certificates          | Authenticated             |
| PATCH  | `/:id/certificates/:certId`      | Update certificate record  | Authenticated             |
| DELETE | `/:id/certificates/:certId`      | Delete certificate         | Authenticated             |

### Notifications — `/api/v1/notifications`
| Method | Endpoint       | Description                | Access        |
|--------|----------------|----------------------------|---------------|
| GET    | `/`            | Get user's notifications   | Authenticated |
| PATCH  | `/:id/read`    | Mark notification as read  | Authenticated |

### Vessels — `/api/v1/vessels`
| Method | Endpoint            | Description                      | Access                         |
|--------|---------------------|----------------------------------|--------------------------------|
| POST   | `/`                 | Create vessel or installation    | Authenticated                  |
| GET    | `/`                 | List all vessels                 | Authenticated                  |
| GET    | `/:id`              | Get vessel details               | Authenticated                  |
| PATCH  | `/:id`              | Update vessel info               | Authenticated                  |
| DELETE | `/:id`              | Remove vessel                    | Authenticated                  |
| GET    | `/:id/pob`          | Real-time POB count (Redis)      | Authenticated                  |
| GET    | `/:id/manifest`     | Full POB manifest snapshot       | Planner, OIM, Sys Admin        |
| GET    | `/:id/manning`      | Current vessel manning           | Planner, OIM, Sys Admin        |

### Rooms
**Under Vessel context — `/api/v1/vessels/:id/rooms`**
| Method | Endpoint   | Description                | Access        |
|--------|------------|----------------------------|---------------|
| POST   | `/`        | Create a room              | Sys Admin     |
| GET    | `/`        | List all rooms on vessel   | Authenticated |
| POST   | `/assign`  | Assign personnel to room   | Planner       |

**Direct access — `/api/v1/rooms`**
| Method | Endpoint          | Description                | Access        |
|--------|-------------------|----------------------------|---------------|
| GET    | `/:id`            | Get room details           | Authenticated |
| PATCH  | `/:id`            | Update room info           | Sys Admin     |
| DELETE | `/:id`            | Delete room                | Sys Admin     |
| GET    | `/:id/occupants`  | Get current room occupants | Planner, Sys Admin |

### Rotation & Scheduling — `/api/v1`
| Method | Endpoint                                          | Description                         | Access     |
|--------|---------------------------------------------------|-------------------------------------|------------|
| POST   | `/rotation-schedules`                             | Create rotation schedule            | Sys Admin  |
| GET    | `/rotation-schedules?role_id=&vessel_id=`         | Get rotation schedules              | Planner    |
| POST   | `/role-assignments/assign`                        | Assign personnel to role            | Planner    |
| POST   | `/role-assignments/:id/end`                       | End role assignment                 | Planner    |
| POST   | `/back-to-back-pairs`                             | Create back-to-back pair            | Sys Admin  |
| GET    | `/back-to-back-pairs?role_id=&vessel_id=`         | List back-to-back pairs             | Planner    |
| POST   | `/rotation/calculate`                             | Calculate next rotation dates       | Planner    |

### Activities — `/api/v1/activities`
| Method | Endpoint                        | Description                         | Access                    |
|--------|---------------------------------|-------------------------------------|---------------------------|
| POST   | `/`                             | Create activity                     | Activity Owner, Sys Admin |
| GET    | `/`                             | List activities by vessel           | Authenticated             |
| GET    | `/:id`                          | Get activity details                | Authenticated             |
| GET    | `/gantt`                        | Gantt chart data                    | Planner, OIM, Sys Admin   |
| GET    | `/conflicts`                    | Check scheduling conflicts          | Planner                   |
| GET    | `/queue`                        | Pending approval queue              | Planner                   |
| POST   | `/submit`                       | Submit for approval                 | Activity Owner            |
| POST   | `/approve`                      | Approve activity                    | Planner                   |
| POST   | `/reject`                       | Reject activity                     | Planner                   |
| GET    | `/:id/requirements`             | Get role requirements               | Planner                   |
| GET    | `/:id/assignments`              | Get personnel assignments           | Planner                   |
| POST   | `/assign`                       | Assign personnel to activity        | Planner                   |
| DELETE | `/:id`                          | Delete activity (draft only)        | Activity Owner, Sys Admin |

### Travel & Mobilization — `/api/v1`
| Method | Endpoint                              | Description                         | Access        |
|--------|---------------------------------------|-------------------------------------|---------------|
| POST   | `/transports`                         | Create transport                    | Sys Admin     |
| GET    | `/transports`                         | List all transports                 | Planner, Sys Admin |
| GET    | `/transports/:id`                     | Get transport details               | Planner, Sys Admin |
| PATCH  | `/transports/:id`                     | Update transport                    | Sys Admin     |
| DELETE | `/transports/:id`                     | Delete transport                    | Sys Admin     |
| POST   | `/travel/schedules`                   | Create travel schedule              | Planner       |
| GET    | `/travel/schedules`                   | List travel schedules               | Planner, Personnel |
| GET    | `/travel/schedules/:id`               | Get schedule details                | Planner       |
| GET    | `/travel/schedules/:id/assignments`   | Get assigned personnel              | Planner       |
| POST   | `/travel/match-activities`            | Auto-match activities to transport  | Planner       |
| POST   | `/travel/assign`                      | Assign personnel to trip            | Planner       |
| GET    | `/travel/alerts`                      | Low utilization alerts              | Planner       |
| POST   | `/travel/consolidate`                 | Trip consolidation suggestions      | Planner       |
| GET    | `/travel/my-travels`                  | View own travel schedule            | Personnel     |

Transport payloads can now include `origin_vessel_id` and `destination_vessel_id` so a route can be shown between two vessels.
`GET /travel/schedules` also supports `transport_id`, `vessel_id`, `origin_vessel_id`, `destination_vessel_id`, `status`, `upcoming_only`, and `limit` query parameters.

### Minimum Manning — `/api/v1/minimum-manning`
| Method | Endpoint       | Description                         | Access                |
|--------|----------------|-------------------------------------|-----------------------|
| POST   | `/activate`    | Activate minimum manning mode       | OIM, Sys Admin        |
| POST   | `/deactivate`  | Deactivate minimum manning mode     | OIM, Sys Admin        |
| GET    | `/active`      | Get active event                    | Planner, OIM, Sys Admin |
| GET    | `/history`     | Get event history                   | Planner, OIM, Sys Admin |

### Dashboard — `/api/v1/dashboard`
| Method | Endpoint | Description                         | Access        |
|--------|----------|-------------------------------------|---------------|
| GET    | `/`      | Role-filtered dashboard data        | Authenticated |

### Reports — `/api/v1/reports`
| Method | Endpoint          | Description                         | Access                |
|--------|-------------------|-------------------------------------|-----------------------|
| GET    | `/daily`          | Daily POB report (JSON)             | Planner, OIM, Sys Admin |
| GET    | `/historical`     | Historical POB data                 | Planner, OIM, Sys Admin |
| GET    | `/export/pdf`     | Export report as PDF                | Planner, Sys Admin     |
| GET    | `/export/csv`     | Export report as CSV                | Planner, Sys Admin     |
```bash
## Project Structure (Clean Architecture)
pob-management/
├── cmd/
│ └── api/
│ └── main.go # Entry point
├── config/
│ ├── config.go # Env config loader
│ └── role.config.go # RBAC roles & permissions matrix
├── internal/
│ ├── domain/ # Entities (all domain models)
│ │ ├── activity.go
│ │ ├── activity_assignment.go
│ │ ├── activity_requirement.go
│ │ ├── back_to_back_pair.go
│ │ ├── role_assignment.go
│ │ ├── rotation_schedule.go
│ │ ├── transport.go
│ │ ├── travel_schedule.go
│ │ ├── travel_assignment.go
│ │ ├── minimum_manning.go
│ │ └── ... (user, vessel, room, personnel, certificate, etc.)
│ ├── delivery/http/
│ │ ├── controllers/ # HTTP handlers
│ │ │ ├── activity.controller.go
│ │ │ ├── rotation.controller.go
│ │ │ ├── travel.controller.go
│ │ │ ├── minimum_manning.controller.go
│ │ │ ├── dashboard.controller.go
│ │ │ ├── report.controller.go
│ │ │ └── ... (auth, user, vessel, room, personnel, etc.)
│ │ ├── middleware/ # JWT, RBAC
│ │ │ ├── auth.go
│ │ │ └── rbac.go
│ │ └── routes/
│ │ └── routes.go # Route registration
│ ├── repository/ # Database queries
│ │ ├── activity.repository.go
│ │ ├── activity_requirement.repository.go
│ │ ├── activity_assignment.repository.go
│ │ ├── rotation_schedule.repository.go
│ │ ├── role_assignment.repository.go
│ │ ├── back_to_back_pair.repository.go
│ │ ├── transport.repository.go
│ │ ├── travel_schedule.repository.go
│ │ ├── travel_assignment.repository.go
│ │ ├── minimum_manning.repository.go
│ │ └── ... (user, vessel, room, personnel, certificate, etc.)
│ └── service/ # Business logic
│ ├── activity.service.go
│ ├── rotation.service.go
│ ├── travel.service.go
│ ├── minimum_manning.service.go
│ ├── dashboard.service.go
│ ├── report.service.go
│ └── ... (auth, user, vessel, room, personnel, compliance, etc.)
├── pkg/
│ ├── database/ # MongoDB + Redis init
│ │ ├── mongo.go
│ │ └── redis.go
│ ├── logger/ # Logging utilities
│ └── response/ # Standardized API responses
├── docs/
│ ├── POB_Management_API.postman_collection.json # Postman API docs
│ └── model-relationships.md # Entity relationship docs
├── scripts/
│ └── test_all_api.sh # Full API test script
├── .env.example
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```
## Non-Functional Requirements
- Dashboard loads within **3 seconds** under normal network
- API response times under **500ms** for standard reads
- Gantt chart renders up to **50 simultaneous activities**
- Supports up to **500 concurrent users**
- **99.5% uptime** target during business hours
- All data in transit encrypted via **HTTPS/TLS**
- Passwords hashed with **bcrypt**
- All user actions logged for audit
- File uploads validated for type and scanned
- Dates stored in UTC, displayed in user's local timezone
- Supports up to **10 vessels per organization**

## System Constraints
- Backend API is the primary deliverable
- Certificate documents accepted in **PDF format only**
- AI optimization is out of scope for the initial build
- No third-party ERP/HR integration in Phase 1 (API designed to support it later)
- All cost values stored and displayed in **USD** unless configured otherwise

## Quick Start
1. Clone the repository
2. Copy `.env.example` to `.env` and configure your MongoDB Atlas and Redis Cloud URLs
3. Run `go mod download`
4. Start the server: `go run cmd/api/main.go`
5. Import the Postman collection from `docs/` and set `base_url` to `http://localhost:8081`

## Testing
Run the full API test script:
```bash
chmod +x scripts/test_all_api.sh
./scripts/test_all_api.sh
Version 2.0 — April 2026 | Testudo Nigeria Limited
All 8 phases complete — production-ready backend.
