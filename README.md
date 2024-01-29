# go_aliddns
go_aliddns是一个基于阿里云API的动态域名解析工具，支持多个域名解析记录的更新，支持微信推送更新结果。

## 配置文件说明：
1. 将.env.example改为.env
2. 修改.env文件中的配置
3.
    aliAccessKeyId与aliAccessKeySecret的获取： https://usercenter.console.aliyun.com/#/manage/ak <br>
    notifyKey:是微信推送的KEY，见https://sct.ftqq.com/<br>
    domainName:是你的一级域名<br>
    subDomainName:是你要设置更新的二级域名<br>

```
aliAccessKeyId=`阿里云accessKeyId`
aliAccessKeySecret=`阿里云accessKeySecret`
domainName=`是你的一级域名`
subDomainName=`你要设置更新的二级域名`
notifyKey=微信推送的KEY
```





