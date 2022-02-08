package models

import (
	"context"
	"crypto-colly/common/db"
	"fmt"
	"time"
)

type NftContractModel struct {
	conn *db.Db
}

type NFTContractCollect struct {
	NftContractHash string `json:"nftcontracthash"`
	CollectionName  string `json:"collection_name"`
	CollectionSymbol string `json:"collection_symbol"`
	TransactionHash string   `json:"transaction_hash"`
	Blockchain     string  `json:"blockchain"`
	BlockHeight       uint64 `json:"block_height"`    // 发现块高
	CreateTime        int64  `json:"create_time"`
	ErcType           string `json:"erc_type"`
	LowestPrice_24h   float64 `json:"lowestPrice_24H"`
	HigestPrice_24h   float64 `json:"higestPrice_24H"`
	AveragePrice_24h float64  `json:"averagePrice_24H"`
	Volume_24h 		float64   `json:"volume_24H"`
	Volume_total	float64   `json:"volume_Total"`

}

type NFTAssetCollect struct {
	NftAssetId int64 `json:"nftAssetId"`
	NftContractHash  string `json:"nftContractHash"`
	MintTransactionHash string `json:"mintTransactionHash"`
	MintBlockHeight uint64  `json:"mintBlockHeight"`
	MintTimeStamp     uint64  `json:"mintTimeStamp"`
	Creator      string `json:"creator"`    // 发现块高
	Holder       string  `json:"holder"`
	ImageUrl           string `json:"imageUrl"`
	Format   string `json:"format"`


}

type NFTTransaction struct {
	TransactionHash string `json:"transactionHash"`
	BlockHeight  uint64 `json:"blockHeight"`
	TimeStamp uint64 `json:"timeStamp"`
	From string  `json:"from"`
	To     string  `json:"to"`
	Value      uint64 `json:"value"`    // 发现块高
	GasPrice       uint64 `json:"gasPrice"`
	GasLimit       uint64 `json:"gasLimit"`
	GasUsedByTranscation   uint64 `json:"gasUsedByTranscation"`
	TransactionFee float64 `json:"transactionFee"`


}

type NFTTransfer struct {
	NFTContractHash string `json:"NFTContractHash"`
	NFTAssetId  int64 `json:"NFTAssetId"`
	TransactionHash string `json:"transactionHash"`
	From string  `json:"from"`
	To     string  `json:"to"`
	Type string `json:"type"`


}

func NewNftModel(conn *db.Db) *NftContractModel{
	return &NftContractModel{conn: conn}
}

func (n *NftContractModel) CreateNftTransfer(nftContractHash string,nftAssetId int64,transactionHash string, from ,to string, transactionType string) error{
	data := NFTTransfer{
		NFTContractHash: nftContractHash,
		NFTAssetId: nftAssetId,
		TransactionHash: transactionHash,
		From: from,
		To: to,
		Type: transactionType,
	}

	_,err := n.conn.GetConn().Database("nftscan").Collection("nfttransfer").InsertOne(context.TODO(),data)
	if err != nil {
		fmt.Println("Insert nft error")
		return  err
	}

	return nil
}


func (n *NftContractModel) CreateNftContract(blockchain string, address, ercType, name, symbol string, blockHeight uint64, tx string) (string, error) {
	data := NFTContractCollect{
		Blockchain:  blockchain,
		NftContractHash:      address,
		ErcType:       ercType,
		CollectionName:       name,
		CollectionSymbol:     symbol,
		BlockHeight:   blockHeight,
		TransactionHash:            tx,
		CreateTime:    time.Now().Unix(),
	}

	_,err := n.conn.GetConn().Database("nftscan").Collection("nftcontract").InsertOne(context.TODO(),data)
	if err != nil {
		fmt.Println("Insert nft error")
		return "", err
	}

	return data.Blockchain, nil
}

func (n *NftContractModel) CreateNftTransaction(transactionHash string,blockHeight, timeStamp uint64,from, to string,value,gasPrice,gasLimit,gasUsedByTransaction uint64,transactionFee float64)  error {
	data := NFTTransaction{
		TransactionHash: transactionHash,
		BlockHeight: blockHeight,
		TimeStamp: timeStamp,
		From: from,
		To: to,
		Value: value,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		GasUsedByTranscation: gasUsedByTransaction,
		TransactionFee: transactionFee,
	}

	_,err := n.conn.GetConn().Database("nftscan").Collection("nfttransaction").InsertOne(context.TODO(),data)
	if err != nil {
		fmt.Println("Insert nft error")
		fmt.Println(err)
		return  err
	}

	return  nil
}

func (n *NftContractModel) CreateNftAsset(nftAssetId int64, nftContractHash, mintTransactionHash string, mintBlockHeight, mintTimeStamp uint64, creator string, ) error {
	data := NFTAssetCollect{
		NftAssetId: nftAssetId,
		NftContractHash: nftContractHash,
		MintTransactionHash: mintTransactionHash,
		MintBlockHeight: mintBlockHeight,
		MintTimeStamp: mintTimeStamp,
		Creator: creator,
		Holder : creator,
		ImageUrl: "",
		Format: "",

	}

	_,err := n.conn.GetConn().Database("nftscan").Collection("nftasset").InsertOne(context.TODO(),data)
	if err != nil {
		fmt.Println("Insert nftasset error")
		return  err
	}

	return  nil
}
