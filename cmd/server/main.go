package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ping/ping"
	"github.com/joho/godotenv"
	"go.mod/internal/apicalls"
	"go.mod/internal/config"
	"go.mod/internal/handlers"
	"go.mod/internal/middlewares"
	"go.mod/internal/services"
	"go.mod/internal/tasks"
	"go.mod/internal/utils"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/option"
)

var GAPIService *apicalls.Caller


func main() {

	fmt.Println("Starting the PMS server...")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// load environment variables
	errenv := godotenv.Load()
	if errenv != nil {
		fmt.Println("Error loading environment variables: ", errenv)
		return
	}

	// check for internet connectivity by pinging popular public DNS like 8.8.8.8 or 1.1.1.1
	err := internetCheck()
	if err != nil {
		fmt.Printf("Error checking internet connection : %v \n", err)
		return
	}

	// initialize the database, cache connections 
	err = config.InitDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	// initialize the API connections to external services
	err = GoogleAPIService()
	if err != nil {
		fmt.Println(err)
		return
	}
	// initialize the asynchronous functions 
	err = AsyncsInit()
	if err != nil {
		fmt.Println(err)
		return
	}

	// a default router, uses additional logger too
	router := gin.Default()
	router.Use(middlewares.Logger())
	routes(router)

	// serve static files, load dynamic templates
	router.Static("/static", "./template/static")
	router.LoadHTMLFiles("./template/company/newtest.html", "./template/student/takeTest.html")

	go func() {
		err := router.Run(os.Getenv("PORT"))
		if err != nil {
			fmt.Printf("Failed to run main router : %v", err)
			return
		}
	} ()

	<-signals


	fmt.Println("\nReceived shutdown signal ...")

	// TODO: perform additional cleanup (if needed), like stopping background tasks

	config.Close()

	// Finally, exit the program
	fmt.Println("Shutdown complete.")
}

func internetCheck() (error) {

	fmt.Println("Checking for internet connectivity...")

	pinger, err := ping.NewPinger(config.PingTesterIP)
	if err != nil {
		return err
	}
	pinger.Timeout = time.Second * 3
	pinger.Count = 10
	err = pinger.Run()
	if err != nil {
		return err
	}
	stats := pinger.Statistics()
	
	fmt.Printf("Pinged %s : %s .\n" + 
			   "Packets: \n" + 
			   "       Sent : %d \n" + 
			   "   Recieved : %d \n" + 
			   "       Loss : %.2f %%\n" + 
			   "Average RTT : %.2f ms\n" + 
			   "Minimum RTT : %.2f ms\n" + 
			   "Maximum RTT : %.2f ms\n",	
			   						stats.Addr, 
			   						stats.IPAddr, 
									stats.PacketsSent, 
									stats.PacketsRecv, 
									stats.PacketLoss,
									stats.AvgRtt.Seconds() * 1000,
									stats.MinRtt.Seconds() * 1000,
									stats.MaxRtt.Seconds() * 1000)


	return nil
}

func routes(router *gin.Engine) {

	wmid := router.Group("/laa")
	wmid.Use(middlewares.Authenticator(), middlewares.Authorizer())
	womid := router.Group("")
	womid.Use()

	// TODO: 
	womid.GET("/favicon.ico", func(ctx *gin.Context) {
		ctx.File("./favicon.ico")
	})

	wmid.POST("/report", func(ctx *gin.Context) {
		utils.RecordReport(ctx)
	})
	
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
	adminService := services.NewAdminService(queries, GAPIService)
	adminHandler := handlers.NewAdminHandler(adminService)
	adminRoute := wmid.Group("/admin")
	adminHandler.RegisterRoute(adminRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	companyService := services.NewCompanyService(queries, GAPIService, redis)
	companyHandler := handlers.NewCompanyHandler(companyService)
	companyRoute := wmid.Group("/company")
	companyHandler.RegisterRoute(companyRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	studentService := services.NewStudentService(queries, redis, GAPIService)
	studentHandler := handlers.NewStudentHandler(studentService)
	studentRoute := wmid.Group("/student")
	studentHandler.RegisterRoute(studentRoute)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	superuserService := services.NewSuperService(queries)
	superuserHandler := handlers.NewSuperUserHandler(superuserService)
	superuserRoute := wmid.Group("/superuser")
	superuserHandler.RegisterRoute(superuserRoute)
}

func GoogleAPIService() (error) {

	serviceAccountKey := os.Getenv("PathToServiceAccountKey")
	if serviceAccountKey == "" {
		return fmt.Errorf("path to service account key not found in environment variables")
	}
	driveService, err := drive.NewService(context.Background(), option.WithCredentialsFile(serviceAccountKey))
	if err != nil {
		return fmt.Errorf("error creating new drive service : %s", err)
	}
	formsService, err := forms.NewService(context.Background(), option.WithCredentialsFile(serviceAccountKey))
	if err != nil {
		return fmt.Errorf("error creating new forms service : %s", err)
	}

	GAPIService = apicalls.NewCaller(driveService, formsService)

	fmt.Println("Getting New DrivePageToken to start with ...")
	_, err  = GAPIService.DriveChanges()
	if err != nil {
		return err
	}

	return nil
}

func AsyncsInit() error {

	aService := tasks.NewAsyncService(config.QueriesPool, GAPIService)
	
	err := aService.StartAsyncs()
	if err != nil {
		return err
	}

	return nil
}