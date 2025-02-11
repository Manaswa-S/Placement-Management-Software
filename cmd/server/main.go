package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-ping/ping"
	"github.com/joho/godotenv"
	"go.mod/internal/apicalls"
	"go.mod/internal/config"
	"go.mod/internal/dto"
	"go.mod/internal/handlers"
	"go.mod/internal/middlewares"
	"go.mod/internal/notify"
	"go.mod/internal/services"
	"go.mod/internal/tasks"
	"go.mod/internal/utils"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/option"
)

var GAPIService *apicalls.Caller

func main() {

	skipTests := flag.Bool("skip-tests", false, "Skip time-consuming startup tests like connectivity checks.")
	flag.Parse()

	fmt.Println("Starting the PMS server...")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	errorsChan := make(chan *dto.ErrorData, 50)


	// load environment variables
	errenv := godotenv.Load()
	if errenv != nil {
		fmt.Println("Error loading environment variables: ", errenv)
		return
	}

	err := ErrorHandler(errorsChan)
	if err != nil {
		fmt.Println(err)
	}


	// skips time-consuming startup tests is flag passed
	if !*skipTests {
		// check for internet connectivity by pinging popular public DNS like 8.8.8.8 or 1.1.1.1
		err := internetCheck()
		if err != nil {
			fmt.Printf("Error checking internet connection : %v \n", err)
			return
		}
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
	// initialize the main router
	err = routerInit(errorsChan)
	if err != nil {
		fmt.Printf("Failed to initialize router : %v", err)
		return 
	}

	<-signals

	fmt.Println("\nReceived shutdown signal ...")
	// TODO: perform additional cleanup (if needed), like stopping background tasks
	config.Close()
	// Finally, exit the program
	fmt.Println("Shutdown complete.")
}

func routerInit(errorsChan chan *dto.ErrorData) error {

	// a default router, uses additional logger too
	router := gin.Default()
	router.Use(middlewares.Logger(errorsChan))
	routes(router)

	// serve static files, load dynamic templates
	router.Static("/static", "./template/static")
	router.Static("/scripts", "./template/scripts")

	router.LoadHTMLFiles("./template/company/newtest.html", "./template/student/takeTest.html")
	
	go func() {
		err := router.Run(os.Getenv("PORT"))
		if err != nil {
			fmt.Printf("Failed to run main router: %v\n", err)
			return
		}
	} ()

	return nil
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

	notifyService := notify.NewNotifyService(redis, queries)

	openService := services.NewOpenService(queries)
	openHandler := handlers.NewOpenHandler(openService)
	openRoute := womid.Group("/open")
	openHandler.RegisterRoute(openRoute)

	publicService := services.NewPublicService(queries, redis)
	publicHandler := handlers.NewPublicHandler(publicService)
	publicRoute := womid.Group("/public")
	publicHandler.RegisterRoute(publicRoute)

	adminService := services.NewAdminService(queries, GAPIService, notifyService)
	adminHandler := handlers.NewAdminHandler(adminService)
	adminRoute := wmid.Group("/admin")
	adminHandler.RegisterRoute(adminRoute)

	companyService := services.NewCompanyService(queries, GAPIService, redis, notifyService)
	companyHandler := handlers.NewCompanyHandler(companyService)
	companyRoute := wmid.Group("/company")
	companyHandler.RegisterRoute(companyRoute)

	studentService := services.NewStudentService(queries, redis, GAPIService, notifyService)
	studentHandler := handlers.NewStudentHandler(studentService)
	studentRoute := wmid.Group("/student")
	studentHandler.RegisterRoute(studentRoute)

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

	opts := option.WithCredentialsFile(serviceAccountKey)

	driveService, err := drive.NewService(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("error creating new drive service : %v", err)
	}
	formsService, err := forms.NewService(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("error creating new forms service : %v", err)
	}



	// this is unused as of now
	firebaseApp, err := firebase.NewApp(context.Background(), nil, opts)
	if err != nil {
		return fmt.Errorf("error creating new firebase app : %s", err)
	}
	fireMsgClient, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		return fmt.Errorf("error creating new firebase messaging client : %s", err)
	}


	GAPIService = apicalls.NewCaller(driveService, formsService, firebaseApp, fireMsgClient)

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

func ErrorHandler(errorsChan chan *dto.ErrorData) error {

	telegramBotToken, exists := os.LookupEnv("TelegramBotToken")
	if !exists || telegramBotToken == "" {
		return errors.New("TelegramBotToken not found or is empty string")
	}

	telegramChatID, exists := os.LookupEnv("TelegramChatID")
	if !exists || telegramChatID == "" {
		return errors.New("TelegramChatID not found or is empty string")
	} 

	go func() {
		for report := range errorsChan {
			// TODO: add the error handler logic where we send notifications to admin

			err := TelegramBot(report, telegramBotToken, telegramChatID)
			if err != nil {
				fmt.Println(err)
			}
		}
	} ()

	return nil
}

func TelegramBot(report *dto.ErrorData, bot_token string, chat_id string) error {

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", bot_token)

	textToSend := ""

	if report.Debug != "" {
		textToSend += fmt.Sprintf("<b>🔵 Debug:</b> %s\n", report.Debug)
	}
	if report.Info != "" {
		textToSend += fmt.Sprintf("<b>ℹ️ Info:</b> %s\n", report.Info)
	}
	if report.Warn != "" {
		textToSend += fmt.Sprintf("<b>⚠️ Warn:</b> %s\n", report.Warn)
	}
	if report.Error != "" {
		textToSend += fmt.Sprintf("<b>🔴 Error:</b> %s\n", report.Error)
	}
	if report.Critical != "" {
		textToSend += fmt.Sprintf("<b>🔥 Critical:</b> %s\n", report.Critical)
	}
	if report.Fatal != "" {
		textToSend += fmt.Sprintf("<b>☠️ Fatal:</b> %s\n", report.Fatal)
	}


	textToSend += fmt.Sprintf(
		"\n" +
		"<b>Start Time:</b> %s\n" + 
		"<b>Client IP:</b> %s\n" +
		"<b>Method:</b> %s\n" + 
		"<b>Path:</b> %s\n" + 
		"<b>Status Code:</b> %d\n" +
		"<b>Internal Error:</b> %s\n" + 
		"<b>Latency:</b> %s\n", 

		report.LogData.StartTime.Format("2006-01-02 15:04:05"), report.LogData.ClientIP, report.LogData.Method, report.LogData.Path,
		report.LogData.StatusCode, report.LogData.InternalError, report.LogData.Latency,
	)

	data := url.Values{}
	data.Set("chat_id", chat_id)
	data.Set("text", textToSend)
	data.Set("parse_mode", "HTML")


	var err error
	for i := 0; i < 3; i++ {
		_, err = http.PostForm(apiURL, data)
		if err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}

	return err
}