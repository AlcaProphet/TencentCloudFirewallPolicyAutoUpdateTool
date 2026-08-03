查询指定轻量应用服务器的防火墙规则。

## 接口说明

可您以通过此接口，查询指定轻量应用服务器的防火墙规则信息，包括端口范围、防火墙规则 ID、传输层协议等信息。

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/SWAS-OPEN/2020-06-01/ListFirewallRules)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/SWAS-OPEN/2020-06-01/ListFirewallRules)

## **授权信息**

下表是API对应的授权信息，可以在RAM权限策略语句的`Action`元素中使用，用来给RAM用户或RAM角色授予调用此API的权限。具体说明如下：

-   操作：是指具体的权限点。
    
-   访问级别：是指每个操作的访问级别，取值为写入（Write）、读取（Read）或列出（List）。
    
-   资源类型：是指操作中支持授权的资源类型。具体说明如下：
    
    -   对于必选的资源类型，用前面加 \* 表示。
        
    -   对于不支持资源级授权的操作，用`全部资源`表示。
        
-   条件关键字：是指云产品自身定义的条件关键字。
    
-   关联操作：是指成功执行操作所需要的其他权限。操作者必须同时具备关联操作的权限，操作才能成功。
    

| **操作** | **访问级别** | **资源类型** | **条件关键字** | **关联操作** |
| --- | --- | --- | --- | --- |
| swas-open:ListFirewallRules | get | FirewallRule `acs:swas-open:{#regionId}:{#accountId}:firewallrule/*` | 无   | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| InstanceId | string | 是   | 指定的轻量应用服务器的实例 ID。 | ace0706b2ac4454d984295a94213\\*\\*\\*\\* |
| RegionId | string | 是   | 指定的轻量应用服务器所属的地域 ID。 | cn-hangzhou |
| PageSize | integer | 否   | 分页查询时设置的每页行数。 最大值：100。 默认值：10。 | 10  |
| PageNumber | integer | 否   | 防火墙规则列表的页码。 起始值：1。 默认值：1。 | 1   |
| Tag | array<object> | 否   | 防火墙规则的标签列表。 |     |
|     | object | 否   | 标签。 |     |
| Key | string | 否   | 防火墙规则的标签键。标签键长度的取值范围：1~64。N 的取值范围：1~20。 | TestKey |
| Value | string | 否   | 防火墙规则的标签值。标签值长度的取值范围：1~64。N 的取值范围：1~20。 | TestValue |
| FirewallRuleId | string | 否   | 防火墙规则 ID。 | 1a16263ab0f541288312a15fa64280de |

## **返回参数**

| **名称** | **类型** | **描述** | **示例值** |
| --- | --- | --- | --- |
|     | object |     |     |
| TotalCount | integer | 防火墙规则的总条数。 | 4   |
| RequestId | string | 请求 ID。 | 20758A-585D-4A41-A9B2-28DA8F4F534F |
| PageSize | integer | 分页查询时设置的每页行数。 | 10  |
| PageNumber | integer | 防火墙规则列表的页码。 | 1   |
| FirewallRules | array<object> | 由防火墙规则信息组成的数组。 |     |
|     | array<object> |     |     |
| Remark | string | 防火墙规则的备注。 | test-MySQL服务器默认端口 |
| Port | string | 端口范围。 | 3306 |
| RuleId | string | 防火墙规则 ID。 | eeea34d9867b4d55a4ff8d5fcfbd\\*\\*\\*\\* |
| RuleProtocol | string | 传输层协议。可能值： - TCP：TCP 协议。 - UDP：UDP 协议。 - TCP+UDP：TCP 和 UDP 协议。 | TCP |
| Policy | string | 防火墙策略。 - accept：允许。 - drop：拒绝。 | accept |
| Tags | array<object> | 防火墙规则的标签列表。 |     |
|     | object | 标签。 |     |
| Key | string | 防火墙规则的标签键。 | TestKey |
| Value | string | 防火墙规则的标签值。 | TestValue |
| SourceCidrIp | string | 源 IP 段。 | 0.0.0.0/0 |

## 示例

正常返回示例

`JSON`格式

```
{
  "TotalCount": 4,
  "RequestId": "20758A-585D-4A41-A9B2-28DA8F4F534F",
  "PageSize": 10,
  "PageNumber": 1,
  "FirewallRules": [
    {
      "Remark": "test-MySQL服务器默认端口",
      "Port": "3306",
      "RuleId": "eeea34d9867b4d55a4ff8d5fcfbd****",
      "RuleProtocol": "TCP",
      "Policy": "accept",
      "Tags": [
        {
          "Key": "TestKey",
          "Value": "TestValue"
        }
      ],
      "SourceCidrIp": "0.0.0.0/0"
    }
  ]
}
```

## 错误码

| **HTTP status code** | **错误码** | **错误信息** | **描述** |
| --- | --- | --- | --- |
| 400 | RegionIdNotMatchHost | The parameter regionId does not match the endpoint host. | 传入的regionId和请求的地址不匹配。 |
| 500 | InternalError | An error occurred while processing your request. | 内部错误，请重试。如果多次尝试失败，请提交工单。 |
| 403 | InvalidParam | The specified parameter value is invalid. | 参数非法。 |
| 404 | InvalidInstanceId.NotFound | The specified InstanceId does not exist. | 指定的实例不存在，请您检查实例ID是否正确。 |

访问[错误中心](https://api.aliyun.com/document/SWAS-OPEN/2020-06-01/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/SWAS-OPEN/2020-06-01/ListFirewallRules#workbench-doc-change-demo)。