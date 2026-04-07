package config

// Role type
type Role string

// Permission type
type Permission string

// Roles
const (
	RoleSystemAdmin   Role = "system_admin"
	RolePlanner       Role = "planner"
	RoleActivityOwner Role = "activity_owner"
	RoleSafetyAdmin   Role = "safety_admin"
	RoleOIM           Role = "oim"
	RolePersonnel     Role = "personnel"
)

// Permissions
const (
	// Auth
	PermViewOwnProfile   Permission = "view_own_profile"
	PermUpdateOwnProfile Permission = "update_own_profile"

	// Users
	PermListUsers      Permission = "list_users"
	PermGetUser        Permission = "get_user"
	PermUpdateUser     Permission = "update_user"
	PermDeactivateUser Permission = "deactivate_user"
	PermAssignUserRole Permission = "assign_user_role"

	// Vessels
	PermCreateVessel         Permission = "create_vessel"
	PermListVessels          Permission = "list_vessels"
	PermGetVessel            Permission = "get_vessel"
	PermUpdateVessel         Permission = "update_vessel"
	PermDeleteVessel         Permission = "delete_vessel"
	PermViewPOBCount         Permission = "view_pob_count"
	PermViewPOBCapacity      Permission = "view_pob_capacity"
	PermSetPOBCap            Permission = "set_pob_cap"
	PermViewManifest         Permission = "view_manifest"
	PermActivateMinManning   Permission = "activate_min_manning"
	PermDeactivateMinManning Permission = "deactivate_min_manning"

	// Rooms
	PermCreateRoom        Permission = "create_room"
	PermListRooms         Permission = "list_rooms"
	PermGetRoom           Permission = "get_room"
	PermUpdateRoom        Permission = "update_room"
	PermDeleteRoom        Permission = "delete_room"
	PermViewRoomOccupants Permission = "view_room_occupants"

	// Offshore Roles
	PermCreateOffshoreRole     Permission = "create_offshore_role"
	PermListOffshoreRoles      Permission = "list_offshore_roles"
	PermGetOffshoreRole        Permission = "get_offshore_role"
	PermUpdateOffshoreRole     Permission = "update_offshore_role"
	PermDeactivateOffshoreRole Permission = "deactivate_offshore_role"
	PermAssignPersonnelToRole  Permission = "assign_personnel_to_role"
	PermListPersonnelInRole    Permission = "list_personnel_in_role"
	PermSetBackToBack          Permission = "set_back_to_back"

	// Personnel
	PermCreatePersonnel         Permission = "create_personnel"
	PermListPersonnel           Permission = "list_personnel"
	PermGetPersonnel            Permission = "get_personnel"
	PermUpdatePersonnel         Permission = "update_personnel"
	PermDeletePersonnel         Permission = "delete_personnel"
	PermViewOwnPersonnelProfile Permission = "view_own_personnel_profile"

	// Certificates
	PermViewCertificates    Permission = "view_certificates"
	PermViewOwnCertificates Permission = "view_own_certificates"
	PermUploadCertificate   Permission = "upload_certificate"
	PermUpdateCertificate   Permission = "update_certificate"
	PermDeleteCertificate   Permission = "delete_certificate"

	// Compliance
	PermViewCompliance              Permission = "view_compliance"
	PermViewExpiringCertificates    Permission = "view_expiring_certificates"
	PermViewExpiredCertificates     Permission = "view_expired_certificates"
	PermValidatePersonnelCompliance Permission = "validate_personnel_compliance"
	PermViewActivityCompliance      Permission = "view_activity_compliance"

	// Activities
	PermCreateActivity     Permission = "create_activity"
	PermListActivities     Permission = "list_activities"
	PermGetActivity        Permission = "get_activity"
	PermUpdateActivity     Permission = "update_activity"
	PermDeleteActivity     Permission = "delete_activity"
	PermSubmitActivity     Permission = "submit_activity"
	PermApproveActivity    Permission = "approve_activity"
	PermRejectActivity     Permission = "reject_activity"
	PermRescheduleActivity Permission = "reschedule_activity"
	PermViewGantt          Permission = "view_gantt"
	PermManageGantt        Permission = "manage_gantt"
	PermViewConflicts      Permission = "view_conflicts"
	PermViewActivityQueue  Permission = "view_activity_queue"

	// Travel
	PermCreateTransport     Permission = "create_transport"
	PermListTransport       Permission = "list_transport"
	PermUpdateTransport     Permission = "update_transport"
	PermDeleteTransport     Permission = "delete_transport"
	PermViewTravelSchedule  Permission = "view_travel_schedule"
	PermAssignTravel        Permission = "assign_travel"
	PermViewTripUtilization Permission = "view_trip_utilization"
	PermViewTravelAlerts    Permission = "view_travel_alerts"
	PermViewOwnTravel       Permission = "view_own_travel"

	// Cost
	PermViewTripCosts       Permission = "view_trip_costs"
	PermViewIdleContractors Permission = "view_idle_contractors"
	PermViewCostRiskAlerts  Permission = "view_cost_risk_alerts"
	PermCompareCosts        Permission = "compare_costs"

	// Optimization
	PermGeneratePlan             Permission = "generate_plan"
	PermViewPlan                 Permission = "view_plan"
	PermApprovePlan              Permission = "approve_plan"
	PermOverridePlan             Permission = "override_plan"
	PermViewOvershoot            Permission = "view_overshoot"
	PermViewOvershootSuggestions Permission = "view_overshoot_suggestions"

	// Dashboard
	PermViewDashboard               Permission = "view_dashboard"
	PermViewPOBToday                Permission = "view_pob_today"
	PermViewUpcomingActivities      Permission = "view_upcoming_activities"
	PermViewExpiringCertsDashboard  Permission = "view_expiring_certs_dashboard"
	PermViewUpcomingTravelDashboard Permission = "view_upcoming_travel_dashboard"
	PermViewCostRisksDashboard      Permission = "view_cost_risks_dashboard"

	// Reports
	PermViewDailyReport      Permission = "view_daily_report"
	PermViewHistoricalReport Permission = "view_historical_report"
	PermExportPDFReport      Permission = "export_pdf_report"
	PermExportCSVReport      Permission = "export_csv_report"
)

