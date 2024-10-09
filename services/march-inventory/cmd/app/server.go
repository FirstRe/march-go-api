package main

import (
	// "fmt"
	auth "core/app/auth"
	"core/app/middlewares"
	"log"
	gormDb "march-inventory/cmd/app/common/gorm"
	graph "march-inventory/cmd/app/graph/generated"
	"march-inventory/cmd/app/graph/model"
	"march-inventory/cmd/app/resolvers"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	// "github.com/jinzhu/gorm"
)

const defaultPort = "8081"

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	db, err := gormDb.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&model.Inventory{}, &model.InventoryBranch{}, &model.InventoryBrand{}, &model.InventoryFile{}, &model.InventoryType{})

	// db.Session(&gorm.Session{SkipHooks: false}).Create(&basedb.UserRe{})
	// db.AutoMigrate(&basedb.UserRe{})
	//_, err := config.InitDb()
	//if err != nil {
	//	log.Fatal(err)
	//}

	router := mux.NewRouter()
	router.Use(middlewares.AuthMiddleware)

	c := graph.Config{Resolvers: &resolvers.Resolver{}}
	c.Directives.Auth = auth.Auth

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(c))

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe("localhost:"+port, router))
}
