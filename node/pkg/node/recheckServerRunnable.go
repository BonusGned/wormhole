package node

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/certusone/wormhole/node/pkg/db"
	"github.com/certusone/wormhole/node/pkg/governor"
	nodev1 "github.com/certusone/wormhole/node/pkg/proto/node/v1"
	recheck "github.com/certusone/wormhole/node/pkg/recheck"
	"github.com/certusone/wormhole/node/pkg/supervisor"
	"github.com/certusone/wormhole/node/pkg/watchers/evm/connectors"
	eth_common "github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	// "github.com/gagliardetto/solana-go/rpc"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func recheckHttpServiceRunnable(
	logger *zap.Logger,
	address string,
	adminRPC string,
	db *db.Database,
	gst *common.GuardianSetState,
	gov *governor.ChainGovernor,
	ethUrl string,
	ethContract string,
	solanaRPC string,
	solanaContract string,

) (supervisor.Runnable, error) {

	return func(ctx context.Context) error {
		// Connect to guardian admin RPC
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		ethContract := eth_common.HexToAddress(ethContract)
		solanaContract := solana.MustPublicKeyFromBase58(solanaContract)
		baseConnector, err := connectors.NewEthereumBaseConnector(ctx, "Ultron", ethUrl, ethContract, logger)
		solanaClient := rpc.New(solanaRPC)
		conn, err := grpc.DialContext(
			ctx,
			fmt.Sprintf("unix:///%s", adminRPC),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return fmt.Errorf("failed to connect to admin RPC %s: %v", adminRPC, err)
		}
		defer conn.Close()

		adminClient := nodev1.NewNodePrivilegedServiceClient(conn)

		// Create and configure the HTTP server
		server, err := recheck.NewRecheckServer(
			adminClient,
			db,
			logger,
			gst,
			gov,
			baseConnector,
			ethContract,
			solanaClient,
			solanaContract,
		)
		if err != nil {
			return fmt.Errorf("failed to create server: %v", err)
		}

		router := mux.NewRouter()
		server.RegisterRoutes(router)

		srv := &http.Server{
			Handler:      router,
			Addr:         address,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		errC := make(chan error)
		go func() {
			logger.Info("HTTP server listening", zap.String("addr", srv.Addr))
			errC <- srv.ListenAndServe()
		}()

		supervisor.Signal(ctx, supervisor.SignalHealthy)

		select {
		case <-ctx.Done():
			// Graceful shutdown
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Error("HTTP server shutdown error", zap.Error(err))
				return err
			}
			return ctx.Err()
		case err := <-errC:
			return fmt.Errorf("HTTP server error: %w", err)
		}
	}, nil
}
