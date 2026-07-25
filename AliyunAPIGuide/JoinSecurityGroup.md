本接口用于将一台ECS实例或一张弹性网卡加入到指定的安全组。

## 接口说明

**说明**

该 API 已不推荐使用。推荐您调用 [ModifyInstanceAttribute](https://help.aliyun.com/zh/ecs/api-modifyinstanceattribute) 接口将 ECS 实例加入或移除安全组；调用 [ModifyNetworkInterfaceAttribute](https://help.aliyun.com/zh/ecs/api-modifynetworkinterfaceattribute) 接口将弹性网卡（ENI）加入或移除安全组。

-   该接口不支持同时将实例和弹性网卡加入一个安全组，即参数`InstanceId`和`NetworkInterfaceId`不能同时传值。
    
-   您的安全组和实例必须属于同一个阿里云地域。
    
-   您的安全组和实例的网络类型必须相同。如果网络类型为专有网络 VPC，则安全组和实例必须属于同一个 VPC。
    
-   加入安全组之前，实例必须处于**已停止**（Stopped）或者**运行中**（Running）状态。
    
-   一台实例或一张弹性网卡最多可以加入五个安全组。更多信息，请参见[安全组使用限制](https://help.aliyun.com/zh/ecs/user-guide/limitations#SecurityGroupQuota1)。
    

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/JoinSecurityGroup)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/Ecs/2014-05-26/JoinSecurityGroup)

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
| ecs:JoinSecurityGroup | update | \\*Instance `acs:ecs:{#regionId}:{#accountId}:instance/{#instanceId}` \\*SecurityGroup `acs:ecs:{#regionId}:{#accountId}:securitygroup/{#securitygroupId}` | 无   | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| SecurityGroupId | string | 是   | 安全组 ID。您可以调用 [DescribeSecurityGroups](https://help.aliyun.com/zh/ecs/api-describesecuritygroups) 查看您可用的安全组。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| InstanceId | string | 否   | 实例 ID。 **说明** 当该参数传入值时，`NetworkInterfaceId`必须为空。 | i-bp67acfmxazb4p\\*\\*\\*\\* |
| NetworkInterfaceId | string | 否   | 弹性网卡 ID。 **说明** 当该参数传入值时，`InstanceId`必须为空。 | eni-bp13kd656hxambfe\\*\\*\\*\\* |
| RegionId | string | 否   | 地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 - 实例加入安全组的操作可以不指定地域 ID。 - 弹性网卡加入安全组的操作必须指定弹性网卡所在的地域 ID。 | cn-hangzhou |

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
    "RequestId": "473469C7-AA6F-4DC5-B3DB-A3DC0DE3C83E"
}
```

## 错误码

| **HTTP status code** | **错误码** | **错误信息** | **描述** |
| --- | --- | --- | --- |
| 400 | InstanceSecurityGroupLimitExceeded | Exceeding the allowed amount of security groups that an instance can be in. | 加入安全组失败，该实例加入的安全组数量已达到上限。 |
| 400 | InvalidInstanceId.Mismatch | Specified instance and security group are not in the same VPC. | 指定的实例和安全组不属于同一个虚拟专有网络。（包含另外两种特殊情况：1.实例不属于VPC类型，安全组属于VPC类型；2.实例属于VPC类型，安全组不属于VPC类型。） |
| 400 | InvalidInstanceId.Malformed | The specified parameter "InstanceId" is not valid. |     |
| 400 | InvalidOperation.NotSupportEnterpriseGroup | The specified instance type doesn't support enterprise level security group. |     |
| 400 | InvalidOperation.MultiGroupType | The specified instance can't join different types of security group. |     |
| 400 | InvalidOperation.InvalidEniState | %s  | 弹性网卡当前的状态不支持此操作。 |
| 400 | InvalidOperation.EniAndGroupNotBelongSameUser | %s  |     |
| 400 | NotBelongUser | %s  | 您没有权限操作此资源，请检查操作的资源是否正确。 |
| 400 | MissingParameter.RegionId | The specified RegionId should not be null. | RegionId是必选参数。 |
| 400 | InvalidStatus.EniOrInstanceIsBeingCreated | %s. | 指定的实例或者网卡正在创建中，请等待实例或网卡创建完成后重试。 |
| 500 | InternalError | The request processing has failed due to some unknown error. |     |
| 403 | IncorrectInstanceStatus | The current status of the resource does not support this operation. |     |
| 403 | InstanceLockedForSecurity | The specified operation is denied as your instance is locked for security reasons. | 实例被安全锁定。 |
| 403 | SecurityGroupInstanceLimitExceeded | The maximum number of instances in a security group is exceeded. | 该安全组内已有的实例数量已达到最大限制。 |
| 403 | InvalidInstanceId.AlreadyExists | The specified instance already exists in the specified security group. | 指定的实例已经在指定的安全组中。 |
| 403 | AclLimitExceed | %s  | 网卡或实例的安全组规则数量超过限额值。 |
| 403 | InstanceSecurityGroupLimitExceeded | %s  | 实例绑定的安全组数量达到最大限制。 |
| 403 | InvalidOperation.NetworkInterfaceCountExceeded | The maximum number of NetworkInterface in a enterprise level security group is exceeded. | 指定的企业级安全组中的网卡数量超限。 |
| 403 | InvalidOperation.ResourceManagedByCloudProduct | %s  | 云产品托管的安全组不支持修改操作。 |
| 403 | InvalidOperation.InvalidEniType | %s  | 当前弹性网卡的类型不支持此操作。 |
| 403 | InvalidOperation.VpcMismatch | %s  | 您的操作无效，请确认该操作中的 VPC 与其它参数是否匹配。 |
| 403 | InvalidOperation.EniServiceManaged | %s  | 操作无效。 |
| 403 | InvalidParam.Malformed | %s  | 无效的参数。 |
| 403 | InvalidParam.EniIdAndInstanceId.Conflict | %s  | InstanceId 与 NetworkInterfaceId 两者冲突。两个参数不能同时被赋值。 |
| 403 | Forbidden.InstanceIsBeingCreated | The specified instance is being created. | 指定的实例正在创建中。 |
| 404 | InvalidSecurityGroupId.NotFound | The specified SecurityGroupId does not exist. | 指定的安全组在该用户账号下不存在，请您检查安全组 ID 是否正确。 |
| 404 | InvalidInstanceId.NotFound | The specified InstanceId does not exist. | 指定的实例ID无效。 |
| 404 | InvalidEniId.NotFound | %s  |     |
| 409 | InvalidOperation.Conflict | Request was denied due to conflict with a previous request, please try again later. | 请求操作的资源与之前的请求冲突 |

访问[错误中心](https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/JoinSecurityGroup#workbench-doc-change-demo)。