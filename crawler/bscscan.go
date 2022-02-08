package crawler

import (
	"context"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type NFTInfo struct {
	ContractHash string
	Symbol string
	TotalSupply float64
	TotalHolders float64

}

type Config struct {
	Mongo_Local struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Pass     string `yaml:"pass"`
		Database string `yaml:"database"`
		DBName   string `yaml:"dbname"`
	} `yaml:"mongo_local"`
}

var (
	pageReg = regexp.MustCompile("tokens-nft")
	cfg,_  = OpenConfigFile()
	ctx = context.TODO()
)

const  bscUrl = "https://bscscan.com/tokens-nft"

func run () {
	//result := make([]NFTInfo,5000)
	co := initializeMongoLocalClient(ctx,cfg)
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
		colly.MaxDepth(2),
		colly.Debugger(&debug.LogDebugger{}),
		)
	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
	})
	detailCollector := c.Clone()
	//rp, err := proxy.RoundRobinProxySwitcher("socks5://127.0.0.1:1081")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//c.SetProxyFunc(rp)
	c.OnHTML("body", func(e *colly.HTMLElement){
		fmt.Println("body")
		e.ForEach("table tr", func(_ int, el *colly.HTMLElement){
			//tokenName := el.ChildText("td:nth-child(2) h3 div a")
			//transfers24h := el.ChildText("td:nth-child(3)")
			//transfers2d := el.ChildText("td:nth-child(4)")
			//fmt.Println(tokenName)
			//fmt.Println(transfers24h)
			//fmt.Println(transfers2d)
			detailURL := "https://bscscan.com"+el.ChildAttr("td:nth-child(2) h3 div a","href")
			detailCollector.Visit(detailURL)

		})
		nextPage := e.ChildAttr("div[id=ContentPlaceHolder1_divPagination] ul li:nth-child(4) a", "href")
		fmt.Println("nextPage:", nextPage)
		if pageReg.MatchString(nextPage){
			c.Visit(e.Request.AbsoluteURL(nextPage))
		}
	})

	detailCollector.OnHTML("div[id=ContentPlaceHolder1_divSummary]>div:nth-child(1)", func(e *colly.HTMLElement){
		totalSupply, err := strconv.ParseFloat(strings.ReplaceAll(e.ChildText("div:nth-child(1)>div>div:nth-child(2)>div:nth-child(1)>div:nth-child(2)>span:nth-of-type(1) "),",",""),64)
		if err != nil {
			fmt.Println("convert totalSupply error")
		}
		symbol := e.ChildText("div:nth-child(1)>div>div:nth-child(2)>div:nth-child(1)>div:nth-child(2)>b")
		holders,err:= strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(e.ChildText("div[id=ContentPlaceHolder1_tr_tokenHolders] div div div div ")," addresses",""),",",""),64)
		if err != nil {
			fmt.Println("convert holders error")
		}
		contractHash := e.ChildText("div:nth-child(2)>div>div:nth-child(2)>div:nth-child(1)>div:nth-child(2) a")
		fmt.Println(totalSupply)
		fmt.Println(symbol)
		fmt.Println(holders)
		fmt.Println(contractHash)
		raw := NFTInfo{
			ContractHash :contractHash,
			Symbol :symbol,
			TotalSupply :totalSupply,
			TotalHolders :holders,
		}
		fmt.Println("raw:",raw)
		insertOne, err := co.Database("crypto").Collection("token").InsertOne(ctx,raw)
		if err != nil {
			fmt.Println("Insert Error")
		}
		fmt.Println(insertOne)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.Visit(fmt.Sprintf("%s?p=%d", bscUrl, 1))
}

func initializeMongoLocalClient( ctx context.Context, cfg Config) *mongo.Client {
	var clientOptions *options.ClientOptions
	clientOptions = options.Client().ApplyURI("mongodb://" + cfg.Mongo_Local.Host + ":" + cfg.Mongo_Local.Port + "/" + cfg.Mongo_Local.Database)
	cl, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("connect mongo error")
	}
	err = cl.Ping(ctx, nil)
	if err != nil {
		fmt.Println("ping mongo error")
	}
	return cl
}
func OpenConfigFile() (Config, error) {
	absPath, _ := filepath.Abs("setting.yml")
	f, err := os.Open(absPath)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, err
}