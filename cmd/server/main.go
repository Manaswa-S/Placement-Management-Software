package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mod/internal/db"
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

	db.InitDB()

	defer db.Close()

	router := gin.Default()
	router.Use(middlewares.Logger())

	wmid := router.Group("/laa")
	wmid.Use(middlewares.Authenticator(), middlewares.Authorizer())
	womid := router.Group("")
	womid.Use()


// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	publicService := services.PublicService{Queries: db.QueriesPool}
	publicHandler := handlers.NewPublicHandler(&publicService)
	publicRoute := womid.Group("/public")
	publicHandler.RegisterRoute(publicRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	adminService := services.AdminService{Queries: db.QueriesPool}
	adminHandler := handlers.NewAdminHandler(&adminService)
	adminRoute := wmid.Group("/admin")
	adminHandler.RegisterRoute(adminRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	companyService := services.CompanyService{Queries: db.QueriesPool}
	companyHandler := handlers.NewCompanyHandler(&companyService)
	companyRoute := wmid.Group("/company")
	companyHandler.RegisterRoute(companyRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	studentService := services.StudentService{Queries: db.QueriesPool}
	studentHandler := handlers.NewStudentHandler(&studentService)
	studentRoute := wmid.Group("/student")
	studentHandler.RegisterRoute(studentRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	superuserService := services.SuperUserService{Queries: db.QueriesPool}
	superuserHandler := handlers.NewSuperUserHandler(&superuserService)
	superuserRoute := wmid.Group("/superuser")
	superuserHandler.RegisterRoute(superuserRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>


	router.Run(os.Getenv("PORT"))
}
