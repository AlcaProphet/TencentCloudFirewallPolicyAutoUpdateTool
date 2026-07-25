本接口用于修改指定安全组中的一条入方向安全组规则。

## 接口说明

指定安全组规则 ID 修改安全组规则时，您需要注意以下使用限制：

-   安全组规则的授权对象分为 IPv4 的 CIDR 地址块（或 IP 地址）、IPv6 的 CIDR 地址块（或 IP 地址）、安全组、前缀列表，您不能通过该接口修改已有安全组规则的授权对象类型。如原来授权对象类型为 IPv4 的 CIDR 地址块，您可以更改为另一个 IPv4 的 CIDR 地址块（或 IP 地址），但不能修改为 IPv6 的 CIDR 地址块（或 IP 地址）、安全组或前缀列表。
    
-   字段不支持从非空修改为空，如果需要修改建议先增加一条新规则，再删除当前规则。
    

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/ModifySecurityGroupRule)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/Ecs/2014-05-26/ModifySecurityGroupRule)

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
| ecs:ModifySecurityGroupRule | update | \\*全部资源 `*` | - ecs:SecurityGroupIpProtocols - ecs:SecurityGroupSourceCidrIps | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| RegionId | string | 是   | 目标安全组所属地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| RegionId | string | 是   | 目标安全组所属地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| ClientToken | string | 否   | 保证请求幂等性。从您的客户端生成一个参数值，确保不同请求间该参数值唯一。**ClientToken** 只支持 ASCII 字符，且不能超过 64 个字符。更多信息，请参见[如何保证幂等性](https://help.aliyun.com/zh/ecs/developer-reference/how-to-ensure-idempotence)。 | 123e4567-e89b-12d3-a456-426655440000 |
| SecurityGroupId | string | 是   | 安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| SecurityGroupRuleId | string | 否   | 安全组规则 ID。您可以通过 [DescribeSecurityGroupAttribute](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describesecuritygroupattribute) 查询安全组规则 ID。 | sgr-bp67acfmxa123b\\*\\*\\* |
| Policy | string | 否   | 访问权限。取值范围： - accept：接受访问。 - drop：拒绝访问，不返回拒绝信息。 默认值：accept。 | accept |
| Priority | string | 否   | 安全组规则优先级。取值范围：1~100。 默认值：1。 | 1   |
| IpProtocol | string | 否   | 网络层/传输层协议。支持两类赋值： 1. 不区分大小写的协议名。取值范围： - ICMP - GRE - TCP - UDP - ALL：支持所有协议。 2. 符合 IANA 规范的协议号取值，即 0 到 255 的整数。目前开放的地域列表： - 菲律宾 - 英国 - 马来西亚 - 呼和浩特 - 青岛 - 美西 - 新加坡 | ALL |
| SourceCidrIp | string | 否   | 设置访问权限的源端 IPv4 CIDR 地址块。支持 CIDR 格式和 IPv4 格式的 IP 地址范围。 默认值：无。 | 10.0.0.0/8 |
| Ipv6SourceCidrIp | string | 否   | 设置访问权限的源端 IPv6 CIDR 地址块。支持 CIDR 格式和 IPv6 格式的 IP 地址范围。 **说明** 仅支持 VPC 类型的 IP 地址，且该参数与`SourceCidrIp`参数不可同时设置。 默认值：无。 | 2001:db8:1233:1a00::\\*\\*\\* |
| SourceGroupId | string | 否   | 设置访问权限的源端安全组 ID。至少设置一项`SourceGroupId`或者`SourceCidrIp`参数。 - 如果指定了`SourceGroupId`没有指定参数`SourceCidrIp`，则参数`NicType`取值只能为 intranet。 - 如果同时指定了`SourceGroupId`和`SourceCidrIp`，则默认以`SourceCidrIp`为准。 | sg-bp67acfmxa123b\\*\\*\\*\\* |
| SourcePrefixListId | string | 否   | 设置访问权限的源端前缀列表 ID。您可以调用 [DescribePrefixLists](https://help.aliyun.com/zh/ecs/api-describeprefixlists) 查询可以使用的前缀列表 ID。 当您指定了`SourceCidrIp`、`Ipv6SourceCidrIp`或`SourceGroupId`参数中的一个时，将忽略该参数。 | pl-x1j1k5ykzqlixdcy\\*\\*\\*\\* |
| PortRange | string | 否   | 目的端安全组开放的传输层协议相关的端口范围。取值范围： - TCP/UDP 协议：取值范围为 1~65535。使用斜线（/）隔开起始端口和终止端口。例如：1/200。 - ICMP 协议：-1/-1。 - GRE 协议：-1/-1。 - ALL：-1/-1。 | 80/80 |
| DestCidrIp | string | 否   | 目的端 IPv4 CIDR 地址块。支持 CIDR 格式和 IPv4 格式的 IP 地址范围。 默认值：无。 | 10.0.0.0/8 |
| Ipv6DestCidrIp | string | 否   | 目的端 IPv6 CIDR 地址段。支持 CIDR 格式和 IPv6 格式的 IP 地址范围。 **说明** 仅支持 VPC 类型的 IP 地址，且该参数与`DestCidrIp`参数不可同时设置。 默认值：无。 | 2001:db8:1234:1a00::\\*\\*\\* |
| SourcePortRange | string | 否   | 源端安全组开放的传输层协议相关的端口范围。取值范围： - TCP/UDP 协议：取值范围为 1~65535。使用斜线（/）隔开起始端口和终止端口。例如：1/200 - ICMP 协议：-1/-1。 - GRE 协议：-1/-1。 - ALL：-1/-1。 | 80/80 |
| SourceGroupOwnerAccount | string | 否   | 跨账户设置安全组规则时，源端安全组所属的阿里云账户。 - 如果`SourceGroupOwnerAccount`及`SourceGroupOwnerID`均未设置，则认为是设置您其他安全组的访问权限。 - 如果已经设置参数`SourceCidrIp`，则参数`SourceGroupOwnerAccount`无效。 | EcsforCloud@Alibaba.com |
| SourceGroupOwnerId | integer | 否   | 跨账户设置安全组规则时，源端安全组所属的阿里云账户。 - 如果`SourceGroupOwnerId`及`SourceGroupOwnerAccount`均未设置，则认为是设置您其他安全组的访问权限。 - 如果您已经设置参数`SourceCidrIp`，则参数`SourceGroupOwnerId`无效。 | 12345678910 |
| NicType | string | 否   | 网卡类型。 **说明** 根据安全组规则 ID 修改规则时，不支持修改该参数。如果需要修改，建议先增加一条新规则，再删除当前规则。 | intranet |
| Description | string | 否   | 安全组规则的描述信息。长度为 1~512 个字符。 | This is a new security group rule. |
| PortRangeListId | string | 否   | 端口列表 ID。 您可以调用`DescribePortRangeLists`查询可以使用的端口列表 ID。 当您指定了 PortRange 参数时，将忽略该参数。 更多信息，请参见[安全组使用限制](https://help.aliyun.com/zh/ecs/user-guide/limitations#SecurityGroupQuota1)。 | prl-2ze9743\\*\\*\\*\\* |

## **返回参数**

| **名称** | **类型** | **描述** | **示例值** |
| --- | --- | --- | --- |
|     | object |     |     |
| RequestId | string | 请求 ID。 | 473469C7-AA6F-4DC5-B3DB-A3DC0DE3C83E |

## 示例

正常返回示例

`JSON`格式

```
{
  "RequestId": "473469C7-AA6F-4DC5-B3DB-A3DC0DE3C83E"
}
```

异常返回示例

`JSON`格式

```
{
    "RequestId":"CEF72CEB-54B6-4AE8-B225-F876FF7BA984"
}
```

## 错误码

| **HTTP status code** | **错误码** | **错误信息** | **描述** |
| --- | --- | --- | --- |
| 400 | OperationDenied | The specified IpProtocol does not exist or IpProtocol and PortRange do not match. | 指定的 IP 协议不存在，或与端口范围不匹配。 |
| 400 | InvalidIpProtocol.Malformed | The specified parameter PortRange is not valid. | IP 协议参数格式不正确，PortRange 参数不正确。 |
| 400 | InvalidSourceCidrIp.Malformed | The specified parameter SourceCidrIp is not valid. | 源 IP 地址范围参数格式不正确。 |
| 400 | InvalidPolicy.Malformed | The specified parameter Policy is not valid. | 指定的参数无效，请您检查该参数是否正确。 |
| 400 | InvalidNicType.ValueNotSupported | The specified NicType does not exist. | 指定的网络类型不存在，请您检查网络类型是否正确。 |
| 400 | InvalidNicType.Mismatch | The specified NicType conflicts with the authorization record. | 指定的网卡类型与已有规则不匹配。 |
| 400 | InvalidSourceGroupId.Mismatch | Specified security group and source group are not in the same VPC. | 指定的安全组和源安全组不在一个 VPC 内。 |
| 400 | InvalidSourceGroup.NotFound | Specified source security group does not exist. | 指定的安全组入方向规则不存在，或相关参数缺失。 |
| 400 | InvalidPriority.Malformed | The parameter Priority is invalid. | 指定的参数 Priority 无效。 |
| 400 | InvalidPriority.ValueNotSupported | The parameter Priority is invalid. | 指定的参数 Priority 无效。 |
| 400 | InvalidSecurityGroupDiscription.Malformed | The specified security group rule description is not valid. | 指定的安全组规则描述不合法。 |
| 400 | MissingParameter.Source | One of the parameters SourceCidrIp, SourceGroupId or SourcePrefixListId must be specified. | 安全组规则的源必须被指定，请指定SourceCidrIp、SourceGroupId或SourcePrefixListId参数中的任何一个。 |
| 400 | InvalidParam.PortRange | The specified parameter %s is not valid. It should be two integers less than 65535 in ?/? format. | 端口范围不合法，应为斜线分隔两个整数的格式。 |
| 400 | InvalidIpProtocol.ValueNotSupported | The parameter IpProtocol must be specified with case insensitive TCP, UDP, ICMP, GRE or All. | 协议类型只能是TCP、UDP、ICMP、GRE或者All。 |
| 400 | InvalidParam.SourceIp | The Parameters SourceCidrIp and Ipv6SourceCidrIp in %s cannot be set at the same time. | 参数SourceCidrIp和Ipv6SourceCidrIp不能被同时设置。 |
| 400 | InvalidParam.DestIp | The Parameters DestCidrIp and Ipv6DestCidrIp in %s cannot be set at the same time. | 参数DestCidrIp和Ipv6DestCidrIp不能被同时设置。 |
| 400 | InvalidParam.Ipv6DestCidrIp | The specified parameter %s is not valid. | 指定的参数Ipv6DestCidrIp不合法。 |
| 400 | InvalidParam.Ipv6SourceCidrIp | The specified parameter %s is not valid. | 指定的参数Ipv6SourceCidrIp不合法。 |
| 400 | InvalidParam.Ipv4ProtocolConflictWithIpv6Address | IPv6 address cannot be specified for IPv4-specific protocol. | IPv4协议不能指定IPv6地址。 |
| 400 | InvalidParam.Ipv6ProtocolConflictWithIpv4Address | IPv4 address cannot be specified for IPv6-specific protocol. | IPv6协议不能指定IPv4地址。 |
| 400 | InvalidParameter.Ipv6CidrIp | The specified Ipv6CidrIp is not valid. | 指定的Ipv6CidrIp参数不合法。 |
| 400 | InvalidParam.DestCidrIp | The specified parameter %s is not valid. | 指定的参数DestCidrIp不合法。 |
| 400 | InvalidSourcePortRange.Malformed | The specified parameter SourcePortRange is not valid. | 指定的参数 SourcePortRange 无效。 |
| 400 | InvalidSecurityGroupId.Malformed | The specified parameter SecurityGroupId is not valid. | 指定的参数 SecurityGroupId 无效。 |
| 400 | InvalidParam.SourceCidrIp | The specified param SourceCidrIp is not valid. | 参数SourceCidrIp不合法。 |
| 400 | InvalidParameter.Conflict | IPv6 and IPv4 addresses cannot exist at the same time. | IPv6地址和IPv4地址不能同时指定。 |
| 400 | InvalidParam.SecurityGroupRuleId | The specified parameter SecurityGroupRuleId is not valid. | 指定的参数SecurityGroupRuleId不合法。 |
| 400 | InvalidOperation.ModifySgRuleEntityType | The source or destination type of the rules cannot be modified. | 规则的源或目的类型不能被修改。 |
| 400 | AuthorizationLimitExceed | The limit of authorization records in the security group reaches. | 安全组授权规则数达到上限，请您检查授权规则是否合理。 |
| 400 | InvalidParam.ProtocolAndPortRangeMismatch | The specified Protocol and PortRange do not match. | 指定和协议和端口范围不匹配。 |
| 400 | InvalidParam.ProtocolAndAddressFamilyMismatch | The specified Protocol and address family do not match. | 指定的协议和地址族不匹配。 |
| 400 | InvalidParam.PrefixListAddressFamilyMismatch | The address family of the prefix list does not match the rule. | 前缀列表的地址族与规则不匹配。 |
| 400 | InvalidParam.InvalidModifyRuleRequest | The request parameters are illegal. | 请求参数不合法。 |
| 400 | InvalidOperation.ModifyNicType | NicType is not allowed to modify. | 不允许修改NicType。 |
| 400 | InvalidParamter.Conflict | The specified SourceCidrIp should be different from the DestCidrIp. | 参数 SourceCidrIp 和 DestCidrIp 不能相同。 |
| 400 | InvalidOperation.RuleDuplicate | %s. | 修改后的规则与已有规则重复。 |
| 400 | InvalidParam.ProtocolNotSupportPortRangeList | The specified protocol does not support the port range list. | 指定的协议不支持端口列表。 |
| 400 | InvalidSourceOrDestGroupId.DirectionMissmatch | The specified SourceGroupId or DestGroupId does not match the direction of the rule. | 指定的SourceGroupId或DestGroupId与安全组规则的方向不匹配。 |
| 400 | InvalidOperation.ModifyPortRangeType | The PortRange type is not allowed to be modified. You cannot modify a rule from using the port list to not using it, and vice versa. | 安全组规则的端口范围类型不能修改。您不能将规则从使用端口列表改为不使用，反过来也不行。 |
| 400 | InvalidPortRangeListId.NotFound | The specified port range list was not found. | 未找到指定的端口列表。 |
| 500 | InternalError | The request processing has failed due to some unknown error. |     |
| 403 | InvalidSourceGroupId.Mismatch | NicType is required or NicType expects intrnet. |     |
| 403 | MissingParameter | The input parameter SourceGroupId or SourceCidrIp cannot be both blank. | 参数 SourceGroupId 和 SourceCidrIp 不能同时为空。 |
| 403 | AuthorizationLimitExceed | The limit of authorization records in the security group reaches. |     |
| 403 | InvalidParamter.Conflict | The specified SecurityGroupId should be different from the SourceGroupId. | 授权与被授权安全组必须不同。 |
| 403 | InvalidNetworkType.Mismatch | The specified SecurityGroup network type should be same with SourceGroup network type (vpc or classic). | 指定的 SecurityGroup 的网络类型必须与 SouceGroup 的网络类型一致。 |
| 403 | InvalidOperation.ResourceManagedByCloudProduct | %s  | 云产品托管的安全组不支持修改操作。 |
| 404 | InvalidSecurityGroupId.NotFound | The specified SecurityGroupId does not exist. | 指定的安全组在该用户账号下不存在，请您检查安全组 ID 是否正确。 |
| 404 | InvalidSourceGroupId.NotFound | The SourceGroupId provided does not exist in our records. | 指定的入方向安全组不存在。 |
| 404 | SecurityGroupRule.NotFound | The target security group rule not exist. | 目标安全组规则不存在。 |
| 404 | InvalidPrefixListId.NotFound | The specified prefix list was not found. | 前缀列表不存在。 |
| 404 | InvalidSecurityGroupRuleId.NotFound | The specified SecurityGroupRuleId is not exists. | 指定的SecurityGroupRuleId不存在。 |

访问[错误中心](https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/ModifySecurityGroupRule#workbench-doc-change-demo)。