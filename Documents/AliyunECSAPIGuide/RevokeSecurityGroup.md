本接口用于删除指定安全组内的一条或多条入方向安全组规则。

## 接口说明

**重要** 阿里云已于 2024 年 7 月 8 日对该接口进行了校验规则调整。当删除不存在的安全组规则时，从返回成功调整为返回错误码：“InvalidParam.SecurityGroupRuleId“。请您及时做好错误码兼容，避免影响线上业务。

该接口存在两种传参方式来删除规则：

-   通过指定安全组规则 ID 参数删除规则（推荐）。
    
    -   如果指定的安全组规则 ID 不存在，接口调用将失败。
        
-   通过指定 Permissions 删除规则。
    
    -   如果匹配的安全组规则不存在，此次调用成功，但不会删除规则。
        
    -   确定一条安全组入方向规则必要的一组相关参数：
        -   源端设置：选择 SourceCidrIp（IPv4 地址）、Ipv6SourceCidrIp（IPv6 地址）、SourcetPrefixListId（前缀列表 ID）、SourceGroupId（源端安全组）中的一项。
            
        -   目的端口范围：PortRange。
            
        -   协议类型：IpProtocol。
            
        -   权限策略：Policy。
            

**说明**

不支持同时设置安全组规则 ID 和 Permissions 参数。

### 请求示例

-   根据指定安全组规则 ID 删除。
    

```
"SecurityGroupId":"sg-bp67acfmxazb4p****", //设置安全组 ID
"SecurityGroupRuleId":["sgr-bpdfmk****","sgr-bpdfmg****"] //设置安全组规则 ID
```

-   根据指定 IP 地址段删除。
    

```
"SecurityGroupId":"sg-bp67acfmxazb4p****",
"Permissions":[
  {
    "SourceCidrIp":"10.0.0.0/8", //设置源端 IP 地址段
    "IpProtocol":"TCP", //设置协议类型
    "PortRange":"80/80", //设置目的端口范围
    "Policy":"accept" //设置访问策略
  }
]
```

-   根据其他安全组删除。
    

```
"SecurityGroupId":"sg-bp67acfmxazb4p****",
"Permissions":[
  {
    "SourceGroupId":"sg-bp67acfmxa123b****", //设置源端安全组 ID
    "IpProtocol":"TCP,"
    "PortRange":"80/80",
    "Policy":"accept"
  ]
}
```

-   根据指定前缀列表删除。
    

