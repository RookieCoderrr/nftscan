package record

import (
	"context"
	"crypto-colly/common/chainutils"
	"crypto-colly/common/db"
	"crypto-colly/common/redis"
	"crypto-colly/contract/erc721"
	"crypto-colly/models"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const (
	ProcessBlockHeightPrefix = "nft_scan_process_block_height_"
)

type RecordBlock struct {
	chain              *models.Blockchain
	db                 *db.Db
	redis              *redis.Redis
	client             *ethclient.Client
	model              *models.NftContractModel
	processBlockHeight *big.Int
	currentBlockHeight *big.Int
	crawling           bool
	startTime          time.Time
	startPrefix        int
}
func NewRecordBlock(chain *models.Blockchain, db *db.Db, redis *redis.Redis,startPrefix int) *RecordBlock {
	return &RecordBlock{
		chain: chain,
		db: db,
		redis: redis,
		model: models.NewNftModel(db),
		processBlockHeight: new(big.Int),
		currentBlockHeight: new(big.Int),
		startTime: time.Now(),
		startPrefix: startPrefix,
	}

}
func (r *RecordBlock) Do(){
	fmt.Println("======开始记录区块======")
	client, err := ethclient.Dial(r.chain.RPC)
	if err != nil {
		output := fmt.Sprintf("连接(%s)[%s]失败: %s\n", r.chain.Name, r.chain.RPC, err)
		fmt.Println(output)
		return
	}
	r.client = client

	output := fmt.Sprintf("连接(%s)[%s]成功\n", r.chain.Name, r.chain.RPC)
	fmt.Println(output)
	lastProcessBlockHeight, err := r.getProcessedBlockHeight()
	if err != nil {
		output = fmt.Sprintf("(%s)获取上次处理块高失败: %s\n", r.chain.Name, err)
		fmt.Println(output)
		return
	}
	r.processBlockHeight = big.NewInt(7999999)
	//r.processBlockHeight = lastProcessBlockHeight
	output = fmt.Sprintf("(%s)开始爬取合约，上次处理块高: %s\n", r.chain.Name, lastProcessBlockHeight.String())
	fmt.Println(output)

	go r.autoGetCurrentBlockHeight()
	r.autoCrawl()

}
func (r *RecordBlock) getProcessedBlockHeight() (*big.Int, error) {
	var (
		blockHeight = new(big.Int)
		err         error
	)
	result, err := r.redis.Do("GET", ProcessBlockHeightPrefix+strings.ToLower(r.chain.Name)+string(r.startPrefix))
	if err != nil {
		return blockHeight, err
	}
	if result == nil {
		return  big.NewInt(int64(1000000 * r.startPrefix)),nil
	}
	blockHeight.SetString(string(result.([]byte)), 10)
	return blockHeight, nil
}

func (r *RecordBlock) autoGetCurrentBlockHeight() {
	tick := time.Tick(3 * time.Second)
	for {
		select {
		case <-tick:
			r.getCurrentBlockHeight()
		}
	}
}

func (r *RecordBlock) getCurrentBlockHeight() {
	header, err := r.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		output := fmt.Sprintf("(%s)获取当前块高失败: %s\n", r.chain.Name, err)
		fmt.Println(output)
		return
	}

	r.currentBlockHeight = header.Number
	var diff = new(big.Int).Sub(r.currentBlockHeight, r.processBlockHeight)
	if diff.Cmp(big.NewInt(10)) > 0 {
		_ = fmt.Sprintf("(%s)待处理: %s 块\n", r.chain.Name, diff.String())
		//fmt.Println(output)
	}
}

func (r *RecordBlock) autoCrawl() {
	tick := time.Tick(3 * time.Second)
	for {
		select {
		case <-tick:
			if !r.crawling && r.processBlockHeight.Cmp(r.currentBlockHeight) <= 0 && r.processBlockHeight != big.NewInt(int64((r.startPrefix+1)*1000000)){
				go r.crawl()
			}
		}
	}
}

