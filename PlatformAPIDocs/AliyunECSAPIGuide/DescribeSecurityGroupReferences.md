本接口用于查询一个或多个指定安全组已经被授权的其他安全组列表信息。

## 接口说明

-   当您无法删除某一安全组（ [DeleteSecurityGroup](https://help.aliyun.com/zh/ecs/api-deletesecuritygroup) ）时，可以调用该接口查看指定的安全组是否已被其他安全组授权。如果指定的安全组已被授权，您可以通过 [RevokeSecurityGroup](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-revokesecuritygroup) 和 [RevokeSecurityGroupEgress](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-revokesecuritygroupegress) 删除对应的安全组规则来解除授权行为。
    

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeSecurityGroupReferences)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeSecurityGroupReferences)

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
| ecs:DescribeSecurityGroupReferences | get | \\*SecurityGroup `acs:ecs:{#regionId}:{#accountId}:securitygroup/{#securitygroupId}` | 无   | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| RegionId | string | 是   | 安全组所属地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| SecurityGroupId | array | 是   | 安全组 ID 数组。数组长度：0~10。 | sg-bp14vtedjtobkvi\\*\\*\\*\\* |
|     | string | 否   | 安全组 ID。 | sg-bp14vtedjtobkvi\\*\\*\\*\\* |

## **返回参数**

| **名称** | **类型** | **描述** | **示例值** |
| --- | --- | --- | --- |
|     | object |     |     |
| RequestId | string | 请求 ID。 | 473469C7-AA6F-4DC5-B3DB-A3DC0DE3\\*\\*\\*\\* |
| SecurityGroupReferences | object |     |     |
| SecurityGroupReference | array<object> | 安全组和被授权的安全组信息集合。 |     |
|     | array<object> | 安全组和被授权的安全组信息。 |     |
| SecurityGroupId | string | 安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| ReferencingSecurityGroups | object |     |     |
| ReferencingSecurityGroup | array<object> | 正在授权给这个安全组的其他安全组信息集合。 |     |
|     | object | 正在授权给这个安全组的其他安全组信息。 |     |
| SecurityGroupId | string | 其他安全组 ID。 | sg-bp67acfmxazb4j\\*\\*\\*\\* |
| AliUid | string | 其他安全组所属用户 ID。 | 123456\\*\\*\\*\\* |

## 示例

正常返回示例

`JSON`格式

```
{
  "RequestId": "473469C7-AA6F-4DC5-B3DB-A3DC0DE3****",
  "SecurityGroupReferences": {
    "SecurityGroupReference": [
      {
        "SecurityGroupId": "sg-bp67acfmxazb4p****",
        "ReferencingSecurityGroups": {
          "ReferencingSecurityGroup": [
            {
              "SecurityGroupId": "sg-bp67acfmxazb4j****",
              "AliUid": "123456****"
            }
          ]
        }
      }
    ]
  }
}
```

## 错误码

| **HTTP status code** | **错误码** | **错误信息** | **描述** |
| --- | --- | --- | --- |
| 400 | InvalidSecurityGroupId.Malformed | The specified parameter SecurityGroupId is essential and size should less than 10 | 参数 SecurityGroupId 不能为空，且要查询的个数不能大于 10 个。 |
| 404 | InvalidRegionId.NotFound | The RegionId provided does not exist in our records. | 地域信息错误 |
| 404 | InvalidSecurityGroupId.NotFound | The SecurityGroupId provided does not exist in our records. | 指定的安全组在该用户账号下不存在，请您检查安全组 ID 是否正确。 |

访问[错误中心](https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/DescribeSecurityGroupReferences#workbench-doc-change-demo)。