package recheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/certusone/wormhole/node/pkg/db"
	"github.com/certusone/wormhole/node/pkg/governor"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	nodev1 "github.com/certusone/wormhole/node/pkg/proto/node/v1"
	"github.com/certusone/wormhole/node/pkg/watchers/evm"
	"github.com/certusone/wormhole/node/pkg/watchers/evm/connectors"
	"github.com/certusone/wormhole/node/pkg/watchers/evm/connectors/ethabi"
	abi2 "github.com/ethereum/go-ethereum/accounts/abi"
	eth_common "github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gorilla/mux"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type RecheckServer struct {
	adminClient    nodev1.NodePrivilegedServiceClient
	ethAbi         abi2.ABI
	logger         *zap.Logger
	db             *db.Database
	gst            *common.GuardianSetState
	gov            *governor.ChainGovernor
	ethConnector   connectors.Connector
	ethContract    *eth_common.Address
	solanaClient   *rpc.Client
	solanaContract *solana.PublicKey
}

type ObservationRequest struct {
	ChainID          string   `json:"chainId"`
	TransactionHashs []string `json:"txHashs"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var AVALIABLE_CHAIN_IDS = [2]uint8{1, 2}

func checkChainID(chainID uint8) bool {
	for _, cid := range AVALIABLE_CHAIN_IDS {
		if cid == chainID {
			return true
		}
	}
	return false
}

func NewRecheckServer(
	adminClient nodev1.NodePrivilegedServiceClient,
	db *db.Database,
	logger *zap.Logger,
	gst *common.GuardianSetState,
	gov *governor.ChainGovernor,
	ethConnector connectors.Connector,
	ethContract eth_common.Address,
	solanaClient *rpc.Client,
	solanaContract solana.PublicKey,
) (*RecheckServer, error) {
	ethAbi, err := abi2.JSON(strings.NewReader(ethabi.AbiABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Eth ABI: %v", err)
	}

	return &RecheckServer{
		adminClient:    adminClient,
		ethAbi:         ethAbi,
		logger:         logger,
		db:             db,
		gst:            gst,
		gov:            gov,
		ethConnector:   ethConnector,
		ethContract:    &ethContract,
		solanaClient:   solanaClient,
		solanaContract: &solanaContract,
	}, nil
}
func (s *RecheckServer) RegisterRoutes(r *mux.Router) {
    r.HandleFunc("/recheck", s.handleObservationRequest).Methods(http.MethodPost)
}

func (s *RecheckServer) handleObservationRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ObservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate chain ID
	chainID, err := vaa.ChainIDFromString(req.ChainID)
	if !checkChainID(uint8(chainID)) {
		writeJSONError(w, "Invalid chain ID", http.StatusBadRequest)
		return
	}
	if err != nil {
		writeJSONError(w, fmt.Sprintf("Invalid chain ID: %v", err), http.StatusBadRequest)
		return
	}

	// Validate and normalize transaction hash
	for _, txHash := range req.TransactionHashs {
		if !strings.HasPrefix(txHash, "0x") {
			writeJSONError(w, "Invalid transaction hash", http.StatusBadRequest)
			return
		}

		txHash := eth_common.HexToHash(txHash)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_, msgs, err := evm.MessageEventsForTransaction(ctx, s.ethConnector, *s.ethContract, chainID, txHash)
		// Send observation request
		if len(msgs) == 0 {
			writeJSONError(w, "Invalid transaction hash", http.StatusBadRequest)
			return
		}

		_, err = s.adminClient.SendObservationRequest(ctx, &nodev1.SendObservationRequestRequest{
			ObservationRequest: &gossipv1.ObservationRequest{
				ChainId: uint32(chainID),
				TxHash:  txHash.Bytes(),
			},
		})
		if err != nil {
			writeJSONError(w, fmt.Sprintf("Failed to send observation request: %v", err), http.StatusInternalServerError)
			return
		}

	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "observation request sent",
	})
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
