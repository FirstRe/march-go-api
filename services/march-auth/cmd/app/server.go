package main

import (
	"context"
	"core/app/auth"
	"core/app/middlewares"
	"log"
	gormDb "march-auth/cmd/app/common/gorm"
	graph "march-auth/cmd/app/graph/generated"
	"march-auth/cmd/app/graph/model"
	"march-auth/cmd/app/resolvers"
	"march-auth/cmd/app/services/uam"
	"net"

	pb "core/app/grpc"
	authService "march-auth/cmd/app/services/auth"

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

const defaultPort = "8080"
const defaultGrpcPort = "50080"

func initConfig() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
}

func graphqlHandler() gin.HandlerFunc {
	c := graph.Config{Resolvers: &resolvers.Resolver{}}
	c.Directives.Auth = auth.Auth
	introspection := viper.GetBool("GRAPHQL_INTROSPECTION")

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

func main() {

	initConfig()
	port := viper.GetString("PORT")
	grpcPort := viper.GetString("GRPCPORT")
	if port == "" {
		port = defaultPort
	}
	if grpcPort == "" {
		grpcPort = defaultGrpcPort
	}

	db, err := gormDb.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	db.AutoMigrate(
		&model.Function{},      // Independent
		&model.Shop{},          // Independent
		&model.Group{},         // Depends on Shop
		&model.User{},          // Depends on Group and Shop
		&model.Task{},          // Depends on Function
		&model.GroupFunction{}, // Depends on Group and Function
		&model.GroupTask{},     // Depends on Group, Task, and Shop
	)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	r.Use(middlewares.AuthMiddleware())
	r.POST("/auth/diviceId", uam.DiviceId)
	r.POST("/graphql", graphqlHandler())
	r.GET("/graphql/playground", playgroundHandler())

	// log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	// log.Fatal(r.Run("localhost:" + port))

	go func() {
		log.Printf("GraphQL server is running at http://localhost:%s/graphql/playground", port)
		if err := r.Run("localhost:" + port); err != nil {
			log.Fatalf("Failed to start GraphQL server: %v", err)
		}
	}()

	// Setup gRPC server
	lis, err := net.Listen("tcp", "localhost:"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuthGrpcServiceServer(grpcServer, &authService.Server{})

	// Start the gRPC server
	log.Printf("gRPC server is running on port %s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}

	select {}

}
