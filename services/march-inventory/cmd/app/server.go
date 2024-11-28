package main

import (
	"context"
	"core/app/auth"
	"core/app/middlewares"
	"log"
	gormDb "march-inventory/cmd/app/common/gorm"
	graph "march-inventory/cmd/app/graph/generated"
	"march-inventory/cmd/app/graph/model"
	translation "march-inventory/cmd/app/i18n"
	"march-inventory/cmd/app/resolvers"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

const defaultPort = "8081"

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
	if port == "" {
		port = defaultPort
	}
	translation.InitI18n()
	db, err := gormDb.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	db.AutoMigrate(&model.Inventory{}, &model.InventoryBranch{}, &model.InventoryBrand{}, &model.InventoryFile{}, &model.InventoryType{})

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	r.Use(middlewares.AuthMiddleware())
	r.POST("/graphql", graphqlHandler())
	r.GET("/graphql/playground", playgroundHandler())

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(r.Run("localhost:" + port))

}
