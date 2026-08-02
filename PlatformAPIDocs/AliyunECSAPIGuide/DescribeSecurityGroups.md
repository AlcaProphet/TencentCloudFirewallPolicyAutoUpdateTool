本接口用于查询安全组基本信息列表，支持您通过地域、安全组ID、安全组类型等不同参数查询。

## 接口说明

-   **分页查询**：推荐您使用`MaxResults`与`NextToken`参数进行查询。
    -   当返回结果中没有`NextToken`时，表示该页为末页，不再有后续页。
        
    -   分页查询首页时，仅需设置`MaxResults`以限制返回信息的条目数，返回结果中的`NextToken`将作为查询后续页的凭证。
        
    -   查询后续页时，将`NextToken`参数设置为上一次返回结果中获取到的`NextToken`作为查询凭证，并设置`MaxResults`限制返回条目数。
        
-   通过阿里云 CLI 调用 API 时，不同数据类型的请求参数取值必须遵循一定的格式要求。更多信息，请参见 [CLI 参数格式说明](https://help.aliyun.com/zh/cli/understanding-command-line-parameters)。
    

## 调试

[您可以在OpenAPI Explorer中直接运行该接口，免去您计算签名的困扰。运行成功后，OpenAPI Explorer可以自动生成SDK代码示例。](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeSecurityGroups)

 [![](https://img.alicdn.com/tfs/TB16JcyXHr1gK0jSZR0XXbP8XXa-24-26.png) 调试](https://api.aliyun.com/api/Ecs/2014-05-26/DescribeSecurityGroups)

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
| ecs:DescribeSecurityGroups | get | SecurityGroup `acs:ecs:{#regionId}:{#accountId}:securitygroup/*` SecurityGroup `acs:ecs:{#regionId}:{#accountId}:securitygroup/{#securitygroupId}` | - ecs:tag - ecs:tag - ecs:tag - ecs:tag | 无   |

## 请求参数

| **名称** | **类型** | **必填** | **描述** | **示例值** |
| --- | --- | --- | --- | --- |
| RegionId | string | 是   | 地域 ID。您可以调用 [DescribeRegions](https://help.aliyun.com/zh/ecs/api-regions-describeregions) 查看最新的阿里云地域列表。 | cn-hangzhou |
| SecurityGroupIds | string | 否   | 安全组 ID 列表。一次最多支持 100 个安全组 ID，ID 之间用半角逗号（,）隔开，格式为 JSON 数组。 | \\["sg-bp67acfmxazb4p\\*\\*\\*\\*", "sg-bp67acfmxazb4p\\*\\*\\*\\*", "sg-bp67acfmxazb4p\\*\\*\\*\\*",....\\] |
| VpcId | string | 否   | 安全组所在的专有网络 ID。 | vpc-bp67acfmxazb4p\\*\\*\\*\\* |
| SecurityGroupType | string | 否   | 安全组类型。取值范围： - normal：普通安全组。 - enterprise：企业安全组。 **说明** 当不为该参数传值时，表示查询所有类型的安全组。 | normal |
| NextToken | string | 否   | 查询凭证（Token）。取值为上一次调用该接口返回的 NextToken 参数值，初次调用接口时无需设置该参数。 | e71d8a535bd9cc11 |
| MaxResults | integer | 否   | 分页查询时每页的最大条目数。一旦设置该参数，即表示使用`MaxResults`与`NextToken`组合参数的查询方式。 最大值为 100。 默认值为 10。 | 10  |
| NetworkType | string | 否   | 安全组的网络类型。取值范围： - vpc：专有网络。 - classic：经典网络。经典网络功能已下线，详情请参见[下线公告](https://help.aliyun.com/zh/ecs/end-of-sale-announcement-for-ecs-instances-of-the-classic-network-type)。 | vpc |
| SecurityGroupName | string | 否   | 安全组名称。 | SGTestName |
| IsQueryEcsCount | boolean | 否   | 是否查询安全组的容量信息。传 True 时，返回值中的`EcsCount`和`AvailableInstanceAmount`有效。 **说明** 该参数已废弃。 | null |
| ResourceGroupId | string | 否   | 安全组所在的企业资源组 ID。使用该参数过滤资源时，资源数量不能超过 1000 个。您可以调用 [ListResourceGroups](https://help.aliyun.com/zh/resource-management/api-listresourcegroups) 查询资源组列表。 **说明** 不支持默认资源组过滤。 | rg-bp67acfmxazb4p\\*\\*\\*\\* |
| Tag | array<object> | 否   | 标签列表。 |     |
|     | object | 否   | 标签列表。 |     |
| key | string | 否   | 安全组的标签键。 **说明** 为提高兼容性，建议您尽量使用 Tag.N.Key 参数。 | testkey |
| Key | string | 否   | 安全组的标签键。N 的取值范围为 1~20。 使用一个标签过滤资源，查询到该标签下的资源数量不能超过 1000 个；使用多个标签过滤资源，查询到同时绑定了多个标签的资源数量不能超过 1000 个。如果资源数量超过 1000 个，请使用 [ListTagResources](https://help.aliyun.com/zh/ecs/api-listtagresources) 接口进行查询。 | TestKey |
| Value | string | 否   | 安全组的标签值。N 的取值范围：1~20。 | TestValue |
| value | string | 否   | 安全组的标签值。 **说明** 为提高兼容性，建议您尽量使用 Tag.N.Value 参数。 | testvalue |
| DryRun | boolean | 否   | 是否只预检此次请求。取值范围： - true：发送检查请求，不会查询资源状况。检查项包括 AccessKey 是否有效、RAM 用户的授权情况和是否填写了必需参数。如果检查不通过，则返回对应错误。如果检查通过，会返回错误码 DryRunOperation。 - false：发送正常请求，通过检查后返回 2XX HTTP 状态码并直接查询资源状况。 默认值为 false。 | false |
| SecurityGroupId | string | 否   | 安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| FuzzyQuery | boolean | 否   | **说明** 该参数已废弃。 | null |
| PageNumber | integer | 否   | **说明** 该参数即将下线，推荐您使用 NextToken 与 MaxResults 完成分页查询操作。 | 1   |
| PageSize | integer | 否   | **说明** 该参数即将下线，推荐您使用 NextToken 与 MaxResults 完成分页查询操作。 | 10  |
| ServiceManaged | boolean | 否   | 是否为托管安全组。取值范围： - true：是托管安全组。 - false：不是托管安全组。 | false |

## **返回参数**

| **名称** | **类型** | **描述** | **示例值** |
| --- | --- | --- | --- |
|     | object |     |     |
| RequestId | string | 请求 ID。 | 473469C7-AA6F-4DC5-B3DB-A3DC0DE3C83E |
| RegionId | string | 安全组所属地域 ID。 | cn-hangzhou |
| NextToken | string | 本次调用返回的查询凭证（Token）。当使用 MaxResults 和 NextToken 方式进行分页查询，且该返回值为空时，表示无更多返回的数据信息。 | e71d8a535bd9cc11 |
| SecurityGroups | object |     |     |
| SecurityGroup | array<object> | 安全组信息集合。 |     |
|     | array<object> | 安全组信息。 |     |
| SecurityGroupId | string | 安全组 ID。 | sg-bp67acfmxazb4p\\*\\*\\*\\* |
| SecurityGroupName | string | 安全组名称。 | SGTestName |
| Description | string | 安全组描述信息。 | TestDescription |
| SecurityGroupType | string | 安全组类型。可能值： - normal：普通安全组。 - enterprise：企业安全组。 | normal |
| VpcId | string | 安全组所属的专有网络。 | vpc-bp67acfmxazb4p\\*\\*\\*\\* |
| CreationTime | string | 创建时间。按照[ISO 8601](https://help.aliyun.com/zh/ecs/developer-reference/iso-8601-time-format)标准表示，并需要使用 UTC 时间。格式为：yyyy-MM-ddThh:mmZ。 | 2021-08-31T03:12:29Z |
| EcsCount | integer | 安全组中已经容纳的私网 IP 数量，参见[安全组容量](https://help.aliyun.com/zh/ecs/user-guide/basic-security-groups-and-advanced-security-groups#section-kj9-e46-6v5)。 当入参 IsQueryEcsCount 传入 True 时，该参数返回值有效。 **说明** 该参数已废弃。返回值中的数量仅供参考，非实时一致。 | 0   |
| AvailableInstanceAmount | integer | 安全组中还可加入的私网 IP 数量，参见[安全组容量](https://help.aliyun.com/zh/ecs/user-guide/basic-security-groups-and-advanced-security-groups#section-kj9-e46-6v5)。 当入参 IsQueryEcsCount 传入 True 时，该参数返回值有效。 **说明** 该参数已废弃。返回值中的数量仅供参考，非实时一致。 | 0   |
| ResourceGroupId | string | 安全组所在的企业资源组 ID。 | rg-bp67acfmxazb4p\\*\\*\\*\\* |
| ServiceManaged | boolean | 安全组的使用者是否为云产品或虚商。 | false |
| ServiceID | integer | 安全组对应的虚商 ID。 | 12345678910 |
| Tags | object |     |     |
| Tag | array<object> | 安全组的标签集合。 |     |
|     | object | 安全组的标签。 |     |
| TagValue | string | 安全组的标签值。 | TestValue |
| TagKey | string | 安全组的标签键。 | TestKey |
| RuleCount | integer | 安全组中的规则数量。 | 100 |
| GroupToGroupRuleCount | integer | 安全组中授权安全组访问的规则数量。 | 5   |
| TotalCount | integer | 安全组的总数。当您使用`MaxResults`与`NextToken`参数查询时，不会返回该参数值。 | 20  |
| PageNumber | integer | 当前页码。 **说明** 该参数即将下线，推荐您使用 NextToken 与 MaxResults 完成分页查询操作。 | 1   |
| PageSize | integer | 每页行数。 **说明** 该参数即将下线，推荐您使用 NextToken 与 MaxResults 完成分页查询操作。 | 10  |

## 示例

正常返回示例

`JSON`格式

```
{
  "RequestId": "473469C7-AA6F-4DC5-B3DB-A3DC0DE3C83E",
  "RegionId": "cn-hangzhou",
  "NextToken": "e71d8a535bd9cc11",
  "SecurityGroups": {
    "SecurityGroup": [
      {
        "SecurityGroupId": "sg-bp67acfmxazb4p****",
        "SecurityGroupName": "SGTestName",
        "Description": "TestDescription",
        "SecurityGroupType": "normal",
        "VpcId": "vpc-bp67acfmxazb4p****",
        "CreationTime": "2021-08-31T03:12:29Z",
        "EcsCount": 0,
        "AvailableInstanceAmount": 0,
        "ResourceGroupId": "rg-bp67acfmxazb4p****",
        "ServiceManaged": false,
        "ServiceID": 12345678910,
        "Tags": {
          "Tag": [
            {
              "TagValue": "TestValue",
              "TagKey": "TestKey"
            }
          ]
        },
        "RuleCount": 100,
        "GroupToGroupRuleCount": 5
      }
    ]
  },
  "TotalCount": 20,
  "PageNumber": 1,
  "PageSize": 10
}
```

## 错误码

| **HTTP status code** | **错误码** | **错误信息** | **描述** |
| --- | --- | --- | --- |
| 400 | NotSupported.PageNumberAndPageSize | The parameters PageNumber and PageSize are currently not supported, please use NextToken and MaxResults instead. | 参数PageNumber和PageSize已经不受支持，请使用参数NextToken和MaxResults。 |
| 400 | InValidParameter.NextToken | The parameter NextToken is invalid. | 指定的参数NextToken不合法。 |
| 400 | MissingParameter.RegionId | The input parameter RegionId that is mandatory for processing this request is not supplied. | 参数 RegionId 不能为空。 |
| 400 | InvalidParameter.SecurityGroupType | The specified SecurityGroupType is not valid. | 指定的 SecurityGroupType 参数无效，请检查该参数取值是否正确。 |
| 400 | InvalidSecurityGroupId.Malformed | The specified parameter SecurityGroupId is not valid. | 指定的参数 SecurityGroupId 无效。 |
| 400 | InvalidSecurityGroupName.Malformed | The specified parameter SecurityGroupName is not valid. | 指定的安全组名称格式不合法。请您按照规则进行配置：默认值为空，长度为 2-128 个英文或中文字符，必须以大小字母或中文开头，可包含数字，英文句号（.），下划线（\\_）或连字符（-），安全组名称会展示在控制台。不能以 http:// 和 https:// 开头。 |
| 500 | InternalError | The request processing has failed due to some unknown error. |     |

访问[错误中心](https://api.aliyun.com/document/Ecs/2014-05-26/errorCode)查看更多错误码。

## **变更历史**

更多信息，参考[变更详情](https://api.aliyun.com/document/Ecs/2014-05-26/DescribeSecurityGroups#workbench-doc-change-demo)。