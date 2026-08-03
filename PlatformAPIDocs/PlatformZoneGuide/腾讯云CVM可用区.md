# 概念介绍

## 地域

地域（Region）是指腾讯云物理数据中心所在的地理区域。您可以在腾讯云官网了解产品 [地理区域部署情况](https://cloud.tencent.com/act/event/global-base?Is=home#/)。

腾讯云不同地域之间的资源物理隔离，保证不同地域间最大程度的稳定性和容错性。

不同地域之间的网络完全隔离，云产品**默认不能通过内网通信。**只能通过访问 [公网 IP](https://cloud.tencent.com/document/product/213/5224) 或 [云联网](https://cloud.tencent.com/document/product/877) 进行通信。

## 可用区

可用区（Available Zone）是指腾讯云在同一地域内电力和网络互相独立的物理数据中心。您可以通过 API 接口 [查询可用区列表](https://cloud.tencent.com/document/product/213/15707) 查看完整的可用区列表。

可用区之间故障相互隔离（大型灾害或者大型电力故障除外），避免机房设备故障带来的业务影响。

在相同地域内，支持跨可用区部署私有网络，不同可用区之间的同一私有网络内网互通，可直接使用 [内网 IP](https://cloud.tencent.com/document/product/213/5225) 访问。

## 地域与可用区的关系

一个地域由一个或多个可用区组成。同一地域内的不同可用区之间 VPC 内网互通。不同地域之间的可用区完全独立。

![](https://qcloudimg.tencent-cloud.cn/image/document/6553426288315f6b0321e2a1297cb853.png)

# 如何选择地域和可用区

关于选择地域和可用区时，您需要考虑以下几个因素：
<table>
<tr>
<td rowspan="1" colspan="1" ><strong>考虑因素</strong></td>

<td rowspan="1" colspan="1" ><strong>选择说明</strong></td>
</tr>

<tr>
<td rowspan="1" colspan="1" >地理位置</td>

<td rowspan="1" colspan="1" >- 用户和资源部署地域的距离越近，网络时延越低，访问速度越快。<br>- 建议您在购买云服务器时，选择最靠近您客户的地域，以降低访问时延、提高访问速度。</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >业务通信</td>

<td rowspan="1" colspan="1" >- 同业务内的多类云产品建议部署在同个地域的同个可用区，通过内网进行通信，降低访问时延、提高访问速度。<br>- 不同业务之间如果要求内网互通，请将其部署在同一地域。如果没有内网互通要求，可按需将其部署在不同地域。</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >容灾考虑</td>

<td rowspan="1" colspan="1" >- 建议您将业务至少部署在同一地域内，不同可用区的同一 VPC 中，保证可用区之间的故障隔离，实现跨可用区容灾。<br>- 不同可用区之间可能会有网络的通信延迟，需要结合业务的实际需求进行评估，在高可用和低延迟之间找到最佳平衡点。</td>
</tr>
</table>

## 

# 地域和可用区列表

## 中国
<table>
<tr>
<td rowspan="1" colspan="1" ><strong>地域</strong></td>

<td rowspan="1" colspan="1" ><strong>地域 ID</strong></td>

<td rowspan="1" colspan="1" ><strong>可用区数量</strong></td>

<td rowspan="1" colspan="1" ><strong>可用区</strong></td>

<td rowspan="1" colspan="1" ><strong>可用区 ID</strong></td>
</tr>

<tr>
<td rowspan="6" colspan="1" >华北地区（北京）</td>

<td rowspan="6" colspan="1" >ap-beijing</td>

<td rowspan="6" colspan="1" >6</td>

<td rowspan="1" colspan="1" >北京三区</td>

<td rowspan="1" colspan="1" >ap-beijing-3</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >北京四区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-beijing-4</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >北京五区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-beijing-5</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >北京六区</td>

<td rowspan="1" colspan="1" >ap-beijing-6</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >北京七区</td>

<td rowspan="1" colspan="1" >ap-beijing-7</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >北京八区</td>

<td rowspan="1" colspan="1" >ap-beijing-8</td>
</tr>

<tr>
<td rowspan="6" colspan="1" >华东地区（上海）</td>

<td rowspan="6" colspan="1" >ap-shanghai</td>

<td rowspan="6" colspan="1" >6</td>

<td rowspan="1" colspan="1" >上海二区</td>

<td rowspan="1" colspan="1" >ap-shanghai-2</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海三区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-3</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海四区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-4</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海五区</td>

<td rowspan="1" colspan="1" >ap-shanghai-5</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海八区</td>

<td rowspan="1" colspan="1" >ap-shanghai-8</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海九区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-9</td>
</tr>

<tr>
<td rowspan="3" colspan="1" >华东地区（南京）</td>

<td rowspan="3" colspan="1" >ap-nanjing</td>

<td rowspan="3" colspan="1" >3</td>

<td rowspan="1" colspan="1" >南京一区</td>

<td rowspan="1" colspan="1" >ap-nanjing-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >南京二区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-nanjing-2</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >南京三区</td>

<td rowspan="1" colspan="1" >ap-nanjing-3</td>
</tr>

<tr>
<td rowspan="5" colspan="1" >华南地区（广州）</td>

<td rowspan="5" colspan="1" >ap-guangzhou</td>

<td rowspan="5" colspan="1" >5</td>

<td rowspan="1" colspan="1" >广州三区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-guangzhou-3</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >广州四区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-guangzhou-4</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >广州五区</td>

<td rowspan="1" colspan="1" >ap-guangzhou-5</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >广州六区</td>

<td rowspan="1" colspan="1" >ap-guangzhou-6</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >广州七区</td>

<td rowspan="1" colspan="1" >ap-guangzhou-7</td>
</tr>

<tr>
<td rowspan="2" colspan="1" >西南地区（成都）</td>

<td rowspan="2" colspan="1" >ap-chengdu</td>

<td rowspan="2" colspan="1" >2</td>

<td rowspan="1" colspan="1" >成都一区</td>

<td rowspan="1" colspan="1" >ap-chengdu-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >成都二区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-chengdu-2</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >西南地区（重庆）</td>

<td rowspan="1" colspan="1" >ap-chongqing</td>

<td rowspan="1" colspan="1" >1</td>

<td rowspan="1" colspan="1" >重庆一区</td>

<td rowspan="1" colspan="1" >ap-chongqing-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >西北地区（中卫）</td>

<td rowspan="1" colspan="1" >ap-zhongwei</td>

<td rowspan="1" colspan="1" >1</td>

<td rowspan="1" colspan="1" >中卫一区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-zhongwei-1</td>
</tr>

<tr>
<td rowspan="3" colspan="1" >港澳台地区（中国香港）</td>

<td rowspan="3" colspan="1" >ap-hongkong</td>

<td rowspan="3" colspan="1" >3</td>

<td rowspan="1" colspan="1" >香港一区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-hongkong-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >香港二区</td>

<td rowspan="1" colspan="1" >ap-hongkong-2</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >香港三区</td>

<td rowspan="1" colspan="1" >ap-hongkong-3</td>
</tr>

<tr>
<td rowspan="2" colspan="1" >北京金融<br>仅限金融机构和企业申请开通</td>

<td rowspan="2" colspan="1" >ap-beijing-fsi	</td>

<td rowspan="2" colspan="1" >2</td>

<td rowspan="1" colspan="1" >北京金融一区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-beijing-fsi-1	</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >北京金融二区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-beijing-fsi-2	</td>
</tr>

<tr>
<td rowspan="4" colspan="1" >上海金融<br>仅限金融机构和企业申请开通</td>

<td rowspan="4" colspan="1" >ap-shanghai-fsi	</td>

<td rowspan="4" colspan="1" >4</td>

<td rowspan="1" colspan="1" >上海金融一区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-fsi-1	</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海金融二区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-fsi-2	</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海金融三区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-fsi-3	</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海金融四区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-fsi-4	</td>
</tr>

<tr>
<td rowspan="3" colspan="1" >深圳金融<br>仅限金融机构和企业申请开通</td>

<td rowspan="3" colspan="1" >ap-shenzhen-fsi	</td>

<td rowspan="3" colspan="1" >3</td>

<td rowspan="1" colspan="1" >深圳金融一区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shenzhen-fsi-1	</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >深圳金融二区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shenzhen-fsi-2	</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >深圳金融三区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shenzhen-fsi-3	</td>
</tr>

<tr>
<td rowspan="4" colspan="1" >上海自动驾驶云</td>

<td rowspan="4" colspan="1" >ap-shanghai-adc	</td>

<td rowspan="4" colspan="1" >4</td>

<td rowspan="1" colspan="1" >上海自动驾驶云一区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-adc-1	</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海自动驾驶云二区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-adc-2	</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海自动驾驶云三区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-adc-3	</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >上海自动驾驶云四区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-shanghai-adc-4	</td>
</tr>
</table>

> **说明：**可用区角标 * 标识其售卖状态与其他可用区不同，需要购买时请联系腾讯云商务或 [在线咨询](https://cloud.tencent.com/online-service?from=sales&source=PRESALE) 沟通购买。 
> 

## 其他国家和地区
<table>
<tr>
<td rowspan="1" colspan="1" ><strong>地域</strong></td>

<td rowspan="1" colspan="1" ><strong>地域 ID</strong></td>

<td rowspan="1" colspan="1" ><strong>可用区数</strong></td>

<td rowspan="1" colspan="1" ><strong>可用区</strong></td>

<td rowspan="1" colspan="1" ><strong>可用区 ID</strong></td>
</tr>

<tr>
<td rowspan="4" colspan="1" >亚太和中东（新加坡）<br>可覆盖亚太东南地区</td>

<td rowspan="4" colspan="1" >ap-singapore</td>

<td rowspan="4" colspan="1" >4</td>

<td rowspan="1" colspan="1" >新加坡一区</td>

<td rowspan="1" colspan="1" >ap-singapore-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >新加坡二区</td>

<td rowspan="1" colspan="1" >ap-singapore-2</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >新加坡三区</td>

<td rowspan="1" colspan="1" >ap-singapore-3</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >新加坡四区</td>

<td rowspan="1" colspan="1" >ap-singapore-4</td>
</tr>

<tr>
<td rowspan="3" colspan="1" >亚太和中东（雅加达）<br>可覆盖亚太东南地区</td>

<td rowspan="3" colspan="1" >ap-jakarta</td>

<td rowspan="3" colspan="1" >3</td>

<td rowspan="1" colspan="1" >雅加达一区</td>

<td rowspan="1" colspan="1" >ap-jakarta-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >雅加达二区</td>

<td rowspan="1" colspan="1" >ap-jakarta-2</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >雅加达三区<sup>*</sup></td>

<td rowspan="1" colspan="1" >ap-jakarta-3</td>
</tr>

<tr>
<td rowspan="2" colspan="1" >亚太和中东（首尔）<br>可覆盖亚太东北地区</td>

<td rowspan="2" colspan="1" >ap-seoul</td>

<td rowspan="2" colspan="1" >2</td>

<td rowspan="1" colspan="1" >首尔一区</td>

<td rowspan="1" colspan="1" >ap-seoul-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >首尔二区</td>

<td rowspan="1" colspan="1" >ap-seoul-2</td>
</tr>

<tr>
<td rowspan="2" colspan="1" >亚太和中东（东京）<br>可覆盖亚太东北地区</td>

<td rowspan="2" colspan="1" >ap-tokyo</td>

<td rowspan="2" colspan="1" >2</td>

<td rowspan="1" colspan="1" >东京一区</td>

<td rowspan="1" colspan="1" >ap-tokyo-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >东京二区</td>

<td rowspan="1" colspan="1" >ap-tokyo-2</td>
</tr>

<tr>
<td rowspan="2" colspan="1" >亚太和中东（曼谷）<br>可覆盖亚太东南地区</td>

<td rowspan="2" colspan="1" >ap-bangkok</td>

<td rowspan="2" colspan="1" >2</td>

<td rowspan="1" colspan="1" >曼谷一区</td>

<td rowspan="1" colspan="1" >ap-bangkok-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >曼谷二区</td>

<td rowspan="1" colspan="1" >ap-bangkok-2</td>
</tr>

<tr>
<td rowspan="2" colspan="1" >亚太和中东（沙特阿拉伯）<br>可覆盖中东地区</td>

<td rowspan="2" colspan="1" >me-saudi-arabia</td>

<td rowspan="2" colspan="1" >2</td>

<td rowspan="1" colspan="1" >利雅得一区</td>

<td rowspan="1" colspan="1" >me-saudi-arabia-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >利雅得二区</td>

<td rowspan="1" colspan="1" >me-saudi-arabia-2</td>
</tr>

<tr>
<td rowspan="2" colspan="1" >欧洲和美洲（圣保罗）<br>可覆盖南美地区</td>

<td rowspan="2" colspan="1" >sa-saopaulo</td>

<td rowspan="2" colspan="1" >1</td>

<td rowspan="2" colspan="1" >圣保罗一区</td>

<td rowspan="2" colspan="1" >sa-saopaulo-1</td>
</tr>

<tr></tr>

<tr>
<td rowspan="2" colspan="1" >欧洲和美洲（硅谷）<br>可覆盖美国西部</td>

<td rowspan="2" colspan="1" >na-siliconvalley</td>

<td rowspan="2" colspan="1" >2</td>

<td rowspan="1" colspan="1" >硅谷一区</td>

<td rowspan="1" colspan="1" >na-siliconvalley-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >硅谷二区</td>

<td rowspan="1" colspan="1" >na-siliconvalley-2</td>
</tr>

<tr>
<td rowspan="2" colspan="1" >欧洲和美洲（弗吉尼亚）<br>可覆盖美国东部地区</td>

<td rowspan="2" colspan="1" >na-ashburn</td>

<td rowspan="2" colspan="1" >2</td>

<td rowspan="1" colspan="1" >弗吉尼亚一区</td>

<td rowspan="1" colspan="1" >na-ashburn-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >弗吉尼亚二区</td>

<td rowspan="1" colspan="1" >na-ashburn-2</td>
</tr>

<tr>
<td rowspan="2" colspan="1" >欧洲和美洲（法兰克福）<br>可覆盖欧洲地区</td>

<td rowspan="2" colspan="1" >eu-frankfurt</td>

<td rowspan="2" colspan="1" >2</td>

<td rowspan="1" colspan="1" >法兰克福一区</td>

<td rowspan="1" colspan="1" >eu-frankfurt-1</td>
</tr>

<tr>
<td rowspan="1" colspan="1" >法兰克福二区</td>

<td rowspan="1" colspan="1" >eu-frankfurt-2</td>
</tr>
</table>

> **说明：**可用区角标 * 标识其售卖状态与其他可用区不同，需要购买时请联系腾讯云商务或 [在线咨询](https://cloud.tencent.com/online-service?from=sales&source=PRESALE) 沟通购买。 
> 

## 资源位置说明

云服务器相关资源存在地域和可用区属性，详情请参见下表：

|资源|资源 ID 格式|类型|说明|
|---------|---------|---------|---------|
|用户账号|不限|全球唯一|用户可以使用同一个账号访问腾讯云全球各地资源。|
|[SSH 密钥](https://cloud.tencent.com/document/product/213/6092)|skey-xxxxxxxx|全地域可用|用户可以使用 SSH 密钥绑定账号下任何地域的云服务器。|
|[CVM 实例](https://cloud.tencent.com/document/product/213/4939)|ins-xxxxxxxx|单可用区可用|用户只能在特定可用区下创建 CVM 实例。|
|[自定义镜像](https://cloud.tencent.com/document/product/213/4941#.E8.87.AA.E5.AE.9A.E4.B9.89.E9.95.9C.E5.83.8F.5B.5D(id.3Acustomos))|img-xxxxxxxx|单地域多可用区可用|用户可以创建实例的自定义镜像，并在同个地域的不同可用区下使用。需要在其他地域使用时请使用复制镜像功能将自定义镜像复制到其他地域下。|
|[弹性 IP](https://cloud.tencent.com/document/product/213/5733)|eip-xxxxxxxx|单地域多可用区可用|弹性 IP 地址在某个地域下创建，并且只能与同一地域的实例相关联。|
|[安全组](https://cloud.tencent.com/document/product/213/112610)|sg-xxxxxxxx|单地域多可用区可用|安全组在某个地域下创建，并且只能与同一地域的实例相关联。腾讯云为用户自动创建三条默认安全组。|
|[云硬盘](https://cloud.tencent.com/document/product/362/2345)|disk-xxxxxxxx|单可用区可用|用户只能在特定可用区下创建云硬盘，并且挂载在同一可用区的实例上。|
|[快照](https://cloud.tencent.com/document/product/362/5754)|snap-xxxxxxxx|单地域多可用区可用|为某块云硬盘创建快照后，用户可在该地域下使用该快照进行其他操作（如创建云硬盘等）。|
|[负载均衡](https://cloud.tencent.com/document/product/214/524)|clb-xxxxxxxx|单地域多可用区可用|负载均衡可以绑定单地域下不同可用区的云服务器进行流量转发。|
|[私有网络](https://cloud.tencent.com/document/product/215/20046)|vpc-xxxxxxxx|单地域多可用区可用|私有网络创建在某一地域下，可以在不同可用区下创建属于同一个私有网络的资源。|
|[子网](https://cloud.tencent.com/document/product/215/20046#.E5.AD.90.E7.BD.91)|subnet-xxxxxxxx|单可用区可用|VPC 内一个子网只能属于一个可用区|
|[路由表](https://cloud.tencent.com/document/product/215/39406)|rtb-xxxxxxxx|单地域多可用区可用|用户创建路由表时需要指定特定的私有网络，因此跟随私有网络的位置属性。|

## 

## 相关操作

### 将实例迁移到其他可用区

一个已经启动的实例是无法更改其可用区的，但是您可以通过其他方法把实例迁移至其他可用区。迁移过程包括从原始实例创建自定义镜像、使用自定义镜像在新可用区中启动实例以及更新新实例的配置。
1. 创建当前实例的自定义镜像。更多信息，请参见 [创建自定义镜像](https://cloud.tencent.com/document/product/213/4942)。

2. 如果当前实例的网络环境为私有网络且需要在迁移后保留当前私有 IP 地址，用户可以先删除当前可用区中的子网，然后在新可用区中用与原始子网相同的 IP 地址范围创建子网。需要注意的是，不包含可用实例的子网才可以被删除。因此，应该将在当前子网中的所有实例移至新子网。

3. 使用刚创建的自定义镜像在新的可用区中创建一个新实例。用户可以选择与原始实例相同的实例类型及配置，也可以选择新的实例类型及配置。更多信息，请参见 [创建实例](https://cloud.tencent.com/document/product/213/4855)。

4. 如果原始实例已关联弹性 IP 地址，则将其与旧实例解除关联并与新实例相关联。更多信息，请参见 [弹性 IP](https://cloud.tencent.com/document/product/213/5733)。

5. （可选）若原有实例为 [按量计费](https://cloud.tencent.com/document/product/213/2180) 类型，可选择销毁原始实例。更多信息，请参见 [销毁实例](https://cloud.tencent.com/document/product/213/4930)。若原有实例为 [包年包月](https://cloud.tencent.com/document/product/213/2180) 类型，可选择等待其过期并回收。

### 将镜像复制到其他地域

启动实例、查看实例等操作存在地域属性。如果您需要启动实例的镜像在当前地域不存在，需要将镜像复制到当前地域。更多信息，请参见 [复制镜像](https://cloud.tencent.com/document/product/213/4943)。

