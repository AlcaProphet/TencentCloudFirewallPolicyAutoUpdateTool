本接口是阿里云 ECS 中用于增加一条或多条安全组出方向规则的接口。通过该接口，用户可以指定安全组出方向的访问权限，允许或拒绝安全组内的实例发送出方向流量到其他设备，从而实现对网络访问的精细控制。

## 接口说明

### 使用须知

-   **数量限制：** 单张弹性网卡关联的所有安全组的规则（包括入方向规则与出方向规则）数量之和不能超过 1000。具体限制请参见[安全组使用限制](https://help.aliyun.com/zh/ecs/user-guide/limitations#SecurityGroupQuota1)。
    
-   **优先级设置：** 安全组出方向规则优先级（Priority）可选范围为 1~100。数字越小，代表优先级越高，优先级相同的安全组规则，优先以拒绝访问（drop）的规则为准。
    

### 注意事项

如果指定的安全组规则已存在，此次调用成功，但不会增加规则。

### 规则确定方式

确定一条安全组出方向规则必要的一组相关参数：

-   目的端设置：选择 DestCidrIp（IPv4 地址）、Ipv6DestCidrIp（IPv6 地址）、DestPrefixListId（前缀列表 ID）、DestGroupId（目的端安全组）中的一项。
    
-   目的端口范围：PortRange。
    
-   协议类型：IpProtocol。
    
-   权限策略：Policy。
    

**说明**

企业安全组不支持授权其他安全组访问，普通安全组支持授权的安全组数量最多为 20 个。

### 请求示例

假设要在杭州地域下指定安全组中增加几条不同目的端的出方向规则：

-   增加指定 IP 地址段的访问权限。
    
    ```
    "RegionId":"cn-hangzhou",  //设置地域
    "SecurityGroupId":"sg-bp17vs63txqxbds9***", //设置安全组
    "Permissions":[
         {
           "DestCidrIp":"10.0.0.0/8", //设置目的端 IPv4 地址
           "PortRange":"-1/-1", //设置端口范围
           "IpProtocol":"ICMP", //设置协议类型
           "Policy":"Accept" //设置访问策略
         }
    ]
    ```
    
-   增加一条其他安全组和一条前缀列表的访问权限。
    
    ```
    "RegionId":"cn-hangzhou",
    "SecurityGroupId":"sg-bp17vs63txqxbds9***",
    "Permissions":[
         {
           "DestGroupId":"sg-bp67acfmxazb4pi***", //设置目的端安全组
           "PortRange":"22/22",
           "IpProtocol":"TCP",
           "Policy":"Drop"
         },{
           "DestPrefixListId":"pl-x1j1k5ykzqlixdcy****", //设置目的端前缀列表
           "PortRange":"22/22",
           "IpProtocol":"TCP",
           "Policy":"Drop"
         }
    ]
    ```
    

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/AuthorizeSecurityGroupEgress)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/Ecs/2014-05-26/AuthorizeSecurityGroupEgress)

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
| ecs:AuthorizeSecurityGroupEgress | create | \\*全部资源 `*` | - ecs:SecurityGroupIpProtocols - ecs:SecurityGroupSourceCidrIps | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| RegionId | string | 是   | 源端安全组所属地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| RegionId | string | 是   | 源端安全组所属地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| RegionId | string | 是   | 源端安全组所属地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| ClientToken | string | 否   | 保证请求幂等性。从您的客户端生成一个参数值，确保不同请求间该参数值唯一。**ClientToken** 只支持 ASCII 字符，且不能超过 64 个字符。更多信息，请参见[如何保证幂等性](https://help.aliyun.com/zh/ecs/developer-reference/how-to-ensure-idempotence)。 | 123e4567-e89b-12d3-a456-426655440000 |
| ClientToken | string | 否   | 保证请求幂等性。从您的客户端生成一个参数值，确保不同请求间该参数值唯一。**ClientToken** 只支持 ASCII 字符，且不能超过 64 个字符。更多信息，请参见[如何保证幂等性](https://help.aliyun.com/zh/ecs/developer-reference/how-to-ensure-idempotence)。 | 123e4567-e89b-12d3-a456-426655440000 |
| SecurityGroupId | string | 是   | 安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| SecurityGroupId | string | 是   | 安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| Permissions | array<object> | 否   | 安全组规则数组。数组长度：1~100。 |     |
|     | object | 否   | 安全组权限规则。 |     |
| Policy | string | 否   | 设置访问权限。取值范围： - accept：接受访问。 - drop：拒绝访问，不返回拒绝信息，表现为发起端请求超时或者无法建立连接的类似信息。 默认值：accept。 | accept |
| Priority | string | 否   | 安全组规则优先级。数字越小，代表优先级越高。取值范围：1~100。 默认值：1。 | 1   |
| IpProtocol | string | 否   | 网络层/传输层协议。支持两类赋值： 1. 不区分大小写的协议名。取值范围： - ICMP - GRE - TCP - UDP - ALL：支持所有协议。 2. 符合 IANA 规范的协议号取值，即 0 到 255 的整数。目前开放的地域列表： - 菲律宾 - 英国 - 马来西亚 - 呼和浩特 - 青岛 - 美西 - 新加坡 | ALL |
| DestCidrIp | string | 否   | 需要设置访问权限的目的端 IPv4 CIDR 地址块。支持 CIDR 格式和 IPv4 格式的 IP 地址范围。 | 10.0.0.0/8 |
| Ipv6DestCidrIp | string | 否   | 需要设置访问权限的目的端 IPv6 CIDR 地址块。支持 CIDR 格式和 IPv6 格式的 IP 地址范围。 **说明** 仅在支持 IPv6 的 VPC 类型 ECS 实例上有效，且该参数与`DestCidrIp`参数不可同时设置。 | 2001:db8:1233:1a00::\\*\\*\\* |
| DestGroupId | string | 否   | 需要设置访问权限的目的端安全组 ID。 - 至少设置`DestGroupId`、`DestCidrIp`、`Ipv6DestCidrIp`或`DestPrefixListId`参数中的一项。 - 如果指定了`DestGroupId`没有指定参数`DestCidrIp`，则参数`NicType`取值默认只能为 intranet。 - 如果同时指定了`DestGroupId`和`DestCidrIp`，则默认以`DestCidrIp`为准。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| DestPrefixListId | string | 否   | 需要设置访问权限的目的端前缀列表 ID。您可以调用 [DescribePrefixLists](https://help.aliyun.com/zh/ecs/api-describeprefixlists) 查询可以使用的前缀列表 ID。 注意事项： 当您指定了`DestCidrIp`、`Ipv6DestCidrIp`或`DestGroupId`参数中的一个时，将忽略该参数。 更多信息，请参见[安全组使用限制](https://help.aliyun.com/zh/ecs/user-guide/limitations#SecurityGroupQuota1)。 | pl-x1j1k5ykzqlixdcy\\*\\*\\*\\* |
| PortRange | string | 否   | 安全组开放的各协议相关的目的端口范围。取值范围： - TCP/UDP：取值范围为 1~65535。使用正斜线（/）隔开起始端口和终止端口。例如：1/200。 - ICMP：-1/-1。 - GRE：-1/-1。 - ALL：-1/-1。 | 80/80 |
| SourceCidrIp | string | 否   | 源端 IPv4 CIDR 地址段。支持 CIDR 格式和 IPv4 格式的 IP 地址范围。 用于支持五元组规则，请参见[安全组五元组规则](https://help.aliyun.com/zh/ecs/user-guide/security-group-quintuple-rules)。 | 10.0.0.0/8 |
| Ipv6SourceCidrIp | string | 否   | 源端 IPv6 CIDR 地址段。支持 CIDR 格式和 IPv6 格式的 IP 地址范围。 用于支持五元组规则，请参见[安全组五元组规则](https://help.aliyun.com/zh/ecs/user-guide/security-group-quintuple-rules)。 **说明** 仅在支持 IPv6 的 VPC 类型 ECS 实例上有效，且该参数与`DestCidrIp`参数不可同时设置。 | 2001:db8:1234:1a00::\\*\\*\\* |
| SourcePortRange | string | 否   | 安全组开放的各协议相关的源端端口范围。取值范围： - TCP/UDP 协议：1~65535。使用正斜线（/）隔开起始端口和终止端口。例如：1/200。 - ICMP 协议：-1/-1。 - GRE 协议：-1/-1。 - ALL：-1/-1。 用于支持五元组规则，请参见[安全组五元组规则](https://help.aliyun.com/zh/ecs/user-guide/security-group-quintuple-rules)。 | 80/80 |
| DestGroupOwnerAccount | string | 否   | 跨账户设置安全组规则时，目的端安全组所属的阿里云账户。 - 如果`DestGroupOwnerAccount`及`DestGroupOwnerId`均未设置，则认为是设置您其他安全组的访问权限。 - 如果已经设置参数`DestCidrIp`，则参数`DestGroupOwnerAccount`无效。 | Test@aliyun.com |
| DestGroupOwnerId | integer | 否   | 跨账户设置安全组规则时，目的端安全组所属的阿里云账户 ID。 - 如果`DestGroupOwnerId`及`DestGroupOwnerAccount`均未设置，则认为是设置您其他安全组的访问权限。 - 如果您已经设置参数`DestCidrIp`，则参数`DestGroupOwnerId`无效。 | 12345678910 |
| NicType | string | 否   | 经典网络类型安全组规则的网卡类型。取值范围： - internet：公网网卡。 - intranet：内网网卡。 - 专有网络 VPC 类型安全组规则无需设置网卡类型，默认只能为 intranet。 - 设置安全组之间互相访问时，即仅指定了 DestGroupId 参数时，默认只能为 intranet。 默认值：internet。 | intranet |
| Description | string | 否   | 安全组规则的描述信息。长度为 1~512 个字符。 | This is description. |
| PortRangeListId | string | 否   | 端口列表 ID。 您可以调用`DescribePortRangeLists`查询可以使用的端口列表 ID。 - 当您指定了`Permissions.N.PortRange`参数时，将忽略该参数。 - 安全组的网络类型为经典网络时，不支持设置端口列表。关于安全组以及端口列表使用限制的更多信息，请参见[安全组使用限制](https://help.aliyun.com/zh/ecs/user-guide/limitations#SecurityGroupQuota1)。 | prl-2ze9743\\*\\*\\*\\* |
| Policy `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.Policy`来设置访问权限。 | accept |
| Priority `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.Priority`来指定规则优先级。 | 1   |
| IpProtocol `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.IpProtocol`来指定协议类型。 | ALL |
| DestCidrIp `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.DestCidrIp`来指定目的端 IPv4 CIDR 地址块。 | 10.0.0.0/8 |
| Ipv6DestCidrIp `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.Ipv6DestCidrIp`来指定目的端 IPv6 CIDR 地址块。 | 2001:db8:1233:1a00::\\*\\*\\* |
| DestGroupId `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.DestGroupId`来指定目的端安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| DestPrefixListId `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.DestPrefixListId`来指定源端前缀列表 ID。 | pl-x1j1k5ykzqlixdcy\\*\\*\\*\\* |
| PortRange `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.PortRange`来指定端口范围。 | 80/80 |
| SourceCidrIp `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.SourceCidrIp`来指定源端 IPv4 CIDR 地址段。 | 10.0.0.0/8 |
| Ipv6SourceCidrIp `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.Ipv6SourceCidrIp`来指定源端 IPv6 CIDR 地址段。 | 2001:db8:1234:1a00::\\*\\*\\* |
| SourcePortRange `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.SourcePortRange`来指定源端端口范围。 | 80/80 |
| DestGroupOwnerAccount `deprecated` | string | 否   | 已废弃。请使用`Permissions.N.DestGroupOwnerAccount`来指定目的端安全组所属的阿里云账户。 | Test@aliyun.com |
| DestGroupOwnerId `deprecated` | integer | 否   | 已废弃。请使用`Permissions.N.DestGroupOwnerId`来指定目的端安全组所属的阿里云账户 ID。 | 12345678910 |
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
| 400 | OperationDenied | The specified IpProtocol does not exist or IpProtocol and PortRange do not match. | 指定的 IP 协议不存在，或与端口范围不匹配。 |
| 400 | InvalidIpProtocol.Malformed | The specified parameter PortRange is not valid. | IP 协议参数格式不正确，PortRange 参数不正确。 |
| 400 | InvalidDestCidrIp.Malformed | The specified parameter DestCidrIp is not valid. | 指定的 DestCidrIp 无效，请您检查该参数是否正确。 |
| 400 | InvalidPolicy.Malformed | The specified parameter Policy is not valid. | 指定的参数无效，请您检查该参数是否正确。 |
| 400 | InvalidNicType.ValueNotSupported | The specified NicType does not exist. | 指定的网络类型不存在，请您检查网络类型是否正确。 |
| 400 | InvalidNicType.Mismatch | The specified NicType conflicts with the authorization record. | 指定的网卡类型与已有规则不匹配。 |
| 400 | InvalidDestGroupId.Mismatch | Specified security group and destination group are not in the same VPC. | 指定的安全组和目标组不在同一个 VPC 下。 |
| 400 | InvalidDestGroup.NotFound | Specified destination security group does not exist. | 指定的出方向安全组不存在。 |
| 400 | InvalidPriority.Malformed | The parameter Priority is invalid. | 指定的参数 Priority 无效。 |
| 400 | InvalidPriority.ValueNotSupported | The specified parameter %s is invalid. | 指定的优先级参数Priority不合法。 |
| 400 | InvalidSecurityGroupDiscription.Malformed | The specified security group rule description parameter %s is not valid. | 指定的安全组规则描述不合法。 |
| 400 | InvalidSecurityGroup.InvalidNetworkType | The specified security group network type is not support this operation, please check the security group network types. For VPC security groups, ClassicLink must be enabled. | 指定的安全组网络类型不支持此操作，请检查安全组网络类型。对于 VPC 安全组，必须启用 ClassicLink。 |
| 400 | MissingParameter.Dest | One of the parameters DestCidrIp, Ipv6DestCidrIp, DestGroupId or DestPrefixListId in %s must be specified. | 请指定参数DestCidrIp，Ipv6DestCidrIp，DestGroupId或DestPrefixListId中的至少一个。 |
| 400 | InvalidParam.PortRange | The specified parameter %s is not valid. It should be two integers less than 65535 in ?/? format. | 端口范围不合法，应为斜线分隔两个整数的格式。 |
| 400 | InvalidIpProtocol.ValueNotSupported | The parameter %s must be specified with case insensitive TCP, UDP, ICMP, GRE or All. | 协议Protocol字段不合法，应指定大小写不敏感的TCP，UDP，ICMP，GRE或All。 |
| 400 | InvalidSecurityGroupId.Malformed | The specified parameter SecurityGroupId is not valid. | 指定的参数 SecurityGroupId 无效。 |
| 400 | InvalidParamter.Conflict | The specified SourceCidrIp should be different from the DestCidrIp. | 参数 SourceCidrIp 和 DestCidrIp 不能相同。 |
| 400 | InvalidSourcePortRange.Malformed | The specified parameter SourcePortRange is not valid. | 指定的参数 SourcePortRange 无效。 |
| 400 | InvalidPortRange.Malformed | The specified parameter PortRange must set. | 参数PortRange必须被设置。 |
| 400 | InvalidParam.SourceIp | The Parameters SourceCidrIp and Ipv6SourceCidrIp in %s cannot be set at the same time. | 参数SourceCidrIp和Ipv6SourceCidrIp不能被同时设置。 |
| 400 | InvalidParam.DestIp | The Parameters DestCidrIp and Ipv6DestCidrIp in %s cannot be set at the same time. | 参数DestCidrIp和Ipv6DestCidrIp不能被同时设置。 |
| 400 | InvalidParam.Ipv6DestCidrIp | The specified parameter %s is not valid. | 指定的参数Ipv6DestCidrIp不合法。 |
| 400 | InvalidParam.Ipv6SourceCidrIp | The specified parameter %s is not valid. | 指定的参数Ipv6SourceCidrIp不合法。 |
| 400 | InvalidParam.Ipv4ProtocolConflictWithIpv6Address | IPv6 address cannot be specified for IPv4-specific protocol. | IPv4协议不能指定IPv6地址。 |
| 400 | InvalidParam.Ipv6ProtocolConflictWithIpv4Address | IPv4 address cannot be specified for IPv6-specific protocol. | IPv6协议不能指定IPv4地址。 |
| 400 | InvalidParameter.Ipv6CidrIp | The specified Ipv6CidrIp is not valid. | 指定的Ipv6CidrIp参数不合法。 |
| 400 | InvalidGroupAuthParameter.OperationDenied | The security group can not authorize to enterprise level security group. | 安全组不能授权给企业级安全组。 |
| 400 | InvalidParameter.Conflict | IPv6 and IPv4 addresses cannot exist at the same time. | IPv6地址和IPv4地址不能同时指定。 |
| 400 | InvalidParam.PrefixListAddressFamilyMismatch | The address family of the specified prefix list does not match the specified CidrIp. | 指定前缀列表的地址族，与指定的CidrIP的地址族不匹配。 |
| 400 | NotSupported.ClassicNetworkPrefixList | The prefix list is not supported when the network type of security group is classic. | 安全组的网络类型为经典网络，不支持前缀列表。 |
| 400 | AuthorizedGroupRule.LimitExceed | You have reached the limit on the number of group authorization rules that you can add to a security group.When authorization object of rule is security group, the limit is 20. | 普通安全组内以安全组为授权对象的规则，最多只能20条。 |
| 400 | InvalidParam.DestCidrIp | The specified parameter %s is not valid. | 指定的参数DestCidrIp不合法。 |
| 400 | InvalidParam.SourceCidrIp | The specified param SourceCidrIp is not valid. | 参数SourceCidrIp不合法。 |
| 400 | InvalidParam.Permissions | The specified parameter Permissions cannot coexist with other parameters. | 指定的参数Permissions不能与其他参数同时存在。 |
| 400 | InvalidParam.DuplicatePermissions | There are duplicate permissions in the specified parameter Permissions. | 指定的参数Permissions中存在重复的规则。 |
| 400 | InvalidGroupParameter.OperationDenied | The attributes Policy, SourceGroupId, DestGroupId of enterprise level security groups are not allowed to be set or modified. | 企业级安全组不能指定Policy，SourceGroupId，DestGroupId等参数。 |
| 400 | InvalidParam.ProtocolNotSupportPortRangeList | The specified protocol does not support the port range list. | 指定的协议不支持端口列表。 |
| 400 | InvalidPortRangeListId.NotFound | The specified port range list was not found. | 未找到指定的端口列表。 |
| 401 | InvalidOperation.SecurityGroupNotAuthorized | The specified security group is not authorized to operate. | 没有权限操作当前安全组 |
| 500 | InternalError | The request processing has failed due to some unknown error. |     |
| 403 | InvalidDestGroupId.Mismatch | NicType is required or NicType expects intranet. | 请指定 NicType，或使用内网模式。 |
| 403 | MissingParameter | The input parameter DestGroupId or DestCidrIp cannot be both blank. | 参数 DestGroupId 和 DestCidrIp 不得为空。 |
| 403 | AuthorizationLimitExceed | The limit of authorization records in the security group reaches. |     |
| 403 | InvalidParamter.Conflict | The specified SecurityGroupId should be different from the SourceGroupId. | 授权与被授权安全组必须不同。 |
| 403 | InvalidNetworkType.Conflict | The specified SecurityGroup network type should be same with SourceGroup network type (vpc or classic). | 指定的 SecurityGroup 的网络类型必须与 SouceGroup 的网络类型一致。 |
| 403 | InvalidSecurityGroup.IsSame | The authorized SecurityGroupId should be different from the DestGroupId. | 已授权的 SecurityGroupId 不能与 DestGroupId 相同。 |
| 403 | InvalidOperation.ResourceManagedByCloudProduct | %s  | 云产品托管的安全组不支持修改操作。 |
| 403 | LimitExceed.PrefixListAssociationResource | The number of resources associated with the prefix list exceeds the limit. | 与前缀列表关联的资源数量超出限制。 |
| 404 | InvalidSecurityGroupId.NotFound | The specified SecurityGroupId does not exist. | 指定的安全组在该用户账号下不存在，请您检查安全组 ID 是否正确。 |
| 404 | InvalidDestGroupId.NotFound | The DestGroupId provided does not exist in our records. | 指定的出方向安全组不存在。 |
| 404 | InvalidPrefixListId.NotFound | The specified prefix list was not found. | 前缀列表不存在。 |

访问[错误中心](https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/AuthorizeSecurityGroupEgress#workbench-doc-change-demo)。