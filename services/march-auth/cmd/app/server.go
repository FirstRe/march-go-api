package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"core/app/auth"
	pb "core/app/grpc"
	"core/app/middlewares"

	gormDb "march-auth/cmd/app/common/gorm"
	graph "march-auth/cmd/app/graph/generated"
	"march-auth/cmd/app/graph/model"
	"march-auth/cmd/app/resolvers"
	authService "march-auth/cmd/app/services/auth"
	"march-auth/cmd/app/services/uam"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	defaultPort     = "8080"
	defaultGrpcPort = "5005"
)

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

func graphqlHandler() gin.HandlerFunc {
	cfg := graph.Config{Resolvers: &resolvers.Resolver{}}
	cfg.Directives.Auth = auth.Auth
	introspectionString := os.Getenv("GRAPHQL_INTROSPECTION")
	introspection, _ := strconv.ParseBool(introspectionString)
	h := handler.NewDefaultServer(graph.NewExecutableSchema(cfg))
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

func setupDatabase() {
	db, err := gormDb.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	db.AutoMigrate(
		&model.Function{},
		&model.Shop{},
		&model.Group{},
		&model.User{},
		&model.Task{},
		&model.GroupFunction{},
		&model.GroupTask{},
	)

}

func setupGinRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
	}))

	r.Use(middlewares.AuthMiddleware())
	r.POST("/auth/diviceId", uam.DiviceId)
	r.POST("/graphql", graphqlHandler())
	r.GET("/graphql/playground", playgroundHandler())

	return r
}

func startGraphQLServer(router *gin.Engine, port string) {
	log.Printf("GraphQL server is running at http://localhost:%s/graphql/playground", port)
	if err := router.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Failed to start GraphQL server: %v", err)
	}
}

func startGrpcServer(grpcPort string) {
	lis, err := net.Listen("tcp", "localhost:"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.UnaryInterceptor))
	pb.RegisterAuthGrpcServiceServer(grpcServer, &authService.Server{})

	log.Printf("gRPC server is running on port %s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}

func main() {
	initConfig()

	port := os.Getenv("PORT")
	grpcPort := viper.GetString("auth.grpc.port")

	if port == "" {
		port = defaultPort
	}
	if grpcPort == "" {
		grpcPort = defaultGrpcPort
	}

	setupDatabase()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	router := setupGinRouter()
	go startGraphQLServer(router, port)
	startGrpcServer(grpcPort)
}
