package api

import "C"
import (
	"bytes"
	"crypto-colly/common/db"
	"crypto-colly/models"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Collection struct {
	chain *models.Blockchain
	listurl string
	collectionurl string
	itemurl string
	db *db.Db
	collectionlist []string
	itemList []string
	searchcollection bool
	searchitem bool
	collectionChan chan string
	itemChan chan string
}

func NewCollection (chain *models.Blockchain,listurl string, collectionurl string,itemurl string, db *db.Db) *Collection{
	return &Collection{
		chain: chain,
		listurl: listurl,
		collectionurl: collectionurl,
		itemurl: itemurl,
		db: db,
	}
}

func (c *Collection) Run(){
	fmt.Println("=============开始查询top 100 collection")
	c.getTopCollections()
	c.getCollectionDetail()
	c.getItem()

}

func (c *Collection)getTopCollections(){
	client := &http.Client{}
	req, err := http.NewRequest("GET",c.listurl,nil)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	resp, err := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	if gjson.Get(string(body),"success").Bool() == false {
		output := fmt.Sprintf("(%s)获取top 100 collection失败: %s\n",  c.chain.Name, err)
		fmt.Println(output)
		return
	}

	for i := 0; i < 100; i++ {
		idPath := "data.list."+strconv.Itoa(i)+".collectionId"
		collectionId := gjson.Get(string(body),idPath).String()
		//fmt.Println(collectionId)
		c.collectionlist = append(c.collectionlist, collectionId)
	}
	output := fmt.Sprintf("Get top 100 collection successfully! %s",c.collectionlist)
	fmt.Println(output)
	c.collectionChan <- "done"
}

func (c *Collection)getCollectionDetail() {
	 <-c.collectionChan
	for _, collectionId := range c.collectionlist {
		client := &http.Client{}
		requestBody := make(map[string]interface{})
		requestBody["orderBy"] = "list_time"
		requestBody["orderType"] = -1
		requestBody["page"] = 1
		requestBody["rows"] = 5
		requestBody["collectionId"] = collectionId
		bytesData, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST",c.collectionurl,bytes.NewReader(bytesData))
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		defer req.Body.Close()
		resp, err := client.Do(req)
		body, _ := ioutil.ReadAll(resp.Body)
		if gjson.Get(string(body),"success").Bool() == false {
			output := fmt.Sprintf("(%s)获取Collection detail失败: %s\n", c.chain.Name, err)
			fmt.Println(output)
			return
		}
		productId := gjson.Get(string(body),"data.rows.0.productId").String()
		c.itemList = append(c.itemList,productId)
	}
	output := fmt.Sprintf("Get 100 items successfully! %s",c.itemList)
	fmt.Println(output)
	c.itemChan <- "done"
}

func (c *Collection)getItem() {
	<- c.itemChan
	for _, item := range c.itemList {
		client := &http.Client{}
		requestBody := make(map[string]interface{})
		requestBody["productId"] = item
		bytesData, _ := json.Marshal(requestBody)
		req, err := http.NewRequest("POST",c.itemurl, bytes.NewReader(bytesData))
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
			output := fmt.Sprintf("(%s)获取productId%s失败: %s, ", c.chain.Name,item, err)
			fmt.Println(output)
			continue
		}
		//contractHash := gjson.Get(string(body),"data.nftInfo.contractAddress").String()
		//filter := bson.M{"contract":contractHash}
		//result := c.db.GetConn().Database("nft").Collection("info").FindOne(context.TODO(),filter)
		//if result.Err() != nil {
		//	output := fmt.Sprintf("Nft: %s ProductId: %s hasn't been recorded in db!",contractHash,item)
		//	fmt.Println(output)
		//} else {
		//	update := bson.M{"$set":bson.M{"ispopular":true}}
		//	_,err  := c.db.GetConn().Database("nft").Collection("info").UpdateOne(context.TODO(),filter,update)
		//	if err != nil {
		//		fmt.Println("update error!")
		//	}
		//}
		time.Sleep(time.Duration(100)*time.Millisecond)
	}

}