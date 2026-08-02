修改指定轻量应用服务器的防火墙规则。

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/SWAS-OPEN/2020-06-01/ModifyFirewallRule)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/SWAS-OPEN/2020-06-01/ModifyFirewallRule)

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
| swas-open:ModifyFirewallRule | update | \\*FirewallRule `acs:swas-open:{#regionId}:{#accountId}:firewallrule/{#RuleId}` | 无   | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| InstanceId | string | 是   | 指定的轻量应用服务器的实例 ID。 | ace0706b2ac4454d984295a94213\\*\\*\\*\\* |
| RegionId | string | 是   | 指定的轻量应用服务器实例所属的地域 ID。 您可以调用 [ListRegions](https://help.aliyun.com/zh/simple-application-server/api-listregions) 查看支持的阿里云地域列表。 | cn-hangzhou |
| ClientToken | string | 否   | 保证请求幂等性。从您的客户端生成一个参数值，确保不同请求间该参数值唯一。**ClientToken** 只支持 ASCII 字符，且不能超过 64 个字符。更多信息，请参见[如何保证幂等性](https://help.aliyun.com/zh/ecs/developer-reference/how-to-ensure-idempotence)。 | 123e4567-e89b-12d3-a456-426655440000 |
| RuleId | string | 是   | 防火墙规则 ID。 | eeea34d9867b4d55a4ff8d5fcfbd\\*\\*\\*\\* |
| RuleProtocol | string | 是   | 传输层协议。取值范围： - TCP：TCP 协议。 - UDP：UDP 协议。 - TCP+UDP：TCP 和 UDP 协议。 | TCP |
| Port | string | 是   | 端口范围。端口的取值范围为 1~65535。使用正斜线（/）隔开起始端口和终止端口，例如：`1024/1055`表示端口范围为 1024~1055。 | 3306 |
| SourceCidrIp | string | 否   | 地址或地址段。 | 10.147.33.\\*\\* |
| Remark | string | 否   | 防火墙规则的备注。 | 自定义 |

## **返回参数**

| **名称** | **类型** | **描述** | **示例值** |
| --- | --- | --- | --- |
|     | object |     |     |
| RequestId | string | 请求 ID。 | 30637AD6-D977-4833-A54C-CC89483E1FEE |

## 示例

正常返回示例

`JSON`格式

```
{
  "RequestId": "30637AD6-D977-4833-A54C-CC89483E1FEE"
}
```

## 错误码

| **HTTP status code** | **错误码** | **错误信息** | **描述** |
| --- | --- | --- | --- |
| 400 | InvalidProtocol.ValueNotSupported | The specified parameter Protocol is invalid. | 指定的防火墙协议不合法。 |
| 400 | InvalidPort.ValueNotSupported | The specified parameter Port is invalid. | 指定的防火墙端口范围不合法。 |
| 400 | RegionIdNotMatchHost | The parameter regionId does not match the endpoint host. | 传入的regionId和请求的地址不匹配。 |
| 400 | InvalidIpProtocol.ValueNotSupported | The parameter Permissions.1.IpProtocol must be specified with case insensitive TCP, UDP, ICMP, GRE or All. | 参数Permissions.1.IpProtocol必须指定为不区分大小写的TCP、UDP、ICMP、GRE或All。 |
| 500 | InternalError | An error occurred while processing your request. | 内部错误，请重试。如果多次尝试失败，请提交工单。 |
| 403 | FirewallRuleLimitExceed | The maximum number of firewall rules in an instance is exceeded. | 防火墙规则总数超过实例限制。 |
| 403 | FirewallRuleAlreadyExist | The specified Rule already exist | 防火墙规则已存在。 |
| 404 | InvalidInstanceId.NotFound | The specified InstanceId does not exist. | 指定的实例不存在，请您检查实例ID是否正确。 |

访问[错误中心](https://api.aliyun.com/document/SWAS-OPEN/2020-06-01/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/SWAS-OPEN/2020-06-01/ModifyFirewallRule#workbench-doc-change-demo)。