// RolePermissions maps each role to its list of permissions
var RolePermissions = map[Role][]Permission{
	RoleSystemAdmin: {
		// Auth
		PermViewOwnProfile,
		PermUpdateOwnProfile,
		// Users
		PermListUsers,
		PermGetUser,
		PermUpdateUser,
		PermDeactivateUser,
		PermAssignUserRole,
		// Vessels
		PermCreateVessel,
		PermListVessels,
		PermGetVessel,
		PermUpdateVessel,
		PermDeleteVessel,
		PermViewPOBCount,
		PermViewPOBCapacity,
		PermSetPOBCap,
		PermViewManifest,
		PermActivateMinManning,
		PermDeactivateMinManning,
		// Rooms
		PermCreateRoom,
		PermListRooms,
		PermGetRoom,
		PermUpdateRoom,
		PermDeleteRoom,
		PermViewRoomOccupants,
		// Offshore Roles
		PermCreateOffshoreRole,
		PermListOffshoreRoles,
		PermGetOffshoreRole,
		PermUpdateOffshoreRole,
		PermDeactivateOffshoreRole,
		PermAssignPersonnelToRole,
		PermListPersonnelInRole,
		PermSetBackToBack,
		// Personnel
		PermCreatePersonnel,
		PermListPersonnel,
		PermGetPersonnel,
		PermUpdatePersonnel,
		PermDeletePersonnel,
		// Certificates
		PermViewCertificates,
		PermUploadCertificate,
		PermUpdateCertificate,
		PermDeleteCertificate,
		// Compliance
		PermViewCompliance,
		PermViewExpiringCertificates,
		PermViewExpiredCertificates,
		PermValidatePersonnelCompliance,
		PermViewActivityCompliance,
		// Activities
		PermCreateActivity,
		PermListActivities,
		PermGetActivity,
		PermUpdateActivity,
		PermDeleteActivity,
		PermViewGantt,
		PermManageGantt,
		PermViewConflicts,
		PermViewActivityQueue,
		// Travel
		PermCreateTransport,
		PermListTransport,
		PermUpdateTransport,
		PermDeleteTransport,
		PermViewTravelSchedule,
		PermAssignTravel,
		PermViewTripUtilization,
		PermViewTravelAlerts,
		// Cost
		PermViewTripCosts,
		PermViewIdleContractors,
		PermViewCostRiskAlerts,
		PermCompareCosts,
		// Optimization
		PermGeneratePlan,
		PermViewPlan,
		PermApprovePlan,
		PermOverridePlan,
		PermViewOvershoot,
		PermViewOvershootSuggestions,
		// Dashboard
		PermViewDashboard,
		PermViewPOBToday,
		PermViewUpcomingActivities,
		PermViewExpiringCertsDashboard,
		PermViewUpcomingTravelDashboard,
		PermViewCostRisksDashboard,
		// Reports
		PermViewDailyReport,
		PermViewHistoricalReport,
		PermExportPDFReport,
		PermExportCSVReport,
	},

	RolePlanner: {
		// Auth
		PermViewOwnProfile,
		PermUpdateOwnProfile,
		// Vessels
		PermListVessels,
		PermGetVessel,
		PermViewPOBCount,
		PermViewPOBCapacity,
		PermViewManifest,
		// Rooms
		PermListRooms,
		PermGetRoom,
		PermViewRoomOccupants,
		// Offshore Roles
		PermListOffshoreRoles,
		PermGetOffshoreRole,
		PermListPersonnelInRole,
		// Personnel
		PermListPersonnel,
		PermGetPersonnel,
		// Certificates
		PermViewCertificates,
		// Compliance
		PermViewCompliance,
		PermViewExpiringCertificates,
		PermValidatePersonnelCompliance,
		PermViewActivityCompliance,
		// Activities
		PermCreateActivity,
		PermListActivities,
		PermGetActivity,
		PermUpdateActivity,
		PermDeleteActivity,
		PermApproveActivity,
		PermRejectActivity,
		PermRescheduleActivity,
		PermViewGantt,
		PermManageGantt,
		PermViewConflicts,
		PermViewActivityQueue,
		// Travel
		PermListTransport,
		PermViewTravelSchedule,
		PermAssignTravel,
		PermViewTripUtilization,
		PermViewTravelAlerts,
		// Cost
		PermViewTripCosts,
		PermViewIdleContractors,
		PermViewCostRiskAlerts,
		PermCompareCosts,
		// Optimization
		PermGeneratePlan,
		PermViewPlan,
		PermApprovePlan,
		PermOverridePlan,
		PermViewOvershoot,
		PermViewOvershootSuggestions,
		// Dashboard
		PermViewDashboard,
		PermViewPOBToday,
		PermViewUpcomingActivities,
		PermViewExpiringCertsDashboard,
		PermViewUpcomingTravelDashboard,
		PermViewCostRisksDashboard,
		// Reports
		PermViewDailyReport,
		PermViewHistoricalReport,
		PermExportPDFReport,
		PermExportCSVReport,
	},

	RoleActivityOwner: {
		// Auth
		PermViewOwnProfile,
		PermUpdateOwnProfile,
		// Vessels
		PermListVessels,
		PermGetVessel,
		PermViewPOBCount,
		PermViewPOBCapacity,
		// Rooms
		PermListRooms,
		PermGetRoom,
		// Offshore Roles
		PermListOffshoreRoles,
		PermGetOffshoreRole,
		// Personnel
		PermGetPersonnel,
		// Compliance
		PermViewActivityCompliance,
		// Activities
		PermCreateActivity,
		PermListActivities,
		PermGetActivity,
		PermUpdateActivity,
		PermDeleteActivity,
		PermSubmitActivity,
		PermViewGantt,
		// Dashboard
		PermViewDashboard,
		PermViewPOBToday,
		PermViewUpcomingActivities,
	},

	RoleSafetyAdmin: {
		// Auth
		PermViewOwnProfile,
		PermUpdateOwnProfile,
		// Vessels
		PermListVessels,
		PermGetVessel,
		// Personnel
		PermListPersonnel,
		PermGetPersonnel,
		// Certificates
		PermViewCertificates,
		PermUploadCertificate,
		PermUpdateCertificate,
		PermDeleteCertificate,
		// Compliance
		PermViewCompliance,
		PermViewExpiringCertificates,
		PermViewExpiredCertificates,
		PermValidatePersonnelCompliance,
		PermViewActivityCompliance,
		// Activities
		PermListActivities,
		PermGetActivity,
		// Dashboard
		PermViewDashboard,
		PermViewPOBToday,
		PermViewExpiringCertsDashboard,
	},

	RoleOIM: {
		// Auth
		PermViewOwnProfile,
		PermUpdateOwnProfile,
		// Vessels
		PermListVessels,
		PermGetVessel,
		PermViewPOBCount,
		PermViewPOBCapacity,
		PermViewManifest,
		PermActivateMinManning,
		PermDeactivateMinManning,
		// Rooms
		PermListRooms,
		PermGetRoom,
		// Offshore Roles
		PermListOffshoreRoles,
		PermGetOffshoreRole,
		// Personnel
		PermListPersonnel,
		PermGetPersonnel,
		// Compliance
		PermViewCompliance,
		// Activities
		PermListActivities,
		PermGetActivity,
		PermViewGantt,
		// Cost
		PermViewTripCosts,
		PermViewIdleContractors,
		PermViewCostRiskAlerts,
		// Dashboard
		PermViewDashboard,
		PermViewPOBToday,
		PermViewUpcomingActivities,
		PermViewCostRisksDashboard,
		PermViewUpcomingTravelDashboard,
		// Reports
		PermViewDailyReport,
		PermViewHistoricalReport,
	},

	RolePersonnel: {
		// Auth
		PermViewOwnProfile,
		PermUpdateOwnProfile,
		// Vessels
		PermListVessels,
		PermGetVessel,
		// Own data
		PermViewOwnPersonnelProfile,
		PermViewOwnCertificates,
		PermViewOwnTravel,
		// Activities
		PermListActivities,
		PermGetActivity,
		// Travel
		PermViewTravelSchedule,
		// Dashboard
		PermViewDashboard,
		PermViewPOBToday,
		PermViewUpcomingActivities,
		PermViewUpcomingTravelDashboard,
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role Role, permission Permission) bool {
	permissions, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a role has at least one of the given permissions
func HasAnyPermission(role Role, permissions []Permission) bool {
	for _, p := range permissions {
		if HasPermission(role, p) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a role has all of the given permissions
func HasAllPermissions(role Role, permissions []Permission) bool {
	for _, p := range permissions {
		if !HasPermission(role, p) {
			return false
		}
	}
	return true
}

// GetPermissions returns all permissions for a given role
func GetPermissions(role Role) []Permission {
	return RolePermissions[role]
}
