# Model Relationships

This document captures the core backend data model for the MongoDB-based POB system.

## Design Rules

- Use MongoDB references with `ObjectID` between collections instead of embedding large, fast-growing arrays.
- Keep authentication data in `users`.
- Keep offshore operational data in `personnel`.
- Link `personnel.user_id` only when a crew member or staff member should log into the system.
- Use assignment collections for time-bound relationships instead of mutating the parent documents directly.

## Core Collections

### `users`

- Purpose: authentication, password management, system authorization.
- Key relation: optional one-to-one link from `personnel.user_id` to `users._id`.

### `personnel`

- Purpose: offshore worker profile and operational status.
- Key relations:
- `user_id -> users._id` (optional)
- `current_vessel_id -> vessels._id` (optional)

### `certificates`

- Purpose: compliance documents and expiry tracking.
- Key relations:
- `personnel_id -> personnel._id`
- `uploaded_by_user_id -> users._id` (optional)

### `vessels`

- Purpose: vessel or installation configuration and POB limits.
- Key relations:
- parent of `rooms`
- parent of `role_assignments`
- parent of `travel_schedules`
- parent of `minimum_manning_events`

### `rooms`

- Purpose: physical accommodation units inside a vessel.
- Key relations:
- `vessel_id -> vessels._id`

### `offshore_roles`

- Purpose: role definitions such as core and flexible positions.
- Key relations:
- optional `vessel_id -> vessels._id`
- parent of `rotation_schedules`
- parent of `role_assignments`
- parent of `activity_requirements`

### `rotation_schedules`

- Purpose: rotation pattern for an offshore role on a vessel.
- Key relations:
- `offshore_role_id -> offshore_roles._id`
- `vessel_id -> vessels._id`

### `role_assignments`

- Purpose: assigns a person to an offshore role for a time range.
- Key relations:
- `offshore_role_id -> offshore_roles._id`
- `personnel_id -> personnel._id`
- `vessel_id -> vessels._id`
- `rotation_schedule_id -> rotation_schedules._id` (optional)
- `room_id -> rooms._id` (optional)
- `assigned_by_user_id -> users._id` (optional)

### `back_to_back_pairs`

- Purpose: keeps the primary and relief personnel tied to the same operational role.
- Key relations:
- `offshore_role_id -> offshore_roles._id`
- `vessel_id -> vessels._id`
- `primary_personnel_id -> personnel._id`
- `relief_personnel_id -> personnel._id`
- `room_id -> rooms._id` (optional)

### `room_assignments`

- Purpose: tracks who occupies which room and for what period.
- Key relations:
- `vessel_id -> vessels._id`
- `room_id -> rooms._id`
- `personnel_id -> personnel._id`
- `offshore_role_id -> offshore_roles._id` (optional)
- `role_assignment_id -> role_assignments._id` (optional)
- `activity_id -> activities._id` (optional)

### `activities`

- Purpose: activity planning and approval workflow.
- Key relations:
- `vessel_id -> vessels._id`
- `created_by_user_id -> users._id`
- `reviewed_by_user_id -> users._id` (optional)

### `activity_requirements`

- Purpose: planned manpower demand per activity.
- Key relations:
- `activity_id -> activities._id`
- `offshore_role_id -> offshore_roles._id`

### `activity_assignments`

- Purpose: actual personnel staffing for an activity.
- Key relations:
- `activity_id -> activities._id`
- `personnel_id -> personnel._id`
- `offshore_role_id -> offshore_roles._id`
- `role_assignment_id -> role_assignments._id` (optional)
- `travel_schedule_id -> travel_schedules._id` (optional)
- `room_assignment_id -> room_assignments._id` (optional)
- `assigned_by_user_id -> users._id` (optional)

### `transports`

- Purpose: reusable transport configuration such as helicopter or boat.

### `travel_schedules`

- Purpose: concrete travel movement for a date and route.
- Key relations:
- `transport_id -> transports._id`
- `vessel_id -> vessels._id` (optional)
- `activity_id -> activities._id` (optional)

### `travel_assignments`

- Purpose: assigns personnel to a travel schedule.
- Key relations:
- `travel_schedule_id -> travel_schedules._id`
- `personnel_id -> personnel._id`
- `activity_id -> activities._id` (optional)
- `approved_by_user_id -> users._id` (optional)

### `minimum_manning_events`

- Purpose: tracks activation and deactivation of minimum manning mode.
- Key relations:
- `vessel_id -> vessels._id`
- `activated_by_user_id -> users._id`
- `deactivated_by_user_id -> users._id` (optional)
- `affected_activity_ids -> activities._id[]`
- `affected_personnel_ids -> personnel._id[]`

### `notifications`

- Purpose: email and in-app communication log.
- Key relations:
- `user_id -> users._id` (optional)
- `personnel_id -> personnel._id` (optional)
- `related_entity_id -> any core entity _id` (optional)

### `audit_logs`

- Purpose: immutable audit trail of user actions.
- Key relations:
- `actor_user_id -> users._id` (optional)
- `entity_id -> any core entity _id` (optional)

## High-Value Relationship Notes

### User vs Personnel

- `users` is for login and RBAC.
- `personnel` is for offshore profile data.
- Not every user must be personnel.
- Not every personnel record must have a user account.

### Role Assignment vs Room Assignment

- `role_assignments` answers: who is assigned to this operational role?
- `room_assignments` answers: who is occupying this room during this time window?
- This separation makes back-to-back swaps easier because a room can remain tied to the operational role while the assigned person changes over time.

### Activity Requirement vs Activity Assignment

- `activity_requirements` stores the demand plan.
- `activity_assignments` stores the actual staffed people.
- This separation helps the planner compare planned vs fulfilled staffing.

### Travel Schedule vs Travel Assignment

- `travel_schedules` is the trip itself.
- `travel_assignments` is the seat allocation per person.

## Recommended Starter Indexes

- `users.email` unique
- `personnel.employee_number` unique
- `personnel.user_id` sparse unique
- `certificates.personnel_id`
- `certificates.expires_at`
- `rooms.vessel_id`
- `offshore_roles.code` unique
- `role_assignments.personnel_id`
- `role_assignments.offshore_role_id`
- `activities.vessel_id`
- `activities.status`
- `activity_assignments.activity_id`
- `travel_schedules.departure_at`
- `travel_assignments.travel_schedule_id`
- `minimum_manning_events.vessel_id`
- `audit_logs.created_at`
