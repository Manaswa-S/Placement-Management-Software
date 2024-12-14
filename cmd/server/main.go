package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mod/internal/config"
	"go.mod/internal/handlers"
	"go.mod/internal/middlewares"
	"go.mod/internal/services"
)

func main() {
	
	errenv := godotenv.Load()
	if errenv != nil {
		fmt.Println("Error loading environment variables: ", errenv)
		return
	}

	config.InitDB()

	defer config.Close()

	router := gin.Default()
	router.Use(middlewares.Logger())

	routes(router)

	router.Run(os.Getenv("PORT"))
}


func routes(router *gin.Engine) {

	wmid := router.Group("/laa")
	wmid.Use(middlewares.Authenticator(), middlewares.Authorizer())
	womid := router.Group("")
	womid.Use()

	queries := config.QueriesPool
	redis := config.RedisClient

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	openService := services.NewOpenService(queries)
	openHandler := handlers.NewOpenHandler(openService)
	openRoute := womid.Group("/open")
	openHandler.RegisterRoute(openRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	publicService := services.NewPublicService(queries, redis)
	publicHandler := handlers.NewPublicHandler(publicService)
	publicRoute := womid.Group("/public")
	publicHandler.RegisterRoute(publicRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	adminService := services.NewAdminService(queries)
	adminHandler := handlers.NewAdminHandler(adminService)
	adminRoute := wmid.Group("/admin")
	adminHandler.RegisterRoute(adminRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	companyService := services.NewCompanyService(queries)
	companyHandler := handlers.NewCompanyHandler(companyService)
	companyRoute := wmid.Group("/company")
	companyHandler.RegisterRoute(companyRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	studentService := services.NewStudentService(queries)
	studentHandler := handlers.NewStudentHandler(studentService)
	studentRoute := wmid.Group("/student")
	studentHandler.RegisterRoute(studentRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	superuserService := services.NewSuperService(queries)
	superuserHandler := handlers.NewSuperUserHandler(superuserService)
	superuserRoute := wmid.Group("/superuser")
	superuserHandler.RegisterRoute(superuserRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>


}