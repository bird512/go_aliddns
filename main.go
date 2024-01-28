package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type DomainDescribe struct {
	DomainRecords struct {
		Record []Record `json:"Record"`
	}
}

type Record struct {
	RR       string `json:"RR"`
	Type     string `json:"Type"`
	Value    string `json:"Value"`
	RecordId string `json:"RecordId"`
}

var subDomainName, accessKeyId, accessKeySecret, notifyKey string

var customLogger = log.New(os.Stdout, "CUSTOM: ", log.Ldate|log.Ltime)

const RegionId = "cn-hangzhou" // Replace with your region ID
func main() {
	customLogger.Println("开始执行")
	// 加载.env文件
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	accessKeyId = os.Getenv("aliAccessKeyId")
	accessKeySecret = os.Getenv("aliAccessKeySecret")
	subDomainName = os.Getenv("subDomainName")
	notifyKey = os.Getenv("notifyKey")

	//REGION_ID := "cn-hangzhou" // Replace with your region ID
	client, err := sdk.NewClientWithAccessKey(RegionId, accessKeyId, accessKeySecret)
	if err != nil {
		// Handle exceptions
		panic(err)
	}

	doJob := func() {
		customLogger.Println("执行定时任务，每1分钟一次")
		record := getRecordId(client)
		ip, err := getCurrentIp()
		if err != nil {
			customLogger.Println(err)
			return
		}
		updateRecord(client, record, ip)
	}
	doJob()
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			doJob()
		}
	}
}

func getCurrentIp() (string, error) {
	ip := getCurrentIp1()
	if isIpv4(ip) {
		return ip, nil
	} else {
		ip = getCurrentIp2()
		if isIpv4(ip) {
			return ip, nil
		}
	}
	return "nil", errors.New("ip is not ipv4")
}

// get record id
func getRecordId(client *sdk.Client) *Record {
	request := requests.NewCommonRequest()                // 构造一个公共请求
	request.Method = "POST"                               // 设置请求方式
	request.Product = "Dns"                               // 指定产品
	request.Domain = "alidns.aliyuncs.com"                // 指定域名则不会寻址，如认证方式为 Bearer Token 的服务则需要指定
	request.Version = "2015-01-09"                        // 指定产品版本
	request.ApiName = "DescribeDomainRecords"             // 指定接口名
	request.QueryParams["DomainName"] = "ctofenglei.top"  // 设置参数值
	request.QueryParams["RegionId"] = RegionId            // 指定请求的区域，不指定则使用客户端区域、默认区域
	request.TransToAcsRequest()                           // 把公共请求转化为acs请求
	response, err := client.ProcessCommonRequest(request) // 发起请求并处理异常

	if err != nil {
		// Handle exceptions
		panic(err)
	}
	var result DomainDescribe
	err = json.Unmarshal([]byte(response.GetHttpContentString()), &result)
	if err != nil {
		return nil
	}
	for _, v := range result.DomainRecords.Record {
		if v.RR == subDomainName {
			customLogger.Println(v.Value)
			return &v
		}
	}
	return nil
}

// update the dns record
func updateRecord(client *sdk.Client, record *Record, newIp string) {
	customLogger.Println("当前的公网IP地址是:", newIp)
	if newIp == record.Value {
		customLogger.Println("ip未发生变化")
		return
	}

	request := requests.NewCommonRequest() // 构造一个公共请求
	request.Method = "POST"                // 设置请求方式
	request.Product = "Dns"                // 指定产品
	request.Domain = "alidns.aliyuncs.com" // 指定域名则不会寻址，如认证方式为 Bearer Token 的服务则需要指定
	request.Version = "2015-01-09"         // 指定产品版本
	request.ApiName = "UpdateDomainRecord" // 指定接口名
	request.QueryParams["DomainName"] = "ctofenglei.top"
	request.QueryParams["RecordId"] = record.RecordId
	request.QueryParams["RR"] = record.RR
	request.QueryParams["Type"] = record.Type
	request.QueryParams["Value"] = newIp
	request.TransToAcsRequest()                           // 把公共请求转化为acs请求
	response, err := client.ProcessCommonRequest(request) // 发起请求并处理异常

	if err != nil {
		// Handle exceptions
		customLogger.Println("更新域名解析失败:", err)
		return
	}
	sendNotify("更新域名解析成功", response.GetHttpContentString())
	customLogger.Println("更新域名解析成功:", response.GetHttpContentString())
}

// get current ip
func getCurrentIp1() string {
	resp, err := http.Get("https://api64.ipify.org?format=text")
	if err != nil {
		customLogger.Println("无法获取公网IP:", err)
		return ""
	}
	defer resp.Body.Close()

	// 读取响应的内容（即公网IP地址）
	ipAddress, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		customLogger.Println("无法读取响应:", err)
		return ""
	}

	// 显示公网IP地址

	return string(ipAddress)
}
func getCurrentIp2() string {
	resp, err := http.Get("https://ip.cn/api/index?ip=&type=0")
	if err != nil {
		customLogger.Println("无法获取公网IP:", err)
		return ""
	}
	defer resp.Body.Close()

	ipAddress, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		customLogger.Println("无法读取响应:", err)
		return ""
	}
	var result map[string]interface{}
	err = json.Unmarshal([]byte(ipAddress), &result)
	if err != nil {
		return ""
	}
	return result["ip"].(string)
}

// check if string is an ipv4 address
func isIpv4(ipString string) bool {
	// check if string is an ipv4 address
	ip := net.ParseIP(ipString)
	if ip == nil {
		customLogger.Printf("%s 不是一个有效的IPv4地址\n", ipString)
		return false
	} else if ip.To4() != nil {
		return true
	} else {
		customLogger.Printf("%s 不是一个IPv4地址\n", ipString)
		return false
	}
}

func sendNotify(text string, desp string) string {
	data := url.Values{}
	data.Set("text", text)
	data.Set("desp", desp)

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://sctapi.ftqq.com/%s.send", notifyKey), strings.NewReader(data.Encode()))
	if err != nil {
		return err.Error()
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err.Error()
	}
	customLogger.Println(string(body))
	return string(body)
}