```
"SecurityGroupId":"sg-bp67acfmxazb4p****",
"Permissions":[
  {
    "SourcePrefixListId":pl-x1j1k5ykzqlixdcy****", //设置源端前缀列表 ID
    "IpProtocol":"TCP",
    "PortRange":"80/80",
    "Policy":"accept"
  }
]
```

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/RevokeSecurityGroup)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/Ecs/2014-05-26/RevokeSecurityGroup)

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
| ecs:RevokeSecurityGroup | delete | \\*SecurityGroup `acs:ecs:{#regionId}:{#accountId}:securitygroup/{#securitygroupId}` | - ecs:tag - ecs:tag - ecs:tag | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| RegionId | string | 是   | 安全组所属地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| RegionId | string | 是   | 安全组所属地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| ClientToken | string | 否   | 保证请求幂等性。从您的客户端生成一个参数值，确保不同请求间该参数值唯一。**ClientToken** 只支持 ASCII 字符，且不能超过 64 个字符。更多详情，请参见[如何保证幂等性](https://help.aliyun.com/zh/ecs/developer-reference/how-to-ensure-idempotence)。 | 123e4567-e89b-12d3-a456-426655440000 |
| ClientToken | string | 否   | 保证请求幂等性。从您的客户端生成一个参数值，确保不同请求间该参数值唯一。**ClientToken** 只支持 ASCII 字符，且不能超过 64 个字符。更多详情，请参见[如何保证幂等性](https://help.aliyun.com/zh/ecs/developer-reference/how-to-ensure-idempotence)。 | 123e4567-e89b-12d3-a456-426655440000 |
| SecurityGroupId | string | 是   | 安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| SecurityGroupId | string | 是   | 安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| SecurityGroupRuleId | array | 否   | 安全组规则 ID 数组。数组长度：0~100。 |     |
|     | string | 否   | 安全组规则 ID。 **说明** 通过安全组规则 ID 删除时，该参数必填。 | sgr-bp67acfmxa123b\\*\\*\\* |
| Permissions | array<object> | 否   | 安全组规则数组。数组长度：0~100。 |     |
|     | object | 否   | 安全组规则。 |     |
| Policy | string | 否   | 访问权限。取值范围： - accept：接受访问。 - drop：拒绝访问，不返回拒绝信息，表现为发起端请求超时或者无法建立连接的类似信息。 默认值：accept。 | accept |
| Priority | string | 否   | 安全组规则优先级，数字越小，代表优先级越高。取值范围：1~100。 默认值：1。 | 1   |
| IpProtocol | string | 否   | 协议类型。取值不区分大小写。取值范围： - TCP。 - UDP。 - ICMP。 - ICMPv6。 - GRE。 - ALL：支持所有协议。 | TCP |
| SourceCidrIp | string | 否   | 需要撤销访问权限的源端 IPv4 CIDR 地址块。支持 CIDR 格式和 IPv4 格式的 IP 地址范围。 | 10.0.0.0/8 |
| Ipv6SourceCidrIp | string | 否   | 需要撤销访问权限的源端 IPv6 CIDR 地址块。支持 CIDR 格式和 IPv6 格式的 IP 地址范围。 **说明** 仅在支持 IPv6 的 VPC 类型 ECS 实例上有效，且该参数与`SourceCidrIp`参数不可同时设置。 | 2001:db8:1234:1a00::\\*\\*\\* |
| SourceGroupId | string | 否   | 需要撤销访问权限的源端安全组 ID。 - 至少设置`SourceGroupId`、`SourceCidrIp`、`Ipv6SourceCidrIp`或`SourcePrefixListId`参数中的一项。 - 如果指定了`SourceGroupId`，没有指定参数`SourceCidrIp`或`Ipv6SourceCidrIp`，则参数 NicType 取值只能为 intranet。 - 如果同时指定了`SourceGroupId`和`SourceCidrIp`，则默认以`SourceCidrIp`为准。 您需要注意： - 企业安全组不支持授权安全组访问。 - 普通安全组支持授权的安全组数量最多为 20 个。 | sg-bp67acfmxa123b\\*\\*\\*\\* |
| SourcePrefixListId | string | 否   | 需要撤销访问权限的源端前缀列表 ID。您可以调用 [DescribePrefixLists](https://help.aliyun.com/zh/ecs/api-describeprefixlists) 查询可以使用的前缀列表 ID。 注意事项： 当您指定了`SourceCidrIp`、`Ipv6SourceCidrIp`或`SourceGroupId`参数中的一个时，将忽略该参数。 更多信息，请参见[安全组使用限制](https://help.aliyun.com/zh/ecs/user-guide/limitations#SecurityGroupQuota1)。 | pl-x1j1k5ykzqlixdcy\\*\\*\\*\\* |
| PortRange | string | 否   | 安全组开放的各协议相关的目的端口范围。取值范围： - TCP/UDP 协议：取值范围为 1~65535。使用正斜线（/）隔开起始端口和终止端口。例如：1/200 - ICMP 协议：-1/-1。 - GRE 协议：-1/-1。 - ALL：-1/-1。 | 1/200 |
| DestCidrIp | string | 否   | 目的端 IPv4 CIDR 地址段。支持 CIDR 格式和 IPv4 格式的 IP 地址范围。 用于支持五元组规则，请参见[安全组五元组规则](https://help.aliyun.com/zh/ecs/user-guide/security-group-quintuple-rules)。 | 10.0.0.0/8 |
| Ipv6DestCidrIp | string | 否   | 目的端 IPv6 CIDR 地址段。支持 CIDR 格式和 IPv6 格式的 IP 地址范围。 用于支持五元组规则，请参见[安全组五元组规则](https://help.aliyun.com/zh/ecs/user-guide/security-group-quintuple-rules)。 **说明** 仅在支持 IPv6 的 VPC 类型 ECS 实例上有效，且该参数与`DestCidrIp`参数不可同时设置。 | 2001:db8:1233:1a00::\\*\\*\\* |
| SourcePortRange | string | 否   | 安全组开放的各协议相关的源端端口范围。取值范围： - TCP/UDP 协议：取值范围为 1~65535。使用正斜线（/）隔开起始端口和终止端口。例如：1/200。 - ICMP 协议：-1/-1。 - GRE 协议：-1/-1。 - ALL：-1/-1。 用于支持五元组规则，请参见[安全组五元组规则](https://help.aliyun.com/zh/ecs/user-guide/security-group-quintuple-rules)。 | 80/80 |
| SourceGroupOwnerAccount | string | 否   | 撤销跨账户授权的安全组规则时，源端安全组所属的阿里云账户。 - 如果`SourceGroupOwnerAccount`及`SourceGroupOwnerId`均未设置，则认为是撤销您其他安全组的访问权限。 - 如果已经设置参数`SourceCidrIp`，则参数`SourceGroupOwnerAccount`无效。 | Test@aliyun.com |
| SourceGroupOwnerId | integer | 否   | 撤销跨账户授权的安全组规则时，源端安全组所属的阿里云账户 ID。 - 如果`SourceGroupOwnerId`及`SourceGroupOwnerAccount`均未设置，则认为是撤销您其他安全组的访问权限。 - 如果您已经设置参数`SourceCidrIp`，则参数`SourceGroupOwnerId`无效。 | 12345678910 |
| NicType | string | 否   | 安全组规则的网卡类型。专有网络 VPC 类型安全组规则无需设置网卡类型，默认为 intranet，只能为 intranet。 **说明** 经典网络功能已下线，详情请参见[下线公告](https://help.aliyun.com/zh/ecs/end-of-sale-announcement-for-ecs-instances-of-the-classic-network-type)。经典网络类型安全组规则的网卡类型。取值范围： - internet：公网网卡。 - intranet：内网网卡。 | intranet |
| Description | string | 否   | 安全组规则描述。长度为 1~512 个字符。 | This is description. |
| PortRangeListId | string | 否   | 端口列表 ID。 您可以调用`DescribePortRangeLists`查询可以使用的端口列表 ID。 当您指定了`Permissions.N.PortRange`参数时，将忽略该参数。 更多信息，请参见[安全组使用限制](https://help.aliyun.com/zh/ecs/user-guide/limitations#SecurityGroupQuota1)。 | prl-2ze9743\\*\\*\\*\\* |
| Policy `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.Policy`来设置访问权限。 | accept |
| Priority `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.Priority`来指定规则优先级。 | 1   |
| IpProtocol `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.IpProtocol`来指定协议类型。 | ALL |
| SourceCidrIp `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.SourceCidrIp`来指定源端 IPv4 CIDR 地址块。 | 10.0.0.0/8 |
| Ipv6SourceCidrIp `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.Ipv6SourceCidrIp`来指定源端 IPv6 CIDR 地址块。 | 2001:db8:1234:1a00::\\*\\*\\* |
| SourceGroupId `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.SourceGroupId`来指定源端安全组 ID。 | sg-bp67acfmxa123b\\*\\*\\*\\* |
| SourcePrefixListId `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.SourcePrefixListId`来指定源端前缀列表 ID。 | pl-x1j1k5ykzqlixdcy\\*\\*\\*\\* |
| PortRange `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.PortRange`来指定端口范围。 | 1/200 |
| DestCidrIp `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.DestCidrIp`来指定目的端 IPv4 CIDR 地址段。 | 10.0.0.0/8 |
| Ipv6DestCidrIp `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.Ipv6DestCidrIp`来指定目的端 IPv6 CIDR 地址段。 | 2001:db8:1233:1a00::\\*\\*\\* |
| SourcePortRange `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.SourcePortRange`来指定源端端口范围。 | 80/80 |
| SourceGroupOwnerAccount `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.SourceGroupOwnerAccount`来指定源端安全组所属的阿里云账户。 | Test@aliyun.com |
| SourceGroupOwnerId `deprecated` | integer | 否   | 已废弃。请使用`Permissions.N.SourceGroupOwnerId`来指定源端安全组所属的阿里云账户 ID。 | 12345678910 |
| NicType `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.NicType`来指定网卡类型。 | intranet |
| Description `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.Description`来指定规则的描述。 | This is description. |

## **返回参数**

| **名称** | **类型** | **描述** | **示例值** |
| --- | --- | --- | --- |
|     | object |     |     |
| RequestId | string | 请求 ID。 | 473469C7-AA6F-4DC5-B3DB-A3DC0DE3\\*\\*\\*\\* |

## 示例

正常返回示例

`JSON`格式

```
{
  "RequestId": "473469C7-AA6F-4DC5-B3DB-A3DC0DE3****"
}
```

## 错误码

| **HTTP status code** | **错误码** | **错误信息** | **描述** |
| --- | --- | --- | --- |
| 400 | InvalidSecurityGroupId.Malformed | The specified parameter SecurityGroupId is not valid. | 指定的参数 SecurityGroupId 无效。 |
| 400 | InvalidIpProtocol.ValueNotSupported | The parameter IpProtocol must be specified with case insensitive TCP, UDP, ICMP, GRE or All. | 协议类型只能是TCP、UDP、ICMP、GRE或者All。 |
| 400 | InvalidIpPortRange.Malformed | The specified parameter PortRange is not valid. | IP协议相关的端口号范围格式不正确。 |
| 400 | InvalidSourceCidrIp.Malformed | The specified parameter SourceCidrIp is not valid. | 源 IP 地址范围参数格式不正确。 |
| 400 | MissingParameter | The input parameter SourceGroupId or SourceCidrIp cannot be both blank. |     |
| 400 | InvalidPolicy.Malformed | The specified parameter %s is not valid. | 指定的授权策略参数Policy不合法。 |
| 400 | InvalidNicType.ValueNotSupported | The specified parameter %s is not valid. | 指定的网卡类型参数NicType不合法。 |
| 400 | InvalidSourceGroupId.Mismatch | Specified security group and source group are not in the same VPC. | 指定的安全组和源安全组不在一个 VPC 内。 |
| 400 | MissingParameter.Source | One of the parameters SourceCidrIp, Ipv6SourceCidrIp, SourceGroupId or SourcePrefixListId in %s must be specified. | 至少需要指定参数SourceCidrIp、SourceGroupId或SourcePrefixListId中的一个。 |
| 400 | InvalidParam.PortRange | The specified parameter %s is not valid. It should be two integers less than 65535 in ?/? format. | 端口范围不合法，应为斜线分隔两个整数的格式。 |
| 400 | InvalidPriority.Malformed | The parameter Priority is invalid. | 指定的参数 Priority 无效。 |
| 400 | InvalidPriority.ValueNotSupported | The specified parameter %s is invalid. | 指定的优先级参数Priority不合法。 |
| 400 | InvalidParamter.Conflict | The specified SecurityGroupId should be different from the SourceGroupId. |     |
| 400 | InvalidDestCidrIp.Malformed | The specified parameter DestCidrIp is not valid. | 指定的 DestCidrIp 无效，请您检查该参数是否正确。 |
| 400 | InvalidParam.SourceIp | The Parameters SourceCidrIp and Ipv6SourceCidrIp in %s cannot be set at the same time. | 参数SourceCidrIp和Ipv6SourceCidrIp不能被同时设置。 |
| 400 | InvalidParam.DestIp | The Parameters DestCidrIp and Ipv6DestCidrIp in %s cannot be set at the same time. | 参数DestCidrIp和Ipv6DestCidrIp不能被同时设置。 |
| 400 | InvalidParam.Ipv6DestCidrIp | The specified parameter %s is not valid. | 指定的参数Ipv6DestCidrIp不合法。 |
| 400 | InvalidParam.Ipv6SourceCidrIp | The specified parameter %s is not valid. | 指定的参数Ipv6SourceCidrIp不合法。 |
| 400 | InvalidParam.Ipv4ProtocolConflictWithIpv6Address | IPv6 address cannot be specified for IPv4-specific protocol. | IPv4协议不能指定IPv6地址。 |
| 400 | InvalidParam.Ipv6ProtocolConflictWithIpv4Address | IPv4 address cannot be specified for IPv6-specific protocol. | IPv6协议不能指定IPv4地址。 |
| 400 | InvalidParameter.Ipv6CidrIp | The specified Ipv6CidrIp is not valid. | 指定的Ipv6CidrIp参数不合法。 |
| 400 | InvalidGroupAuthParameter.OperationDenied | The security group can not authorize to enterprise level security group. | 安全组不能授权给企业级安全组。 |
| 400 | InvalidPortRange.Malformed | The specified parameter PortRange must set. | 参数PortRange必须被设置。 |
| 400 | InvalidSourcePortRange.Malformed | The specified parameter SourcePortRange is not valid. | 指定的参数 SourcePortRange 无效。 |
| 400 | InvalidSecurityGroupDiscription.Malformed | The specified security group rule description is not valid. | 指定的安全组规则描述不合法。 |
| 400 | NotSupported.ClassicNetworkPrefixList | The prefix list is not supported when the network type of security group is classic. | 安全组的网络类型为经典网络，不支持前缀列表。 |
| 400 | InvalidParam.SourceCidrIp | The specified parameter %s is not valid. | 指定的参数SourceCidrIp不合法。 |
| 400 | InvalidParam.DestCidrIp | The specified parameter %s is not valid. | 指定的参数DestCidrIp不合法。 |
| 400 | InvalidParam.Permissions | The specified parameter Permissions cannot coexist with other parameters. | 指定的参数Permissions不能与其他参数同时存在。 |
| 400 | InvalidParam.DuplicatePermissions | There are duplicate permissions in the specified parameter Permissions. | 指定的参数Permissions中存在重复的规则。 |
| 400 | InvalidSecurityGroupId.NotFound | The specified parameter SecurityGroupId is not valid. | 指定的安全组在该用户账号下不存在，请您检查安全组 id 是否正确。 |
| 400 | InvalidParam.SecurityGroupRuleId | The specified parameter SecurityGroupRuleId is not valid. | 指定的参数SecurityGroupRuleId不合法。 |
| 400 | InvalidParam.SecurityGroupRuleIdRepeated | The specified parameter SecurityGroupRuleId is repeated. | 指定的参数SecurityGroupRuleId存在重复。 |
| 400 | InvalidGroupParameter.OperationDenied | The attributes Policy, SourceGroupId, DestGroupId of enterprise level security groups are not allowed to be set or modified. | 企业级安全组不能指定Policy，SourceGroupId，DestGroupId等参数。 |
| 400 | InvalidSecurityGroupRule.RuleNotExist | The specified rule does not exist. | 指定的安全组规则不存在。 |
| 400 | InvalidParam.ProtocolNotSupportPortRangeList | The specified protocol does not support the port range list. | 指定的协议不支持端口列表。 |
| 400 | InvalidPortRangeListId.NotFound | The specified port range list was not found. | 未找到指定的端口列表。 |
| 400 | InvalidSecurityGroup.InvalidNetworkType | The specified security group network type is not support this operation, please check the security group network types. For VPC security groups, ClassicLink must be enabled. | 指定的安全组网络类型不支持此操作，请检查安全组网络类型。对于 VPC 安全组，必须启用 ClassicLink。 |
| 401 | InvalidOperation.SecurityGroupNotAuthorized | The specified security group is not authorized to operate. | 没有权限操作当前安全组 |
| 500 | InternalError | The request processing has failed due to some unknown error. |     |
| 403 | InvalidNicType.Mismatch | The specified NicType conflicts with the authorization record. |     |
| 403 | InvalidGroupAuthItem.NotFound | Specified group authorized item does not exist in our records. |     |
| 403 | InvalidOperation.ResourceManagedByCloudProduct | %s  | 云产品托管的安全组不支持修改操作。 |
| 404 | InvalidSecurityGroupId.NotFound | The specified SecurityGroupId does not exist. | 指定的安全组在该用户账号下不存在，请您检查安全组 ID 是否正确。 |
| 404 | InvalidSourceGroupId.NotFound | The SourceGroupId provided does not exist in our records. | 指定的入方向安全组不存在。 |
| 404 | InvalidPrefixListId.NotFound | The specified prefix list was not found. | 前缀列表不存在。 |
| 404 | InvalidSecurityGroupRuleId.NotFound | The specified SecurityGroupRuleId is not exists. | 指定的SecurityGroupRuleId不存在。 |

访问[错误中心](https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/RevokeSecurityGroup#workbench-doc-change-demo)。