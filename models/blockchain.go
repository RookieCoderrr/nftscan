package models

type Blockchain struct {
	Name    string `json:"name"`
	ChainId int64    `json:"chain_id"`
	RPC     string `json:"rpc"`
}


func NewBlockChain(name string,chainId int64, rpc string) *Blockchain{
	return &Blockchain{
		Name: name,
		ChainId: chainId,
		RPC: rpc,
	}
}