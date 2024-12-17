package recheck

import (
	"context"
	"fmt"
	"net/http"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	nodev1 "github.com/certusone/wormhole/node/pkg/proto/node/v1"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	postMessageInstructionID = 0x01
)

func checkAndSendObservationSolana(s *RecheckServer, w http.ResponseWriter, req ObservationRequest, chainID vaa.ChainID) bool {
	if len(req.TransactionHashs) == 0 {
		writeJSONError(w, "No transaction hashes provided", http.StatusBadRequest)
		return false
	}

	for _, txHash := range req.TransactionHashs {
		// Validate Solana transaction hash
		signature, err := solana.SignatureFromBase58(txHash)
		if err != nil {
			writeJSONError(w, "Invalid Solana transaction hash", http.StatusBadRequest)
			return false
		}

		ctx := context.Background()

		// Validate solanaClient is not nil
		if s.solanaClient == nil {
			s.logger.Error("solana client is nil")
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return false
		}

		// Fetch transaction details
		txResponse, err := s.solanaClient.GetTransaction(ctx, signature, &rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase58,
			MaxSupportedTransactionVersion: nil,
		})
		if err != nil {
			s.logger.Error("failed to fetch Solana transaction", zap.Error(err), zap.Any("txHash", txHash))
			writeJSONError(w, fmt.Sprintf("Failed to fetch Solana transaction: %v", err), http.StatusBadRequest)
			return false
		}

		if txResponse == nil {
			s.logger.Error("transaction response is nil", zap.Any("txHash", txHash))
			writeJSONError(w, "Failed to fetch transaction details", http.StatusBadRequest)
			return false
		}

		// Validate transaction and find Wormhole messages
		acc, err := findSolanaWormholeMessage(txResponse, s.solanaContract)
		if err != nil {
			writeJSONError(w, fmt.Sprintf("Failed to parse Wormhole messages: %v", err), http.StatusBadRequest)
			return false
		} else if acc == nil {
			writeJSONError(w, "No Wormhole message found in transaction", http.StatusBadRequest)
			return false
		}

		// Send observation request
		if s.adminClient == nil {
			s.logger.Error("admin client is nil")
			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return false
		}

		_, err = s.adminClient.SendObservationRequest(ctx, &nodev1.SendObservationRequestRequest{
			ObservationRequest: &gossipv1.ObservationRequest{
				ChainId: uint32(chainID),
				TxHash:  acc[:],
			},
		})
		if err != nil {
			writeJSONError(w, fmt.Sprintf("Failed to send observation request: %v", err), http.StatusInternalServerError)
			return false
		}
	}

	return true
}

func findSolanaWormholeMessage(txResponse *rpc.GetTransactionResult, solanaContract *solana.PublicKey) (*solana.PublicKey, error) {
	acc, err := processTransaction(txResponse, solanaContract)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func processTransaction(out *rpc.GetTransactionResult, solanaContract *solana.PublicKey) (*solana.PublicKey, error) {
	program, err := solana.PublicKeyFromBase58(solanaContract.String())
	if err != nil {
		return nil, err
	}

	tx, err := out.Transaction.GetTransaction()
	if err != nil {
		return nil, err
	}

	// signature := tx.Signatures[0]
	var programIndex uint16
	for n, key := range tx.Message.AccountKeys {
		if key.Equals(program) {
			programIndex = uint16(n)
		}
	}
	if programIndex == 0 {
		return nil, nil
	}

	txs := make([]solana.CompiledInstruction, 0, len(tx.Message.Instructions))
	txs = append(txs, tx.Message.Instructions...)
	for _, inner := range out.Meta.InnerInstructions {
		txs = append(txs, inner.Instructions...)
	}

	for _, inst := range txs {
		if inst.ProgramIDIndex != programIndex {
			continue
		}
		if inst.Data[0] != postMessageInstructionID {
			continue
		}
		acc := tx.Message.AccountKeys[inst.Accounts[1]]
		return &acc, nil
	}

	return nil, nil
}
