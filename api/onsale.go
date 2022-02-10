package api

import (
	"bytes"
	"crypto-colly/common/redis"
	"crypto-colly/models"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	ProcessProductIdPrefix = "process_product_id_"
)

type Api struct {
	chain              *models.Blockchain
	detailurl			string
	listurl				string
	db 					*gorm.DB
	redis              *redis.Redis
	model              *models.NftContractModel
	processProductId *big.Int
	currentProductId *big.Int
	crawling           bool
	startTime          time.Time
}

func NewApi(chain *models.Blockchain,detailurl string,listurl string, db *gorm.DB, redis *redis.Redis) *Api {
	return &Api{
		chain: chain,
		detailurl: detailurl,
		listurl: listurl,
		db: db,
		redis: redis,
		model: models.NewNftModel(db),
		processProductId: new(big.Int),
		currentProductId: new(big.Int),
		startTime: time.Now(),
	}
}

func (a *Api) Run() {
	fmt.Println("======开始查询BSC MarketPlace======")
	lastProcessProductId, err := a.getProcessedProductId()
	if err != nil {
		output := fmt.Sprintf("(%s)获取上次处理productId失败: %s\n", a.chain.Name, err)
		fmt.Println(output)
		return
	}
	a.processProductId = lastProcessProductId
	//a.processProductId = big.NewInt(22843297)
	output := fmt.Sprintf("(%s)开始查询market，上次处理productId: %s\n", a.chain.Name, lastProcessProductId.String())
	fmt.Println(output)

	go a.autoGetCurrentProductId()
	a.autoCrawl()


}
func (a *Api)getProcessedProductId()(*big.Int,error){
	var (
		productId = new(big.Int)
		err error
	)
	result, err :=  a.redis.Do("GET", ProcessProductIdPrefix+strings.ToLower(a.chain.Name))
	if err != nil {
		return productId,err
	}
	if result == nil {
		return productId, nil
	}
	productId.SetString(string(result.([]byte)), 10)
	return productId,nil
}

func (a *Api)autoGetCurrentProductId(){
	tick := time.Tick(15 * time.Second)
	for {
		select {
		case <- tick:
			a.getCurrentProductId()
		}
	}
}
func (a *Api)getCurrentProductId(){
	client := &http.Client{}
	requestBody := make(map[string]interface{})
	requestBody["orderBy"] = "list_time"
	requestBody["orderType"] = -1
	requestBody["page"] = 1
	requestBody["rows"] = 5
	bytesData, _ := json.Marshal(requestBody)
	req, err := http.NewRequest("POST",a.listurl, bytes.NewReader(bytesData))
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	defer req.Body.Close()
	resp, err := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	if gjson.Get(string(body),"success").Bool() == false {
		output := fmt.Sprintf("(%s)获取当前productId失败: %s\n", a.chain.Name, err)
		fmt.Println(output)
		return
	}
	productId := big.NewInt(gjson.Get(string(body),"data.rows.0.productId").Int())
	output := fmt.Sprintf("Current productId: %s",productId)
	fmt.Println(output)

	a.currentProductId = productId
	var diff = new(big.Int).Sub(a.currentProductId,a.processProductId)
	if diff.Cmp(big.NewInt(10)) > 0 {
		output := fmt.Sprintf("(%s)待处理产品: %s 个\n", a.chain.Name, diff.String())
		fmt.Println(output)
	}
}

func (a *Api)autoCrawl(){
	tick := time.Tick(3 * time.Second)
	for {
		select {
		case <-tick:
			if !a.crawling && a.processProductId.Cmp(a.currentProductId) <= 0 {
				go a.crawl()
			}
		}
	}
}

func (a *Api)crawl(){
	a.crawling = true
	for {
		client := &http.Client{}
		requestBody := make(map[string]interface{})
		requestBody["productId"] = a.processProductId
		bytesData, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST",a.detailurl, bytes.NewReader(bytesData))
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		defer req.Body.Close()
		resp, err := client.Do(req)
		body, _ := ioutil.ReadAll(resp.Body)
		//fmt.Println(string(body))
		if gjson.Get(string(body),"success").Bool() == false {
			output := fmt.Sprintf("(%s)获取当前productId失败: %s\n", a.chain.Name, err)
			fmt.Println(output)
			continue
		}

		//contractHash := gjson.Get(string(body),"data.nftInfo.contractAddress").String()
		//filter := bson.M{"contract":contractHash}
		//result := a.db.GetConn().Database("nft").Collection("info").FindOne(context.TODO(),filter)
		//if result.Err() != nil {
		//	output := fmt.Sprintf("Nft: %s ProductId: %s hasn't been recorded in db!",contractHash,a.processProductId)
		//	fmt.Println(output)
		//} else {
		//	update := bson.M{"$set": bson.M{"$inc": bson.M{"marketplace": 1}}}
		//	_,err  := a.db.GetConn().Database("nft").Collection("info").UpdateOne(context.TODO(),filter,update)
		//	if err != nil {
		//		fmt.Println("update error!")
		//	}
		//}
		err = a.saveProcessedProductId(a.processProductId)
		if err != nil {
			fmt.Sprintf("(%s)[%d]保存处理进度失败: %s\n", a.chain.Name, a.processProductId, err)
			break
		}
		a.processProductId = new(big.Int).Add(a.processProductId,big.NewInt(1))
		output := fmt.Sprintf("productId : %d",a.processProductId)
		fmt.Println(output)
		if a.processProductId.Cmp(a.currentProductId) > 0 {
			break
		}
		time.Sleep(time.Duration(100)*time.Millisecond)
	}
	a.crawling = false
}

func (a *Api) saveProcessedProductId(productId *big.Int) error{
	_, err := a.redis.Do("SET", ProcessProductIdPrefix+strings.ToLower(a.chain.Name), productId.String())
	fmt.Sprintf("Save block height: %d",productId)
	return err
}
