package btc

import (
	"math/big"
	"time"
	"xnode/intents/basechain"
	"xnode/nodeutil"

	"github.com/imroc/req"
)

// BTC header struct. SUBJECT TO CHANGE
type BtcHeader struct {
	Version    uint32
	PrevBlock  [32]byte
	MerkleRoot [32]byte
	Timestamp  uint32
	Bits       uint32
	Nonce      uint32
}

// Get pending txids via JSON-RPC

type getRawMempoolJsonRequest struct {
	Jsonrpc string            `json:"jsonrpc"`
	Id      string            `json:"id"`
	Method  string            `json:"method"`
	Params  getRawMempoolArgs `json:"params"`
}

type getRawMempoolArgs struct {
	Verbose         bool `json:"verbose"`
	MempoolSequence bool `json:"mempool_sequence"`
}

type getRawMempoolResponse []string

// Get ONE pending tx by txid

type getTxOutJsonRequest struct {
	Jsonrpc string       `json:"jsonrpc"`
	Id      string       `json:"id"`
	Method  string       `json:"method"`
	Params  getTxOutArgs `json:"params"`
}

type getTxOutArgs struct {
	Txid           string `json:"txid"`
	VoutNumber     string `json:"n"` // just put 1, apparently
	IncludeMempool string `json:"include_mempool"`
}

type getTxOutResponse struct {
	Confirmations int          `json:"confirmations"`
	Value         int          `json:"value"`
	ScriptPubKey  scriptPubKey `json:"scriptPubKey"`
}

// Updates pending transactions every second
func JsonRpcListenMempool(processor *BtcProcessor) {
	for {
		// do stuff here

		time.Sleep(1 * time.Second)
	}
}

// Get block headers via JSON-RPC

type getBlockHeaderJsonRequest struct {
	Jsonrpc string             `json:"jsonrpc"`
	Id      string             `json:"id"`
	Method  string             `json:"method"`
	Params  getBlockHeaderArgs `json:"params"`
}

type getBlockHeaderArgs struct {
	Hash    string `json:"hash"`
	Verbose bool   `json:"verbose"`
}

type getBlockHeaderResponse struct {
	Result getBlockHeaderResult `json:"result"`
	Error  interface{}          `json:"error"`
	Id     interface{}          `json:"id"`
}

type getBlockHeaderResult struct {
	Hash              string  `json:"hash"`
	Confirmations     int     `json:"confirmations"`
	Height            int     `json:"height"`
	Version           int     `json:"version"`
	VersionHex        string  `json:"versionHex"`
	Merkleroot        string  `json:"merkleRoot"`
	Time              int     `json:"time"`
	MedianTime        int     `json:"medianTime"`
	Nonce             int     `json:"nonce"`
	Bits              string  `json:"bits"`
	Difficulty        float32 `json:"difficulty"`
	Chainwork         string  `json:"chainwork"`
	Ntx               int     `json:"nTx"`
	Previousblockhash string  `json:"previousblockhash"`
	Nextblockhash     string  `json:"nextblock"`
}

// Get full blocks JSON-RPC

type getBlockJsonRequest struct {
	Jsonrpc string       `json:"jsonrpc"`
	Id      string       `json:"id"`
	Method  string       `json:"method"`
	Params  getBlockArgs `json:"params"`
}

type getBlockArgs struct {
	Blockhash string `json:"blockhash"`
	Verbosity int    `json:"verbosity"`
}

type getBlockResponse struct {
	Result getBlockResult `json:"result"`
}

type getBlockResult struct {
	Version           int     `json:"version"`
	Height            int     `json:"height"`
	Confirmations     int     `json:"confirmations"`
	Merkleroot        string  `json:"merkleroot"`
	Tx                []tx    `json:"tx"`
	Time              int     `json:"time"`
	Nonce             int     `json:"nonce"`
	Difficulty        float32 `json:"difficulty"`
	Previousblockhash string  `json:"previousblockhash"`
}

type tx struct {
	Txid string `json:"txid"`
	Hash string `json:"hash"`
	Vout []vout `json:"vout"`
}

type vout struct {
	Value        float32      `json:"value"`
	ScriptPubKey scriptPubKey `json:"scriptPubKey"`
}

type scriptPubKey struct {
	Hex       string   `json:"hex"`
	Addresses []string `json:"addresses"`
}

// Updates block data every new block
func JsonRpcListenBlocks(processor *BtcProcessor) {
	for {

		blockHash := <-processor.NewBlockEvents

		resultGb := getBlockResponse{}

		optsGb := getBlockArgs{Blockhash: blockHash, Verbosity: 2}
		payloadGb := getBlockJsonRequest{Jsonrpc: "1.0", Id: "xyz", Method: "getblock", Params: optsGb}
		respGb, err := req.Post("http://user:pass@127.0.0.1:5003", req.BodyJSON(&payloadGb))

		if err != nil {
			panic(err)
		}

		respGb.ToJSON(&resultGb)

		resultGbh := getBlockHeaderResponse{}
		optsGbh := getBlockHeaderArgs{Hash: blockHash, Verbose: true}
		payloadGbh := getBlockHeaderJsonRequest{Jsonrpc: "1.0", Id: "xyz", Method: "getblockheader", Params: optsGbh}
		respGbh, err := req.Post("http://user:pass@127.0.0.1:5003", req.BodyJSON(&payloadGbh))

		if err != nil {
			panic(err)
		}

		respGbh.ToJSON(&resultGbh)

		newBlock := basechain.Block{}
		for _, tx := range resultGb.Result.Tx {

			addresses := tx.Vout[0].ScriptPubKey.Addresses
			addr := ""
			if len(addresses) != 0 {
				addr = addresses[0]
			}

			newTx := basechain.Tx{
				Txid:   tx.Txid,
				Amount: big.NewFloat((float64)(tx.Vout[0].Value)),
				To:     addr,
			}
			newBlock.Txs = append(newBlock.Txs, newTx)
		}

		rgbh := resultGbh.Result
		newHeader := BtcHeader{
			Version:    (uint32)(rgbh.Version),
			PrevBlock:  nodeutil.StringToByte32(rgbh.Previousblockhash),
			MerkleRoot: nodeutil.StringToByte32(rgbh.Merkleroot),
			Timestamp:  (uint32)(rgbh.Time),
			Bits:       nodeutil.HexStringToUint32(rgbh.Bits),
			Nonce:      (uint32)(rgbh.Nonce),
		}

		processor.Chain.NewBlock(newBlock)
		processor.Chain.NewHeader(newHeader)
	}

}
