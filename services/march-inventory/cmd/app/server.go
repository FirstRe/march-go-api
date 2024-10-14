package main

import (
	// "fmt"

	"core/app/auth"
	"core/app/middlewares"
	"log"
	gormDb "march-inventory/cmd/app/common/gorm"
	graph "march-inventory/cmd/app/graph/generated"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/resolvers"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	// CORS package
)

const defaultPort = "8081"

func graphqlHandler() gin.HandlerFunc {
	c := graph.Config{Resolvers: &resolvers.Resolver{}}
	c.Directives.Auth = auth.Auth
	h := handler.NewDefaultServer(graph.NewExecutableSchema(c))

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

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

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
