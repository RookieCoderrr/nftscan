package chainutils

import (
	"crypto-colly/contract/erc721"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

const (
	Erc721 = iota
	Erc1155
	Unknown
)

func JudgmentErcType(contract common.Address, backend bind.ContractBackend) (int, error) {
	instance, err := erc721.NewErc721(contract, backend)
	if err != nil {
		return 0, err
	}

	if ok, err := instance.SupportsInterface(&bind.CallOpts{}, [4]byte{0x80, 0xac, 0x58, 0xcd}); err == nil && ok {
		return Erc721, nil
	}

	if ok, err := instance.SupportsInterface(&bind.CallOpts{}, [4]byte{0xd9, 0xb6, 0x7a, 0x26}); err == nil && ok {
		return Erc1155, nil
	}

	return Unknown, nil
}
