package cmd

import (
	"fmt"
	"net/http"
	"os"
	"sig_graph_scp/cmd/middleware"
	"sig_graph_scp/cmd/view"
	service_sig_graph "sig_graph_scp/internal/sig_graph/service"
	controller_server "sig_graph_scp/pkg/server/controller"
	repository_server "sig_graph_scp/pkg/server/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() int {
	router := gin.Default()

	// middleware
	authenticator := middleware.NewAuthenticatorSimple()

	// service
	smartContractService, err := service_sig_graph.NewSmartContractServiceHyperledger()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize smart contract service: %s", err))
	}
	assetSigGraphService := service_sig_graph.NewAssetService(smartContractService)

	// init db
	connectionStr := os.Getenv("DB_CONNECTION")
	if connectionStr == "" {
		panic("missing DB_CONNECTION env var")
	}

	db, err := gorm.Open(postgres.Open(connectionStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("could not connect to db: %s", err))
	}

	// repository
	transactionManager := repository_server.NewTransactionManagerGorm(db)
	nodeRepository := repository_server.NewNodeRepositoryGorm(transactionManager)
	assetRepository := repository_server.NewAssetRepositoryGorm(transactionManager, nodeRepository)
	userKeyPairRepository := repository_server.NewUserKeyRepositoryGorm(transactionManager)

	// controller
	assetController := controller_server.NewAssetController(assetSigGraphService, assetRepository, userKeyPairRepository, transactionManager)

	// view
	assetView := view.NewAssetView(assetController)

	// api
	api := router.Group("/api")
	api.Use(authenticator.Authenticate)
	{
		// asset
		api.Use()
		api.GET("/asset", assetView.GetAssetById)
	}

	router.NoRoute(func(ctx *gin.Context) { ctx.JSON(http.StatusNotFound, gin.H{}) })

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = "localhost:8080"
	}

	fmt.Println("Starting host at ", serverAddress)
	err = router.Run(serverAddress)
	if err != nil {
		panic(fmt.Sprintf("could not start server: %s", err))
	}

	return 0
}
