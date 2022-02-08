package erc721

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ContractNFT721Abi abi.ABI
)

func InitInstance(uri string, contract string) (*Erc721, error) {
	var err error
	client, err := ethclient.Dial(uri)
	if err != nil {
		return nil, err
	}
	contractAddress := common.HexToAddress(contract)
	instance, err := NewErc721(contractAddress, client)
	if err != nil {
		return nil, err
	}
	ContractNFT721Abi, err = abi.JSON(strings.NewReader(string(Erc721ABI)))
	if err != nil {
		return nil, err
	}
	return instance, err
}

func GetInstance(uri string, contract string) (*Erc721, error) {
	var err error
	client, err := ethclient.Dial(uri)
	if err != nil {
		return nil, err
	}
	contractAddress := common.HexToAddress(contract)
	instance, err := NewErc721(contractAddress, client)
	if err != nil {
		return nil, err
	}
	return instance, err
}
