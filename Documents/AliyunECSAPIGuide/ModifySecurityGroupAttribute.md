本接口用于修改一个指定安全组的名称或者描述信息。

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/ModifySecurityGroupAttribute)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/Ecs/2014-05-26/ModifySecurityGroupAttribute)

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
| ecs:ModifySecurityGroupAttribute | update | \\*SecurityGroup `acs:ecs:{#regionId}:{#accountId}:securitygroup/{#securitygroupId}` | 无   | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| SecurityGroupId | string | 是   | 安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| Description | string | 否   | 安全组描述信息。长度为 2~256 个英文或中文字符，不能以`http://`和`https://`开头。 默认值：空，不会进行修改。 | TestDescription |
| SecurityGroupName | string | 否   | 安全组名称。长度为 2~128 个字符，必须以大小写字母或中文开头，不能以`http://`和`https://`开头。支持 Unicode 中 letter 分类下的字符（其中包括英文、中文等）和数字。可以包含半角冒号（:）、下划线（\\_）、半角句号（.）或者短划线（-）。 默认值：空，不会进行修改。 | SecurityGroupTestName |
| RegionId | string | 是   | 地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |

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
| 400 | InvalidSecurityGroupName.Malformed | Specified security group name is not valid. | 指定的安全组名称格式不合法。请您按照规则进行配置：默认值为空，长度为 2-128 个英文或中文字符，必须以大小字母或中文开头，可包含数字，英文句号（.），下划线（\\_）或连字符（-），安全组名称会展示在控制台。不能以 http:// 和 https:// 开头。 |
| 400 | InvalidSecurityGroupDiscription.Malformed | Specified security group description is not valid. | 指定的安全组描述不合法。 |
| 400 | InvalidParameter | Invalid Parameter. |     |
| 403 | InvalidOperation.ResourceManagedByCloudProduct | %s  | 云产品托管的安全组不支持修改操作。 |
| 404 | InvalidSecurityGroupId.NotFound | The specified SecurityGroupId does not exist. | 指定的安全组在该用户账号下不存在，请您检查安全组 ID 是否正确。 |

访问[错误中心](https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/ModifySecurityGroupAttribute#workbench-doc-change-demo)。