package recheck

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/certusone/wormhole/node/pkg/db"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	nodev1 "github.com/certusone/wormhole/node/pkg/proto/node/v1"
	"github.com/certusone/wormhole/node/pkg/watchers/evm"
	eth_common "github.com/ethereum/go-ethereum/common"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

func checkAndSendObservationEVM(s *RecheckServer, w http.ResponseWriter, req ObservationRequest, chainID vaa.ChainID) bool {
	for _, txHash := range req.TransactionHashs {
		if !strings.HasPrefix(txHash, "0x") {
			writeJSONError(w, "Invalid transaction hash", http.StatusBadRequest)
			return false
		}

		txHash := eth_common.HexToHash(txHash)
		ctx := context.Background()
		_, msgs, err := evm.MessageEventsForTransaction(ctx, s.ethConnector, *s.ethContract, chainID, txHash)
		// Send observation request
		if len(msgs) == 0 {
			writeJSONError(w, "Invalid transaction hash", http.StatusBadRequest)
			return false
		}
		for _, msg := range msgs {

			_, err := s.db.GetSignedVAABytes(db.VAAID{
				EmitterChain:   chainID,
				EmitterAddress: msg.EmitterAddress,
				Sequence:       msg.Sequence,
			})

			if err != nil {
				if err == db.ErrVAANotFound {
					continue
				}
				s.logger.Error("failed to fetch VAA", zap.Error(err), zap.Any("request", req))
				writeJSONError(w, fmt.Sprintf("Failed to fetch VAA: %v", err), http.StatusInternalServerError)
				return false
			} else {
				msg := fmt.Sprintf("VAA already exists: emitterChain=%d emitterAddress=%s sequence=%d txHash=%s", msg.EmitterChain, msg.EmitterAddress, msg.Sequence, msg.TxHash)
				writeJSONError(w, msg, http.StatusBadRequest)
				return false
			}
		}

		_, err = s.adminClient.SendObservationRequest(ctx, &nodev1.SendObservationRequestRequest{
			ObservationRequest: &gossipv1.ObservationRequest{
				ChainId: uint32(chainID),
				TxHash:  txHash.Bytes(),
			},
		})
		if err != nil {
			writeJSONError(w, fmt.Sprintf("Failed to send observation request: %v", err), http.StatusInternalServerError)
			return false
		}
	}
	return true
}
