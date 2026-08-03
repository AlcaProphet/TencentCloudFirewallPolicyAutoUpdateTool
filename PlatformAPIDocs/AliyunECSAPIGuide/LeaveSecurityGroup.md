本接口用于将一台ECS实例或一张弹性网卡移出指定的安全组。

## 接口说明

**说明**

该 API 已不推荐使用。推荐您调用 [ModifyInstanceAttribute](https://help.aliyun.com/zh/ecs/api-modifyinstanceattribute) 接口将 ECS 实例加入或移除安全组；调用 [ModifyNetworkInterfaceAttribute](https://help.aliyun.com/zh/ecs/api-modifynetworkinterfaceattribute) 接口将弹性网卡（ENI）加入或移除安全组。

**重要** 阿里云已于 2024 年 7 月 8 日对该接口进行了校验规则调整。当移除不在组内的实例或者网卡时，从返回成功调整为返回错误码：“InvalidSecurityGroupAssociation.NotFound”。请您及时做好错误码兼容，避免影响线上业务。

-   不支持同时将实例和弹性网卡移出一个安全组，即参数`InstanceId`和`NetworkInterfaceId`不能同时传值。
    
-   移出安全组之前，实例必须处于**已停止**（Stopped）或者**运行中**（Running）状态。
    
-   一台实例或者一张弹性网卡必须至少加入一个安全组，当该实例或者弹性网卡只加入了一个安全组时，移出请求失败报错。
    
-   移除不在组内的实例或者网卡时，移出请求失败报错。
    

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/LeaveSecurityGroup)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/Ecs/2014-05-26/LeaveSecurityGroup)

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
| ecs:LeaveSecurityGroup | update | \\*Instance `acs:ecs:{#regionId}:{#accountId}:instance/{#instanceId}` \\*SecurityGroup `acs:ecs:{#regionId}:{#accountId}:securitygroup/{#securitygroupId}` | 无   | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| SecurityGroupId | string | 是   | 安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| InstanceId | string | 否   | 实例 ID。 **说明** 当该参数传入值时，`NetworkInterfaceId`必须为空。 | i-bp67acfmxazb4p\\*\\*\\*\\* |
| NetworkInterfaceId | string | 否   | 弹性网卡 ID。 **说明** 当该参数传入值时，`InstanceId`必须为空。 | eni-bp13kd656hxambfe\\*\\*\\*\\* |
| RegionId | string | 否   | 地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 - 实例移出安全组的操作可以不指定地域 ID。 - 弹性网卡移出安全组的操作必须指定弹性网卡所在的地域 ID。 | cn-hangzhou |

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
| 400 | InvalidInstanceId.Malformed | The specified parameter "InstanceId" is not valid. |     |
| 400 | MissingParameter.RegionId | The specified RegionId should not be null. | RegionId是必选参数。 |
| 400 | InvalidOperation.InvalidEniState | %s  | 弹性网卡当前的状态不支持此操作。 |
| 400 | InvalidSecurityGroupAssociation.NotFound | %s. | 当前实例或者网卡不在安全组内 |
| 403 | InstanceLastSecurityGroup | The specified security group is the last security group for the instance. | 指定的安全组是实例的最后一个安全组。 |
| 403 | IncorrectInstanceStatus | The current status of the resource does not support this operation. |     |
| 403 | InstanceLockedForSecurity | The specified operation is denied as your instance is locked for security reasons. | 实例被安全锁定。 |
| 403 | InstanceNotInSecurityGroup | The instance not in the group. |     |
| 403 | InvalidOperation.ResourceManagedByCloudProduct | %s  | 云产品托管的安全组不支持修改操作。 |
| 403 | InvalidOperation.AtLeastInOneGroup | %s  | 指定的资源必须至少加入一个安全组。当指定资源只加入了一个安全组时，将调用失败并返回该错误码。 |
| 403 | InvalidOperation.EniServiceManaged | %s  | 操作无效。 |
| 403 | InvalidOperation.InvalidEniType | %s  | 当前弹性网卡的类型不支持此操作。 |
| 403 | InvalidParam.Malformed | %s  | 无效的参数。 |
| 403 | InvalidParam.EniIdAndInstanceId.Conflict | %s  | InstanceId 与 NetworkInterfaceId 两者冲突。两个参数不能同时被赋值。 |
| 404 | InvalidInstanceId.NotFound | The specified InstanceId does not exist. | 指定的实例ID无效。 |
| 404 | InvalidSecurityGroupId.NotFound | The specified SecurityGroupId does not exist. | 指定的安全组在该用户账号下不存在，请您检查安全组 ID 是否正确。 |
| 404 | InvalidEniId.NotFound | %s  |     |
| 404 | InvalidVSwitchId.NotFound | The specified virtual switch %s does not exist. | 指定的交换机不存在。 |
| 504 | RequestTimeout | The request encounters an upstream server timeout. |     |
| 409 | InvalidOperation.Conflict | Request was denied due to conflict with a previous request, please try again later. | 请求操作的资源与之前的请求冲突 |

访问[错误中心](https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/LeaveSecurityGroup#workbench-doc-change-demo)。