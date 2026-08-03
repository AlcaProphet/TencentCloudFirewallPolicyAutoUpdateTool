本接口用于删除一个安全组，并关联删除组内所有安全组规则。

## 接口说明

-   请确保安全组内不存在 ECS 实例，您可以通过 [DescribeInstances](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describeinstances) 进行查询。
-   请确保安全组内不存在弹性网卡，您可以通过 [DescribeNetworkInterfaces](https://help.aliyun.com/zh/ecs/developer-reference/api-ecs-2014-05-26-describenetworkinterfaces) 进行查询。
-   请确保没有其他安全组与该安全组有授权行为，您可以通过 [DescribeSecurityGroupReferences](https://help.aliyun.com/zh/ecs/api-describesecuritygroupreferences) 进行查询。
-   在您使用该接口删除安全组时若返回错误码`InvalidOperation.DeletionProtection`，说明开启了删除保护功能。创建 ACK 集群时，关联的安全组会开启删除保护功能，来防止误删除。删除保护功能无法手动关闭，只有在删除了关联的 ACK 集群后，才能够自动关闭。更多信息，请参见[关闭安全组删除保护](https://help.aliyun.com/zh/ack/ack-managed-and-ack-dedicated/user-guide/configure-security-group-rules-to-enforce-access-control-on-ack-clusters)。

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/DeleteSecurityGroup)

[![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png)调试](https://api.aliyun.com/api/Ecs/2014-05-26/DeleteSecurityGroup)

## 授权信息

下表是API对应的授权信息，可以在RAM权限策略语句的`Action`元素中使用，用来给RAM用户或RAM角色授予调用此API的权限。具体说明如下：

-   操作：是指具体的权限点。
-   访问级别：是指每个操作的访问级别，取值为写入（Write）、读取（Read）或列出（List）。
-   资源类型：是指操作中支持授权的资源类型。具体说明如下：
    -   对于必选的资源类型，用前面加 \* 表示。
    -   对于不支持资源级授权的操作，用`全部资源`表示。
-   条件关键字：是指云产品自身定义的条件关键字。
-   关联操作：是指成功执行操作所需要的其他权限。操作者必须同时具备关联操作的权限，操作才能成功。

| 操作  | 访问级别 | 资源类型 | 条件关键字 | 关联操作 |
| --- | --- | --- | --- | --- |
| ecs:DeleteSecurityGroup | delete | \\*SecurityGroup `acs:ecs:{#regionId}:{#accountId}:securitygroup/{#securitygroupId}` | 无   | 无   |

## 请求参数

| 名称  | 类型  | 必填  | 描述  | 示例值 |
| --- | --- | --- | --- | --- |
| RegionId | string | 是   | 地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| SecurityGroupId | string | 是   | 安全组 ID。您可以调用 [DescribeSecurityGroups](https://help.aliyun.com/zh/ecs/api-describesecuritygroups) 查看安全组 ID。 | sg-bp1fg655nh68xyz9\\*\\*\\*\\* |

## 返回参数

| 名称  | 类型  | 描述  | 示例值 |
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

| HTTP status code | 错误码 | 错误信息 | 描述  |
| --- | --- | --- | --- |
| 400 | MissingParameter.RegionId | The parameter "RegionId" should not be null. | \\- |
| 401 | InvalidOperation.SecurityGroupNotAuthorized | The specified security group is not authorized to operate. | 没有权限操作当前安全组 |
| 403 | DependencyViolation | There is still instance(s) in the specified security group. | 安全组中还有未释放的实例，请您先释放实例再进行该操作。 |
| 403 | DependencyViolation | The specified security group has been authorized in another one. | 指定的安全组被另一个安全组的规则引用，不允许删除。 |
| 403 | InvalidOperation.ResourceManagedByCloudProduct | %s  | 云产品托管的安全组不支持修改操作。 |
| 404 | InvalidSecurityGroup.NotFound | The specified security group is not found. | 找不到指定的安全组 |
| 500 | InternalError | The request processing has failed due to some unknown error. | 内部错误，请重试。 |

访问[错误中心]( https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## 变更历史

| 变更时间 | 变更内容概要 | 操作  |
| --- | --- | --- |
| 2025-03-12 | OpenAPI 描述信息更新、OpenAPI 错误码发生变更 | [查看变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/DeleteSecurityGroup?updateTime=2025-03-12#workbench-doc-change-demo) |
| 2024-01-03 | OpenAPI 错误码发生变更 | [查看变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/DeleteSecurityGroup?updateTime=2024-01-03#workbench-doc-change-demo) |