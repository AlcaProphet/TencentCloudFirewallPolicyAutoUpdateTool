概述
文档中 SDK 关于 API 的示例代码仅供参考，各 API 的完整使用步骤与说明请参见SDK 示例 和 OpenAPI 文档。
Latest Stable Version

环境要求
Go 环境版本必须不低于 1.10.x
安装 SDK 核心库 OpenAPI
go get github.com/alibabacloud-go/darabonba-openapi/v2/client
发布地址
https://github.com/alibabacloud-go/swas-open-20200601/
源码仓库地址
https://github.com/alibabacloud-go/swas-open-20200601/
安装方式
Go Get
go get github.com/alibabacloud-go/swas-open-20200601/v3
示例背景
以下代码详细介绍了阿里云 V2 Go SDK 的使用步骤，仅作步骤示范。示例展示了如何调用 CreateFirewallRule API 进行创建实例的防火墙规则请求。
完整代码示例
以下为基于 Go SDK 提供的示例代码。

package main

import (
  "encoding/json"
  "strings"
  "fmt"
  "os"
  swas_open20200601  "github.com/alibabacloud-go/swas-open-20200601/v3/client"
  openapi  "github.com/alibabacloud-go/darabonba-openapi/v2/client"
  util  "github.com/alibabacloud-go/tea-utils/v2/service"
  credential  "github.com/aliyun/credentials-go/credentials"
  "github.com/alibabacloud-go/tea/tea"
)


// Description:
// 
// 使用凭据初始化账号 Client
// 
// @return Client
// 
// @throws Exception
func CreateClient () (_result *swas_open20200601.Client, _err error) {
  // 工程代码建议使用更安全的无 AK 方式，凭据配置方式请参见：https://help.aliyun.com/document_detail/378661.html。
  credential, _err := credential.NewCredential(nil)
  if _err != nil {
    return _result, _err
  }

  config := &openapi.Config{
    Credential: credential,
  }
  // Endpoint 请参考 https://api.aliyun.com/product/SWAS-OPEN
  config.Endpoint = tea.String("swas.cn-hangzhou.aliyuncs.com")
  _result = &swas_open20200601.Client{}
  _result, _err = swas_open20200601.NewClient(config)
  return _result, _err
}

func _main (args []*string) (_err error) {
  client, _err := CreateClient()
  if _err != nil {
    return _err
  }

  createFirewallRuleRequest := &swas_open20200601.CreateFirewallRuleRequest{
    InstanceId: tea.String("your_value"),
    RegionId: tea.String("your_value"),
  }
  tryErr := func()(_e error) {
    defer func() {
      if r := tea.Recover(recover()); r != nil {
        _e = r
      }
    }()
    resp, _err := client.CreateFirewallRuleWithOptions(createFirewallRuleRequest, &util.RuntimeOptions{})
    if _err != nil {
      return _err
    }

    fmt.Printf("[LOG] %v\n", resp)

    return nil
  }()

  if tryErr != nil {
    var error = &tea.SDKError{}
    if _t, ok := tryErr.(*tea.SDKError); ok {
      error = _t
    } else {
      error.Message = tea.String(tryErr.Error())
    }
    // 此处仅做打印展示，请谨慎对待异常处理，在工程项目中切勿直接忽略异常。
    // 错误 message
    fmt.Println(tea.StringValue(error.Message))
    // 诊断地址
    var data interface{}
    d := json.NewDecoder(strings.NewReader(tea.StringValue(error.Data)))
    d.Decode(&data)
    if m, ok := data.(map[string]interface{}); ok {
      recommend, _ := m["Recommend"]
      fmt.Println(recommend)
    }
  }
  return _err
}


func main() {
  err := _main(tea.StringSlice(os.Args[1:]))
  if err != nil {
    panic(err)
  }
}

步骤介绍
您需要在代码中引入依赖包：
import (
  swas_open "github.com/alibabacloud-go/swas-open-20200601/v3/client"
  openapi "github.com/alibabacloud-go/darabonba-openapi/client"
)
初始化配置对象 &openapi.Config。 Config 对象存放 Credential、Endpoint 等配置，Endpoint 如示例中的 swas.cn-hangzhou.aliyuncs.com 。

注意：工程代码建议使用更安全的无 AK 方式

先实例化 credential 对象，设置到 Config 中的 Credential 字段，阿里云 SDK 将会尝试按照默认凭据链的顺序查找相关凭据信息。

更多凭据配置方式请参阅：管理访问凭据

credential, _err := credential.NewCredential(nil)
if _err != nil {
  return _result, _err
}
config := &openapi.Config{
  Credential: credential
}
// 访问的域名
config.Endpoint = tea.String("swas.cn-hangzhou.aliyuncs.com")
实例化一个客户端，从 &swas_open.Client 类生成对象 client 。

client, _ := swas_open.NewClient(config)
创建对应 API 的 Request。 方法的命名规则为 API 方法名再加上 Request 。例如：

request := &swas_open.CreateFirewallRuleRequest{}
设置请求 request 的参数。 通过 request 设置必要的信息，即 API 中必须要提供的信息。


// 该参数值为假设值，请您根据实际情况进行填写
request.SetInstanceId("your_value")

// 该参数值为假设值，请您根据实际情况进行填写
request.SetRegionId("your_value")

通过 client 对象获得对应 request 响应 response 。

response, err := client.CreateFirewallRule(request)
fmt.Println(response)
通过 err 可以处理请求报错。

if err != nil {
  fmt.Println(err)
  return
}
调用 response 获得返回的参数值。 假设您需要获取 requestId ：

fmt.Println(tea.StringValue(response.Body.ImageId))
根据 API 方法的不同，返回的信息中可能会包含多层的信息。