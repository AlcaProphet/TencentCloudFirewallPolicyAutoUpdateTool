本接口用于修改一个普通安全组的组内连通策略。

## 接口说明

-   企业级安全组不支持修改组内连通策略，默认**组内隔离**。
    
-   您可以通过 [DescribeSecurityGroupAttribute](https://help.aliyun.com/zh/ecs/api-describesecuritygroupattribute) 查询当前安全组组内连通策略。
    
-   安全组的组内连通策略是**组内互通**时，会忽略其他自定义访问规则，组内所有实例的内网保持默认连通。
    
-   安全组的组内连通策略是**组内隔离**时，在不添加其他访问规则的情况下，组内所有实例的内网默认不连通。但您可以自定义安全组规则改变内网状态，例如，您可以通过 [AuthorizeSecurityGroup](https://help.aliyun.com/zh/ecs/api-authorizesecuritygroup) 使安全组内的两台 ECS 实例网络互通。
    

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/ModifySecurityGroupPolicy)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/Ecs/2014-05-26/ModifySecurityGroupPolicy)

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
| ecs:ModifySecurityGroupPolicy | update | \\*SecurityGroup `acs:ecs:{#regionId}:{#accountId}:securitygroup/{#securitygroupId}` | 无   | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| SecurityGroupId | string | 是   | 安全组的 ID。 | sg-bp67acfmxazb4ph\\*\\*\\*\\* |
| RegionId | string | 是   | 安全组所属地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| InnerAccessPolicy | string | 是   | 安全组内的 ECS 实例之间的内网连通策略。取值范围： - Accept：互通。 - Drop：隔离。 **说明** 取值不区分大小写。 | Drop |
| ClientToken | string | 否   | 保证请求幂等性。从您的客户端生成一个参数值，确保不同请求间该参数值唯一。**ClientToken** 只支持 ASCII 字符，且不能超过 64 个字符。更多信息，请参见[如何保证幂等性](https://help.aliyun.com/zh/ecs/developer-reference/how-to-ensure-idempotence)。 | 123e4567-e89b-12d3-a456-426655440000 |

## **返回参数**

| **名称** | **类型** | **描述** | **示例值** |
| --- | --- | --- | --- |
|     | object |     |     |
| RequestId | string | 请求 ID。 | CEF72CEB-54B6-4AE8-B225-F876FF7BA984 |

## 示例

正常返回示例

`JSON`格式

```
{
  "RequestId": "CEF72CEB-54B6-4AE8-B225-F876FF7BA984"
}
```

异常返回示例

`JSON`格式

```
{
    "RequestId": "CEF72CEB-54B6-4AE8-B225-F876FF7BA984"
}
```

## 错误码

| **HTTP status code** | **错误码** | **错误信息** | **描述** |
| --- | --- | --- | --- |
| 400 | MissingParamter.RegionId | The RegionId should not be null. |     |
| 400 | InvalidSecurityGroupId.Malformed | The SecurityGroupId is invalid. Only letters, numbers and underscores are supported. Maximum length is 100 characters. | 指定的参数 SecurityGroupId 格式错误，该参数仅支持字母，数字和下划线且最大长度为 100 个字符。 |
| 400 | InvalidPolicy.Malformed | The Policy is invalid. Only 'Accept' and 'Drop' are supported. Ignore case. |     |
| 403 | InvalidOperation.ResourceManagedByCloudProduct | %s  | 云产品托管的安全组不支持修改操作。 |
| 404 | InvalidSecurityGroupId.NotFound | The specified SecurityGroupId does not exist. | 指定的安全组在该用户账号下不存在，请您检查安全组 ID 是否正确。 |
| 404 | InvalidParameter.InnerAccessPolicy | The InnerAccessPolicy attribute of enterprise level security group can't be modified. |     |

访问[错误中心](https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/ModifySecurityGroupPolicy#workbench-doc-change-demo)。