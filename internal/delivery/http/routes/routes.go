package routes

import (
	"context"
	"log"

	"github.com/codingninja/pob-management/config"
	"github.com/codingninja/pob-management/internal/delivery/http/controllers"
	"github.com/codingninja/pob-management/internal/delivery/http/middleware"
	"github.com/codingninja/pob-management/internal/repository"
	"github.com/codingninja/pob-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Setup(r *gin.Engine, db *mongo.Database, rdb *redis.Client, cfg *config.Config) {

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userRepository := repository.NewUserRepository(db)
	organizationRepository := repository.NewOrganizationRepository(db)
	tokenManager := service.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTLMinutes, cfg.RefreshTokenTTLHours)
	authService := service.NewAuthService(userRepository, organizationRepository, tokenManager)
	userService := service.NewUserService(userRepository)
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	authMiddleware := middleware.NewAuthMiddleware(tokenManager)

	// Phase 2 Initialization
	certTypeRepo := repository.NewCertificateTypeRepository(db)
	roleRepo := repository.NewOffshoreRoleRepository(db)
	personnelRepo := repository.NewPersonnelRepository(db)
	certRepo := repository.NewCertificateRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	certTypeRepo.EnsureIndexes(context.Background())
	roleRepo.EnsureIndexes(context.Background())
	personnelRepo.EnsureIndexes(context.Background())
	certRepo.EnsureIndexes(context.Background())
	notifRepo.EnsureIndexes(context.Background())

	_ = service.NewCertificateTypeService(certTypeRepo)
	certSvc := service.NewCertificateService(certRepo, certTypeRepo)
	roleSvc := service.NewOffshoreRoleService(roleRepo)
	personnelSvc := service.NewPersonnelService(personnelRepo, roleRepo)
	compSvc := service.NewComplianceService(personnelRepo, roleRepo, certRepo, certTypeRepo)
	notifSvc := service.NewNotificationService(notifRepo)

	roleCtrl := controllers.NewOffshoreRoleController(roleSvc)
	personnelCtrl := controllers.NewPersonnelController(personnelSvc, compSvc)
	certCtrl := controllers.NewCertificateController(certSvc)
	notifCtrl := controllers.NewNotificationController(notifSvc)

	// Reminder Service (Background)
	reminderSvc := service.NewReminderService(personnelRepo, certRepo, notifSvc, userRepository)
	reminderSvc.Start(context.Background())

	if err := authService.Initialize(context.Background()); err != nil {
		log.Printf("failed to initialize auth indexes: %v", err)
	}

	api := r.Group("/api/v1")
	auth := api.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.Refresh)

		authenticated := auth.Group("")
		authenticated.Use(authMiddleware.RequireAuth())
		{
			authenticated.POST("/logout", authController.Logout)
			authenticated.GET("/me", authController.Me)
			authenticated.PATCH("/me", authController.UpdateMe)
			authenticated.POST("/change-password", authController.ChangePassword)
		}
	}

	users := api.Group("/users")
	users.Use(authMiddleware.RequireAuth(), middleware.RequirePermission(config.PermListUsers))
	{
		users.POST("", middleware.RequirePermission(config.PermAssignUserRole), userController.CreateUser)
		users.GET("", userController.ListUsers)
		users.GET("/", userController.ListUsers)
		users.GET("/:id", userController.GetUser)
		users.PATCH("/:id", middleware.RequirePermission(config.PermUpdateUser), userController.UpdateUser)
		users.DELETE("/:id", middleware.RequirePermission(config.PermDeactivateUser), userController.DeactivateUser)
		users.PATCH("/:id/role", middleware.RequirePermission(config.PermAssignUserRole), userController.UpdateRole)
	}

	// Phase 2 Routes (Authenticated)
	apiSecured := api.Group("")
	apiSecured.Use(authMiddleware.RequireAuth())

	roles := apiSecured.Group("/positions")
	{
		roles.POST("", roleCtrl.CreateRole)
		roles.GET("", roleCtrl.ListRoles)
		roles.GET("/:id", roleCtrl.GetRole)
		roles.PATCH("/:id", roleCtrl.UpdateRole)
		roles.POST("/:id/certificates", roleCtrl.AddRequiredCertificate)
		roles.DELETE("/:id/certificates/:certTypeId", roleCtrl.RemoveRequiredCertificate)
	}

	personnel := apiSecured.Group("/personnel")
	{
		personnel.POST("", personnelCtrl.CreatePersonnel)
		personnel.GET("", personnelCtrl.ListPersonnel)
		personnel.GET("/:id", personnelCtrl.ListPersonnel)
		personnel.PATCH("/:id", personnelCtrl.UpdatePersonnel)
		personnel.DELETE("/:id", personnelCtrl.DeletePersonnel)
		personnel.GET("/:id/compliance", personnelCtrl.CheckCompliance)

		// Certificate Routes
		personnel.POST("/:id/certificates", certCtrl.CreateCertificate)
		personnel.GET("/:id/certificates", certCtrl.ListCertificates)
		personnel.PATCH("/:id/certificates/:certId", certCtrl.UpdateCertificate)
		personnel.DELETE("/:id/certificates/:certId", certCtrl.DeleteCertificate)
	}

	notifications := apiSecured.Group("/notifications")
	{
		notifications.GET("", notifCtrl.GetMyNotifications)
		notifications.PATCH("/:id/read", notifCtrl.MarkAsRead)
	}

	// Phase 3 Initialization
	vesselRepo := repository.NewVesselRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	roomAssignRepo := repository.NewRoomAssignmentRepository(db)
	vesselEventRepo := repository.NewVesselEventRepository(db)

	vesselRepo.EnsureIndexes(context.Background())
	roomRepo.EnsureIndexes(context.Background())
	roomAssignRepo.EnsureIndexes(context.Background())
	vesselEventRepo.EnsureIndexes(context.Background())

	vesselSvc := service.NewVesselService(vesselRepo, roomAssignRepo, vesselEventRepo, rdb)
	roomSvc := service.NewRoomService(roomRepo, roomAssignRepo, vesselSvc, compSvc)

	vesselCtrl := controllers.NewVesselController(vesselSvc)
	roomCtrl := controllers.NewRoomController(roomSvc)

	// Phase 3 Routes
	vessels := apiSecured.Group("/vessels")
	{
		vessels.POST("", vesselCtrl.CreateVessel)
		vessels.GET("", vesselCtrl.ListVessels)
		vessels.GET("/:id", vesselCtrl.GetVessel)
		vessels.PATCH("/:id", vesselCtrl.UpdateVessel)
		vessels.DELETE("/:id", vesselCtrl.DeleteVessel)

		vessels.GET("/:id/pob", vesselCtrl.GetRealTimePOB)
		vessels.GET("/:id/manifest", vesselCtrl.GetManifest)
		vessels.GET("/default", vesselCtrl.GetDefaultVessel)
		vessels.PATCH("/:id/set-default", vesselCtrl.SetDefaultVessel)
		vessels.POST("/:id/timeline", vesselCtrl.AddVesselEvent)
		vessels.GET("/:id/timeline", vesselCtrl.GetVesselTimeline)

		// Room Routes within Vessel context
		vessels.POST("/:id/rooms", roomCtrl.CreateRoom)
		vessels.GET("/:id/rooms", roomCtrl.ListRooms)
		vessels.GET("/:id/rooms/by-deck", roomCtrl.GetRoomsByDeck)
		vessels.POST("/:id/rooms/assign", roomCtrl.AssignRoom)
	}

	rooms := apiSecured.Group("/rooms")
	{
		rooms.GET("/:id", roomCtrl.GetRoom)
		rooms.PATCH("/:id", roomCtrl.UpdateRoom)
		rooms.DELETE("/:id", roomCtrl.DeleteRoom)
		rooms.GET("/:id/occupants", roomCtrl.GetRoomOccupants)
	}

	// ==================== PHASE 4: ROLES & ROTATION SCHEDULING ====================

	// Phase 4 Initialization
	rotationScheduleRepo := repository.NewRotationScheduleRepository(db)
	roleAssignmentRepo := repository.NewRoleAssignmentRepository(db)
	backToBackPairRepo := repository.NewBackToBackPairRepository(db)

	rotationScheduleRepo.EnsureIndexes(context.Background())
	roleAssignmentRepo.EnsureIndexes(context.Background())
	backToBackPairRepo.EnsureIndexes(context.Background())

	rotationSvc := service.NewRotationService(
		rotationScheduleRepo,
		roleAssignmentRepo,
		backToBackPairRepo,
		roleRepo,
		personnelRepo,
		roomAssignRepo,
	)

	rotationCtrl := controllers.NewRotationController(rotationSvc)

	// Rotation Schedule Routes
	rotationSchedules := apiSecured.Group("/rotation-schedules")
	{
		rotationSchedules.POST("", rotationCtrl.CreateSchedule)
		rotationSchedules.GET("", rotationCtrl.GetSchedules)
	}

	// Role Assignment Routes
	roleAssignments := apiSecured.Group("/role-assignments")
	{
		roleAssignments.POST("/assign", rotationCtrl.AssignRole)
		roleAssignments.POST("/:id/end", rotationCtrl.EndAssignment)
		roleAssignments.POST("/:id/handover", rotationCtrl.TriggerShiftHandover)
	}

	// Vessel Manning (add to existing vessels group)
	vessels.GET("/:id/manning", rotationCtrl.GetVesselManning)
	vessels.GET("/:id/active-assignments", rotationCtrl.GetActiveAssignments)

	// Back-to-Back Pairs
	backToBack := apiSecured.Group("/back-to-back-pairs")
	{
		backToBack.POST("", rotationCtrl.CreateBackToBackPair)
		backToBack.GET("", rotationCtrl.GetBackToBackPairs)
	}

	// Rotation Calculation
	apiSecured.POST("/rotation/calculate", rotationCtrl.CalculateNextRotation)

	// ==================== PHASE 5: ACTIVITY MANAGEMENT ====================

	// Phase 5 Initialization
	activityRepo := repository.NewActivityRepository(db)
	requirementRepo := repository.NewActivityRequirementRepository(db)
	assignmentRepo := repository.NewActivityAssignmentRepository(db)

	activityRepo.EnsureIndexes(context.Background())
	requirementRepo.EnsureIndexes(context.Background())
	assignmentRepo.EnsureIndexes(context.Background())

	activitySvc := service.NewActivityService(
		activityRepo,
		requirementRepo,
		assignmentRepo,
		roleRepo,
		personnelRepo,
		roleAssignmentRepo,
	)

	activityCtrl := controllers.NewActivityController(activitySvc)

	// Activity Routes
	activities := apiSecured.Group("/activities")
	{
		activities.POST("", activityCtrl.CreateActivity)
		activities.GET("", activityCtrl.ListActivities)
		activities.GET("/:id", activityCtrl.GetActivity)
		activities.GET("/gantt", activityCtrl.GetGanttData)
		activities.GET("/conflicts", activityCtrl.CheckConflicts)
		activities.GET("/queue", activityCtrl.GetPendingQueue)
		activities.POST("/submit", activityCtrl.SubmitForApproval)
		activities.POST("/approve", activityCtrl.ApproveActivity)
		activities.POST("/reject", activityCtrl.RejectActivity)
		activities.GET("/:id/requirements", activityCtrl.GetRequirements)
		activities.GET("/:id/assignments", activityCtrl.GetAssignments)
		activities.POST("/assign", activityCtrl.AssignPersonnel)
		activities.DELETE("/:id", activityCtrl.DeleteActivity)
	}

	// ==================== PHASE 6: TRAVEL & MOBILIZATION ====================

	// Phase 6 Initialization
	transportRepo := repository.NewTransportRepository(db)
	travelScheduleRepo := repository.NewTravelScheduleRepository(db)
	travelAssignmentRepo := repository.NewTravelAssignmentRepository(db)

	transportRepo.EnsureIndexes(context.Background())
	travelScheduleRepo.EnsureIndexes(context.Background())
	travelAssignmentRepo.EnsureIndexes(context.Background())

	travelSvc := service.NewTravelService(
		transportRepo,
		travelScheduleRepo,
		travelAssignmentRepo,
		activityRepo,
		personnelRepo,
		vesselRepo,
		compSvc,
	)

	travelCtrl := controllers.NewTravelController(travelSvc)

	// Transport Routes
	transports := apiSecured.Group("/transports")
	{
		transports.POST("", travelCtrl.CreateTransport)
		transports.GET("", travelCtrl.ListTransports)
		transports.GET("/:id", travelCtrl.GetTransport)
		transports.PATCH("/:id", travelCtrl.UpdateTransport)
		transports.DELETE("/:id", travelCtrl.DeleteTransport)
	}

	// Travel Schedule Routes
	travel := apiSecured.Group("/travel")
	{
		travel.POST("/schedules", travelCtrl.CreateTravelSchedule)
		travel.GET("/schedules", travelCtrl.ListUpcomingSchedules)
		travel.GET("/schedules/:id", travelCtrl.GetTravelSchedule)
		travel.GET("/schedules/:id/assignments", travelCtrl.GetTripAssignments)
		travel.POST("/match-activities", travelCtrl.MatchActivities)
		travel.POST("/assign", travelCtrl.AssignPersonnelToTrip)
		travel.GET("/alerts", travelCtrl.GetUtilizationAlerts)
		travel.POST("/consolidate", travelCtrl.SuggestConsolidation)
		travel.GET("/my-travels", travelCtrl.GetMyTravels)
	}

	// ==================== PHASE 7: MINIMUM MANNING MODE ====================

	// Phase 7 Initialization
	mmRepo := repository.NewMinimumManningRepository(db)
	mmRepo.EnsureIndexes(context.Background())

	mmSvc := service.NewMinimumManningService(
		mmRepo,
		vesselRepo,
		activityRepo,
		personnelRepo,
		roleAssignmentRepo,
		roleRepo,
		notifSvc,
	)

	mmCtrl := controllers.NewMinimumManningController(mmSvc)

	// Minimum Manning Routes
	manning := apiSecured.Group("/minimum-manning")
	{
		manning.POST("/activate", mmCtrl.Activate)
		manning.POST("/deactivate", mmCtrl.Deactivate)
		manning.GET("/active", mmCtrl.GetActiveEvent)
		manning.GET("/history", mmCtrl.GetEventHistory)
	}

	// ==================== PHASE 8: DASHBOARD & REPORTING ====================

	// Dashboard Service
	dashboardSvc := service.NewDashboardService(
		vesselRepo,
		activityRepo,
		certRepo,
		transportRepo,
		travelScheduleRepo,
		travelAssignmentRepo,
		personnelRepo,
		roomAssignRepo,
		vesselSvc,
	)
	dashboardCtrl := controllers.NewDashboardController(dashboardSvc)

	// Report Service
	reportSvc := service.NewReportService(
		vesselRepo,
		activityRepo,
		personnelRepo,
		roomAssignRepo,
		vesselSvc,
	)
	reportCtrl := controllers.NewReportController(reportSvc)

	// Dashboard Routes
	dashboard := apiSecured.Group("/dashboard")
	{
		dashboard.GET("", dashboardCtrl.GetDashboard)
	}

	// Report Routes
	reports := apiSecured.Group("/reports")
	{
		reports.GET("/daily", reportCtrl.DailyPOBReport)
		reports.GET("/historical", reportCtrl.HistoricalPOBReport)
		reports.GET("/export/pdf", reportCtrl.ExportPDF)
		reports.GET("/export/csv", reportCtrl.ExportCSV)
	}
}