func (r *RecordBlock) crawl() {
	r.crawling = true
	for {
		block, err := r.client.BlockByNumber(context.Background(), r.processBlockHeight)
		if err != nil {
			output := fmt.Sprintf("(%s)[%d]获取块数据失败: %s\n", r.chain.Name, r.processBlockHeight, err)
			fmt.Println(output)
			break
		}
		output := fmt.Sprintf("BlockHash: %s",block.Hash())
		fmt.Println(output)
		//fmt.Print("BlockTransactions:")
		//fmt.Println(block.Transactions())
		for _, tx := range block.Transactions() {
			output := fmt.Sprintf("Checking transaction: %s",tx.Hash())
			fmt.Println(output)
			if tx.To() == nil {
				fmt.Println("==============================================Satisfy Standard================================================")
				err := r.analyzeTx(tx)
				if err != nil {
					continue
				}
			}
			r.analyzeNftTransfer(tx,block)
		}

		err = r.saveProcessedBlockHeight(r.processBlockHeight)
		if err != nil {
			fmt.Sprintf("(%s)[%d]保存处理进度失败: %s\n", r.chain.Name, r.processBlockHeight, err)
			break
		}

		r.processBlockHeight = new(big.Int).Add(r.processBlockHeight, big.NewInt(1))
		output = fmt.Sprintf("Block height : %d",r.processBlockHeight)
		fmt.Println(output)
		if r.processBlockHeight.Cmp(r.currentBlockHeight) > 0 {
			break
		}
		if r.processBlockHeight == big.NewInt(int64(r.startPrefix*1000000)) {
			break
		}
	}
	r.crawling = false
}
func (r *RecordBlock) saveProcessedBlockHeight(blockHeight *big.Int) error {
	_, err := r.redis.Do("SET", ProcessBlockHeightPrefix+strings.ToLower(r.chain.Name)+string(r.startPrefix), blockHeight.String())
	fmt.Sprintf("Save block height: %d",blockHeight)
	return err
}

func (r *RecordBlock) analyzeTx(tx *types.Transaction) error {
	fmt.Println("analyzeTx")
	receipt, err := r.client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		output := fmt.Sprintf("(%s)[%d]解析交易失败(%s): %s\n", r.chain.Name, r.processBlockHeight, tx.Hash().Hex(), err)
		fmt.Println(output)
		return err
	}

	//tx, isPedding, err := r.client.TransactionByHash(context.Background(), tx.Hash())
	//receipt, err := bind.WaitMined(context.Background(), r.client, tx)

	if receipt.ContractAddress.Hex() != "0x0000000000000000000000000000000000000000" {
		fmt.Println("======================================================Success==============================================================")
		ercType, err := chainutils.JudgmentErcType(receipt.ContractAddress, r.client)
		if err != nil {
			output := fmt.Sprintf("(%s)[%d]判断合约类型失败(tx: %s, contract: %s): %s\n", r.chain.Name,
				r.processBlockHeight, tx.Hash().Hex(), receipt.ContractAddress.Hex(), err)
			fmt.Println(output)
			return err
		}

		switch ercType {
		case chainutils.Erc721:
			err := r.recordErc721(receipt.ContractAddress, tx.Hash().Hex())
			if err != nil {
				output := fmt.Sprintf("(%s)[%d]保存合约失败(tx: %s, contract: %s, erc_type: %s): %s\n", r.chain.Name,
					r.processBlockHeight, tx.Hash().Hex(), receipt.ContractAddress.Hex(), "erc721", err)
				fmt.Println(output)
				return err
			}
			break
		case chainutils.Unknown:
			break
		}
	}

	return nil
}

