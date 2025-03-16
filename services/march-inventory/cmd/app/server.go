package main

import (
	"context"
	"core/app/auth"
	"core/app/middlewares"
	"fmt"
	"log"
	gormDb "march-inventory/cmd/app/common/gorm"
	"march-inventory/cmd/app/common/redis"
	graph "march-inventory/cmd/app/graph/generated"
	"march-inventory/cmd/app/graph/model"
	translation "march-inventory/cmd/app/i18n"
	"march-inventory/cmd/app/repositories"
	"march-inventory/cmd/app/resolvers"
	"march-inventory/cmd/app/services/newInventory"
	"net/http"
	"os"
	"strconv"
	"strings"

	grpcCilent "march-inventory/cmd/app/common/grpc"
	// grpcAuth "march-inventory/cmd/app/common/grpc/auth"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const defaultPort = "8081"

func initConfig() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/appname/")
	viper.AddConfigPath("$HOME/.appname")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app/services/march-auth/")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error configs file: %w", err))
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func graphqlHandler(inventoryService newInventory.InventoryService) gin.HandlerFunc {
	c := graph.Config{Resolvers: &resolvers.Resolver{
		InventoryService: inventoryService,
	}}
	c.Directives.Auth = auth.Auth
	introspectionString := os.Getenv("GRAPHQL_INTROSPECTION")
	introspection, _ := strconv.ParseBool(introspectionString)
	h := handler.NewDefaultServer(graph.NewExecutableSchema(c))
	h.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		if !introspection {
			graphql.GetOperationContext(ctx).DisableIntrospection = true
		}

		return next(ctx)
	})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/graphql")
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func setupDatabase() *gorm.DB {
	db, err := gormDb.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	db.AutoMigrate(&model.Inventory{},
		&model.InventoryBranch{},
		&model.InventoryBrand{},
		&model.InventoryFile{},
		&model.InventoryType{},
	)
	return db
}

func setupGinRouter(inventoryService newInventory.InventoryService) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	r.Use(middlewares.AuthMiddleware())
	r.POST("/graphql", graphqlHandler(inventoryService))
	r.GET("/graphql/playground", playgroundHandler())
	r.POST("/upload", func(c *gin.Context) {
		// Uploading file via multipart form
		fmt.Println("Uploading file via multipart form")
		file, _ := c.FormFile("file")
		if err := c.SaveUploadedFile(file, "./"+file.Filename); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("File %s uploaded successfully!", file.Filename)})
	})
	return r
}

func startGraphQLServer(router *gin.Engine, port string) {
	log.Printf("GraphQL server is running at http://localhost:%s/graphql/playground", port)
	if err := router.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Failed to start GraphQL server: %v", err)
	}
}

// func startGrpcServer(grpcPort string) {
// 	lis, err := net.Listen("tcp", "localhost:"+grpcPort)
// 	if err != nil {
// 		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
// 	}

// 	grpcServer := grpc.NewServer()
// 	pb.RegisterAuthGrpcServiceServer(grpcServer, &authService.Server{})

// 	log.Printf("gRPC server is running on port %s", grpcPort)
// 	if err := grpcServer.Serve(lis); err != nil {
// 		log.Fatalf("Failed to start gRPC server: %v", err)
// 	}
// }

func main() {

	initConfig()
	port := os.Getenv("PORT")
	// grpcPort := viper.GetString("inventory.grpc.port")

	if port == "" {
		port = defaultPort
	}
	// if grpcPort == "" {
	// 	grpcPort = defaultGrpcPort
	// }
	
	redis.InitRedis()

	translation.InitI18n()
	db := setupDatabase()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	inventoryRepo := repositories.NewProductRepository(db)
	inventoryService := newInventory.NewInventoryServiceRedis(redis.RedisClient, inventoryRepo)

	//grpc cilent
	connections := grpcCilent.Init()

	defer func() {
		for _, conn := range connections {
			conn.Close()
		}
	}()

	// shopIds := "984d0d87-7d74-45c5-9d94-6ebcb74a98de"
	// r, err := grpcAuth.GetPermission(shopIds, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiU1VQRVJBRE1JTiIsImluZm8iOnsiZnVuY3Rpb25zIjpbIk1FTlU6SU5WRU5UT1JZIiwiTUVOVTpVU0VSIiwiTUVOVTpEQVNIQk9BUkQiLCJNRU5VOkNVU1RPTUVSIiwiTUVOVTpTQUxFUyJdLCJ0YXNrcyI6WyJJTkJyYW5jaFZpZXdlciIsIklOQnJhbmNoTWFrZXIiLCJJTkJyYW5kVmlld2VyIiwiSU5NYWtlciIsIklOQnJhbmRNYWtlciIsIklOVmlld2VyIiwiSU5DU1YiLCJJTlRyYXNoTWFrZXIiLCJJTlR5cGVNYWtlciIsIklOVHlwZVZpZXdlciJdLCJwYWdlIjp7Ik1FTlU6Q1VTVE9NRVIiOlsiQ1JFQVRFIiwiVklFVyIsIlVQREFURSJdLCJNRU5VOkRBU0hCT0FSRCI6WyJDUkVBVEUiLCJWSUVXIiwiVVBEQVRFIl0sIk1FTlU6SU5WRU5UT1JZIjpbIkNSRUFURSIsIlZJRVciLCJVUERBVEUiXSwiTUVOVTpTQUxFUyI6WyJDUkVBVEUiLCJWSUVXIiwiVVBEQVRFIl0sIk1FTlU6VVNFUiI6WyJDUkVBVEUiLCJWSUVXIiwiVVBEQVRFIl19fSwiZGV2aWNlSWQiOiJmNzYyNTg1YS04ODUxLTRmMmYtOTRjNi1lMWFkNTdlMWY2ZGMiLCJ1c2VySWQiOiJmOGJhYjJhNC1jYjM5LTQ5OWMtODI3Yy1jMjI1MDczNjk4NDciLCJzaG9wc0lkIjoiOTg0ZDBkODctN2Q3NC00NWM1LTlkOTQtNmViY2I3NGE5OGRlIiwic2hvcE5hbWUiOiJmaXJzdF9zaG9wIiwidXNlck5hbWUiOiJOb2NoVGljaCIsInBpY3R1cmUiOiJodHRwczovL2xoMy5nb29nbGV1c2VyY29udGVudC5jb20vYS9BQ2c4b2NKa0JSS3RreEU4eFhXTWN0ajBxaVduMG0wLUpxWFNCWGsxaFhBd19xSG5NUlB5Umk2Nj1zOTYtYyIsImV4cCI6MTczNjQzMzk3NSwiaWF0IjoxNzM1ODI5MTc1fQ.kHqJ3R2Cwa7wUnWC2cJeNeQtXu6Tz87tpGMXNy8dpmc")
	// if err != nil {
	// 	log.Printf("could not greet: %v", err)
	// }
	// log.Printf("Greeting22: %s", r.GetShop())

	//grpc cilent

	router := setupGinRouter(inventoryService)
	startGraphQLServer(router, port)
	// startGrpcServer(grpcPort)

}
