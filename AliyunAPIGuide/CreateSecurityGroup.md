本接口用于创建一个安全组。

## 接口说明

-   通过该接口创建的普通安全组组内连通策略默认是**组内互通**，您可以通过 [ModifySecurityGroupPolicy](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-modifysecuritygrouppolicy) 进行修改。
    
-   通过该接口创建的企业安全组组内连通策略默认是**组内隔离**，并且不能修改。
    
-   单个地域下，安全组的数量有限制，您最少可以创建 100 个安全组。详细信息，请参见[安全组使用限制](https://help.aliyun.com/zh/ecs/user-guide/limitations#SecurityGroupQuota1)。
    
-   创建专有网络 VPC 类型的安全组时，您必须指定参数 VpcId。
    

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/CreateSecurityGroup)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/Ecs/2014-05-26/CreateSecurityGroup)

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
| ecs:CreateSecurityGroup | create | \\*SecurityGroup `acs:ecs:{#regionId}:{#accountId}:securitygroup/*` \\*VPC `acs:vpc:{#regionId}:{#accountId}:vpc/{#vpcId}` | 无   | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| RegionId | string | 是   | 安全组所属地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| Description | string | 否   | 安全组描述信息。长度为 2~256 个英文或中文字符，不能以`http://`和`https://`开头。 默认值：空。 | testDescription |
| ClientToken | string | 否   | 保证请求幂等性。从您的客户端生成一个参数值，确保不同请求间该参数值唯一。**ClientToken** 只支持 ASCII 字符，且不能超过 64 个字符。更多详情，请参见[如何保证幂等性](https://help.aliyun.com/zh/ecs/developer-reference/how-to-ensure-idempotence)。 | 123e4567-e89b-12d3-a456-426655440000 |
| SecurityGroupName | string | 否   | 安全组名称。长度为 2~128 个字符，必须以大小写字母或中文开头，不能以`http://`和`https://`开头。支持 Unicode 中 letter 分类下的字符（其中包括英文、中文等）和数字。可以包含半角冒号（:）、下划线（\\_）、半角句号（.）或者短划线（-）。 | testSecurityGroupName |
| VpcId | string | 否   | 安全组所属 VPC ID。 | vpc-bp1opxu1zkhn00gzv\\*\\*\\*\\* |
| SecurityGroupType | string | 否   | 安全组类型，分为普通安全组与企业安全组。取值范围： - normal：普通安全组。 - enterprise：企业安全组。更多详情，请参见[企业安全组概述](https://help.aliyun.com/zh/document_detail/120621.html)。 默认值：normal。 | enterprise |
| ServiceManaged | boolean | 否   | 该参数暂未开放使用。 | false |
| ResourceGroupId | string | 否   | 安全组所在的企业资源组 ID。 | rg-bp67acfmxazb4p\\*\\*\\*\\* |
| Tag | array<object> | 否   | 安全组绑定的标签数组。数组长度：0~20。 |     |
|     | object | 否   | 安全组绑定的标签。 |     |
| key | string | 否   | 安全组的标签键。 **说明** 为提高兼容性，建议您尽量使用 Tag.N.Key 参数。 | null |
| Key | string | 否   | 安全组的标签键。 一旦传入该值，则不允许为空字符串。最多支持 128 个字符，不能以`aliyun`和`acs:`开头，不能包含`http://`或者`https://`。 | TestKey |
| Value | string | 否   | 安全组的标签值。 一旦传入该值，允许为空字符串。最多支持 128 个字符，不能包含`http://`或者`https://`。 | TestValue |
| value | string | 否   | 安全组的标签值。 **说明** 为提高兼容性，建议您尽量使用 Tag.N.Value 参数。 | null |

## **返回参数**

| **名称** | **类型** | **描述** | **示例值** |
| --- | --- | --- | --- |
|     | object |     |     |
| SecurityGroupId | string | 安全组 ID。 | sg-bp1fg655nh68xyz9\\*\\*\\*\\* |
| RequestId | string | 请求 ID。 | 473469C7-AA6F-4DC5-B3DB-A3DC0DE3C83E |

## 示例

正常返回示例

`JSON`格式

```
{
  "SecurityGroupId": "sg-bp1fg655nh68xyz9****",
  "RequestId": "473469C7-AA6F-4DC5-B3DB-A3DC0DE3C83E"
}
```

异常返回示例

`JSON`格式

```
{
    "RequestId":"CEF72CEB-54B6-4AE8-B225-F876FF7BA984",
    "SecurityGroupId":"sg-F876FF7BA"
}
```

## 错误码

| **HTTP status code** | **错误码** | **错误信息** | **描述** |
| --- | --- | --- | --- |
| 400 | InvalidDescription.Malformed | The specified parameter "Description" is not valid. | 指定的资源描述格式不合法。长度为2-256个字符，不能以http://和https://开头。 |
| 400 | InvalidSecurityGroupDiscription.Malformed | Specified security group description is not valid. | 指定的安全组描述不合法。 |
| 400 | IncorrectVpcStatus | Current VPC status does not support this operation. | 当前专有网络 VPC 的状态无法支持这个操作。 |
| 400 | InvalidTagKey.Malformed | Specified tag key is not valid. | 指定的标签的键格式错误。 |
| 400 | InvalidTagValue.Malformed | Specified tag value is not valid. | 指定的标签的值格式错误。 |
| 400 | Duplicate.TagKey | The Tag.N.Key contain duplicate key. | 标签中存在重复的键，请保持键的唯一性。 |
| 400 | InvalidParams.GroupType | The specified security group type is not valid. | 指定的安全组类型无效。请检查 SecurityGroupType 参数值是否正确。 |
| 400 | InvalidParams.VpcIdGroupType | Only VPC instance supports enterprise level security group. | 仅 VPC 网络类型的 ECS 实例支持企业级安全组。 |
| 400 | InvalidSecurityGroupName.Malformed | The specified parameter SecurityGroupName is not valid. | 指定的安全组名称格式不合法。请您按照规则进行配置：默认值为空，长度为 2-128 个英文或中文字符，必须以大小字母或中文开头，可包含数字，英文句号（.），下划线（\\_）或连字符（-），安全组名称会展示在控制台。不能以 http:// 和 https:// 开头。 |
| 500 | InternalError | The request processing has failed due to some unknown error. |     |
| 500 | ServiceUnavailable | The service is unavailable, please try again later. | 服务暂时不可用，请稍候重试。 |
| 403 | QuotaExceed.SecurityGroup | The maximum number of security groups is reached. | 安全组数量超过额度限制。 |
| 403 | InvalidVpcId.NotFound | The VpcId must not empty when only support vpc vm. | 指定的 VPC ID 不能为空。 |
| 403 | IncorrectVpcStatus | Current VPC status does not support this operation. |     |
| 403 | IdempotentProcessing | The previous idempotent request(s) is still processing. | 先前的幂等请求仍在处理中，请稍后重试。 |
| 403 | QuotaExceed.Tags | %s  | 标签数超过可以配置的最大数量。%s为变量，将根据调用API的实际情况动态返回错误信息。 |
| 403 | InvalidOperation.ResourceManagedByCloudProduct | %s  | 云产品托管的安全组不支持修改操作。 |
| 404 | InvalidRegionId.NotFound | The specified region does not exist. | 指定的 RegionId 不存在，请您检查此产品在该地域是否可用。 |
| 404 | InvalidVpcId.NotFound | Specified VPC does not exist. | 指定的专有网络 VPC ID 不存在。 |
| 404 | InvalidResourceGroup.NotFound | The ResourceGroup provided does not exist in our records. | 资源组并不在记录中。 |

访问[错误中心](https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/CreateSecurityGroup#workbench-doc-change-demo)。