func (r *RecordBlock)analyzeNftTransfer(tx *types.Transaction, block *types.Block) error {
	receipt, err := r.client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		output := fmt.Sprintf("(%s)[%d]解析交易失败(%s): %s\n", r.chain.Name, r.processBlockHeight, tx.Hash().Hex(), err)
		fmt.Println(output)
		return err
	}
	var flag bool = false
	for _, log := range receipt.Logs{
		if checkErc721Transfer(log) {
			flag = true
			fmt.Println("====================================nft transfer success====================")
			r.recordNftTransfer(log)
			if checkType( "0x"+log.Topics[1].String()[26:], "0x"+log.Topics[2].String()[26:]) == "mint"{
				r.recordNft(log,tx,block)
			}
		}

	}
	if flag {
		r.recordNftTransaction(receipt,block,tx)
	}
	return nil

}
func (r *RecordBlock) recordNftTransfer(log *types.Log)error {
	nftContractHash := log.Address.String()
	nftAssetId := getTokenId(log.Topics[3].String())
	transactionHash := log.TxHash.String()
	from := "0x"+log.Topics[1].String()[26:]
	to := "0x"+log.Topics[2].String()[26:]
	transfertype := checkType(from,to)
	err := r.model.CreateNftTransfer(nftContractHash,int64(nftAssetId),transactionHash,from,to,transfertype)
	return err
}

func (r *RecordBlock) recordNftTransaction(receipt *types.Receipt, block *types.Block, tx *types.Transaction) error{
	transactionHash := receipt.TxHash.String()
	blockHeight := r.processBlockHeight.Uint64()
	timeStamp := block.Time()
	chainID, err := r.client.NetworkID(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	asMessage, e := tx.AsMessage(types.NewEIP155Signer(chainID),tx.GasTipCap())
	var from string
	if e == nil {
		from = asMessage.From().Hex()
	}
	to := tx.To().String()
	value := tx.Value().Uint64()
	gasPrice := tx.GasPrice().Uint64()
	gasLimit := tx.Gas()
	gasUsedByTransaction := receipt.GasUsed
	transactionFee,_ := strconv.ParseFloat(fmt.Sprintf("%.8f",  math.Pow(10,-18) *float64(gasUsedByTransaction)*float64(gasPrice)), 64)
	err = r.model.CreateNftTransaction(transactionHash,blockHeight,timeStamp,from,to,value,gasPrice,gasLimit,gasUsedByTransaction,transactionFee)
	return err

}

func (r *RecordBlock) recordNft(log *types.Log, tx *types.Transaction, block *types.Block) error {
	nftAssetId := getTokenId(log.Topics[3].String())
	mintTxHash := tx.Hash().String()
	mintBlockHeight := log.BlockNumber
	mintTimeStamp := block.Time()
	creator := "0x"+log.Topics[2].String()[26:]
	nftContractHash := log.Address.Hex()

	err := r.model.CreateNftAsset(int64(nftAssetId),nftContractHash,mintTxHash,mintBlockHeight,mintTimeStamp,creator)
	return err

}

func getTokenId(string16Hex string) uint32 {
	string16 := string16Hex[2:]
	n, err := strconv.ParseUint(string16, 16, 32)
	if err != nil {
		fmt.Println(err)
	}

	n2 := uint32(n)
	return n2
}
func checkType(from string, to string) string {
	if from == "0x0000000000000000000000000000000000000000" {
		return "mint"
	} else if to == "0x0000000000000000000000000000000000000000" {
		return "burn"
	} else {
		return "transfer"
	}
}

func checkErc721Transfer(log *types.Log) bool {
	if len(log.Topics) == 4 && log.Topics[0].Hex() == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
		return true
	} else {
		return false
	}
}

func (r *RecordBlock) recordErc721(address common.Address, tx string) error {
	addr := strings.ToLower(address.Hex())
	instance, _ := erc721.InitInstance(r.chain.RPC, address.Hex())
	name, _ := instance.Name(&bind.CallOpts{})
	symbol, _ := instance.Symbol(&bind.CallOpts{})
	_, err := r.model.CreateNftContract(r.chain.Name, addr, "erc721", name, symbol, r.processBlockHeight.Uint64(), tx)
	output := fmt.Sprintf("(%s)[%d]新收录(contract: %s, erc_type: %s, name: %s, symbol: %s)\n", r.chain.Name,
		r.processBlockHeight, addr, "erc721", name, symbol)
	fmt.Println(output)
	return err
}

