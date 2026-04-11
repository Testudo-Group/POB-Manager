package routes

import (
	"context"
	"log"

	"github.com/codingninja/pob-management/config"
	"github.com/codingninja/pob-management/internal/delivery/http/controllers"
	"github.com/codingninja/pob-management/internal/delivery/http/middleware"
	"github.com/codingninja/pob-management/internal/repository"
	"github.com/codingninja/pob-management/internal/service"
	_ "github.com/codingninja/pob-management/docs" // Custom docs
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Setup(r *gin.Engine, db *mongo.Database, rdb *redis.Client, cfg *config.Config) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

	certTypeRepo.EnsureIndexes(context.Background())
	roleRepo.EnsureIndexes(context.Background())
	personnelRepo.EnsureIndexes(context.Background())
	certRepo.EnsureIndexes(context.Background())

	_ = service.NewCertificateTypeService(certTypeRepo)
	certSvc := service.NewCertificateService(certRepo, certTypeRepo)
	_ = certSvc // To avoid unused variable issue in the future
	roleSvc := service.NewOffshoreRoleService(roleRepo)
	personnelSvc := service.NewPersonnelService(personnelRepo, roleRepo)
	compSvc := service.NewComplianceService(personnelRepo, roleRepo, certRepo, certTypeRepo)
	
	roleCtrl := controllers.NewOffshoreRoleController(roleSvc)
	personnelCtrl := controllers.NewPersonnelController(personnelSvc, compSvc)

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
	}

	personnel := apiSecured.Group("/personnel")
	{
		personnel.POST("", personnelCtrl.CreatePersonnel)
		personnel.GET("", personnelCtrl.ListPersonnel)
		personnel.GET("/:id/compliance", personnelCtrl.CheckCompliance)
	}

	// Phase 3 Initialization
	vesselRepo := repository.NewVesselRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	roomAssignRepo := repository.NewRoomAssignmentRepository(db)

	vesselRepo.EnsureIndexes(context.Background())
	roomRepo.EnsureIndexes(context.Background())
	roomAssignRepo.EnsureIndexes(context.Background())

	vesselSvc := service.NewVesselService(vesselRepo, rdb)
	roomSvc := service.NewRoomService(roomRepo, roomAssignRepo, vesselSvc)

	vesselCtrl := controllers.NewVesselController(vesselSvc)
	roomCtrl := controllers.NewRoomController(roomSvc)

	// Phase 3 Routes
	vessels := apiSecured.Group("/vessels")
	{
		vessels.POST("", vesselCtrl.CreateVessel)
		vessels.GET("", vesselCtrl.ListVessels)
		vessels.GET("/:id/pob", vesselCtrl.GetRealTimePOB)

		vessels.POST("/:vesselId/rooms", roomCtrl.CreateRoom)
		vessels.POST("/:vesselId/rooms/assign", roomCtrl.AssignRoom)
	}
}
