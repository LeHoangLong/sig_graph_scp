package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sig_graph_scp/cmd/middleware"
	"sig_graph_scp/cmd/view"
	api_asset_transfer "sig_graph_scp/pkg/asset_transfer/api"
	controller_server "sig_graph_scp/pkg/server/controller"
	repository_server "sig_graph_scp/pkg/server/repository"
	api_sig_graph "sig_graph_scp/pkg/sig_graph/api"
	"sig_graph_scp/pkg/utility"

	EventBus "github.com/asaskevich/eventbus"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// event bus
	eventBus := EventBus.New()

	// sig graph api
	assetApi, err := api_sig_graph.NewAssetClientApi("sgp://hyper:[http://localhost:7051,http://localhost:9051]:public")
	if err != nil {
		panic(fmt.Sprintf("could not create asset client api: %s", err))
	}

	// asset transfer api
	assetTransferApi, err := api_asset_transfer.NewAssetTransferServiceApi(
		"sgp://hyper:[http://localhost:7051,http://localhost:9051]:public",
		&api_asset_transfer.Options{
			NumberOfCandidates: 6,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("could not create asset transfer api: %s", err))
	}

	// asset transfer server api
	assetTransferServerGrpcAddress := os.Getenv("ASSET_TRANSFER_SERVER_GRPC_ADDRESS")
	if assetTransferServerGrpcAddress == "" {
		assetTransferServerGrpcAddress = "localhost:5000"
	}
	assetTransferServerApi, err := api_asset_transfer.NewAssetTransferServerApi(
		assetTransferServerGrpcAddress,
		api_asset_transfer.AssetTransferServerApiOptions{
			EventBus: eventBus,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("could not create asset transfer api: %s", err))
	}
	go func() {
		err := assetTransferServerApi.Start()
		if err != nil {
			panic(fmt.Sprintf("could not start asset transfer grpc server: %s", err.Error()))
		}
	}()

	router := gin.Default()

	// api

	// init db
	connectionStr := os.Getenv("DB_CONNECTION")
	if connectionStr == "" {
		panic("missing DB_CONNECTION env var")
	}

	db, err := gorm.Open(postgres.Open(connectionStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("could not connect to db: %s", err))
	}

	// utility classes
	clock := utility.NewClockWall()

	// transaction manager
	transactionManager := repository_server.NewTransactionManagerGorm(db)

	// migrate database
	versionRepository := repository_server.NewMigratorVersionRepositoryGorm(*transactionManager)
	{
		ctx := context.Background()
		err := versionRepository.CreatTableIfNotExists(ctx)
		if err != nil {
			panic(fmt.Sprintf("could not create verion table: %s", err))
		}
	}
	migrator := repository_server.NewMigratorGorm(&versionRepository, transactionManager)
	{
		ctx := context.Background()
		err := migrator.Up(ctx, 3)
		if err != nil {
			panic(fmt.Sprintf("could not migrate database: %s", err))
		}
	}

	// repository
	nodeRepository := repository_server.NewNodeRepositoryGorm(transactionManager)
	assetRepository := repository_server.NewAssetRepositoryGorm(transactionManager, nodeRepository)
	userKeyPairRepository := repository_server.NewUserKeyRepositoryGorm(transactionManager)
	peerRepository := repository_server.NewPeerRepositoryGorm(transactionManager)
	assetTransferRepository := repository_server.NewAssetTransferRepositoryGorm(*transactionManager)

	// controller
	assetController := controller_server.NewAssetController(assetApi, assetRepository, userKeyPairRepository, transactionManager)
	userKeyPairController := controller_server.NewUserKeyPairController(userKeyPairRepository, transactionManager)
	peerController := controller_server.NewPeerController(transactionManager, peerRepository)
	assetTransferController := controller_server.NewAssetTransferController(
		clock,
		assetTransferApi,
		assetRepository,
		transactionManager,
		userKeyPairRepository,
		nodeRepository,
		peerRepository,
		assetController,
		assetTransferRepository,
	)

	{
		ctx := context.Background()
		assetTransferController.SubscribeNewAcceptAssetRequestReceivedEvent(
			ctx,
			eventBus,
			assetTransferServerApi.GetDefaultBusName(),
		)
	}

	// view
	assetView := view.NewAssetView(assetController)
	userKeyPairView := view.NewUserKeyPairView(userKeyPairController)
	peerView := view.NewPeerView(peerController)
	assetTransferView := view.NewAssetTransferView(assetTransferController)

	//authenticator
	auth := middleware.NewAuthenticatorSimple()

	// api
	api := router.Group("/api")
	{
		// asset
		api.GET("/assets", auth.Authenticate, assetView.GetAssetById)
		api.POST("/assets", auth.Authenticate, assetView.CreateAsset)

		// user key pair
		api.GET("/key_pairs", auth.Authenticate, userKeyPairView.GetUserKeyPairsByUser)
		api.POST("/key_pairs", auth.Authenticate, userKeyPairView.AddUserKeyPairToUser)

		// peers
		api.GET("/peers", auth.Authenticate, peerView.GetPeers)
		api.POST("/peers", auth.Authenticate, peerView.CreatePeer)

		// asset transfer
		api.POST("/asset_accept_requests", auth.Authenticate, assetTransferView.CreateRequestToAcceptAsset)
		api.GET("/asset_accept_requests", auth.Authenticate, assetTransferView.GetReceivedRequestToAcceptAsset)
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
}
