# 资产转移chaincode

Chaincode又称智能合约，其与特定的业务应用场景相关。本例中的智能合约将模拟实现以下业务场景。

* 小张去某数字资产交易所开户。
* 开户后小张购买AAA机构发行的A1数字资产100股及BBB机构发行的B1数字资产200股。
* 小张欲将AAA机构的A1数字资产转移50股给小王。

## 基本术语

* 帐户：存储与帐户相关联的信息，如资产。
* 资产：各类数字资产的总称，每个帐户都可以拥有多种资产。
* 增加资产：对资产进行初始化，若帐户中已存在此相同资产则进行累加。
* 转移资产：将帐户中的某类资产转移到另一帐户。
* 帐户查询：查询帐户的资产信息。

## 结构定义

### 帐户

	type Account struct {	
      AccountId    string //帐户id
      Assets       []*Asset  //该帐户的资产列表
	}

在fabric底层的【key:value】存储中以AccountId作为key, Account作为value存储的。

### 资产
	
	type Asset struct { 
	    Issuer     string //资产发行机构
    	Code       string //资产代码
   		Amount   int64 //资产数量
	}

### 创建帐户

	type CreateAccount struct {
    	AccountId string //帐户id
	}

帐户创建过程：

1. 创建一个Account对象。
2. 将AccountId值赋给Account对象的AccountId。
3. 以AccountId为key,Account为value进行存储。

### 增加资产

	type AddAsset struct {
     	Asset *Asset //资产
	}

相同资产认定条件：

1. 同一发行机构的资产代码相同属相同资产。
2. 同一发行机构的不同资产代码不属于相同资产。
3. 不同发行机构的资产代码相同不属于相同资产。

资产发行规则：

1. 该帐户下存在该资产则进行数量累加。
2. 该帐户下不存在该相相同同资产则在该帐户的资产列表中增加一类新资产。


### 转移资产

	type TransferAsset struct {
     	AccountId string //转移目的帐号
     	Asset *Asset //欲转移的资产
	}

转移规则：

1. 转移目的账号必须存在。
2. 转移方必须存在欲转移的资产，且数量必须不少于欲转移的数量。
3. 对于接收方按发行资产的规则处理。


### 帐户查询

	type GetAccount struct {
     	AccountId string //查询的帐号id
	}

若帐号存在返回Account结构。

## Chaincode接口

* CreateAccount （创建帐户）

	调用参数：{“invoke”，"CreateAccount", CreateAccount}

* AddAsset （增加资产）

	调用参数：{“invoke”，“AddAsset”, “AccountId”， AddAsset}

* MoveAsset （资产转移）

	调用参数：{“invoke”，“TransferAsset”,“AccountId”, TransferAsset}

* GetAccount （查询帐户）

	调用参数：{“invoke”，“GetAccount”,GetAccount}

## 业务场景实现

以下是实现模拟场景的执行过程：

1. 小张去某数字资产交易所开户

		{“invoke”，"CreateAccount", CreateAccount{AccountId = xiaozhang}}

2. 开户后小张购买AAA机构发行的A1数字资产100股及BBB机构发行的B1数字资产200股

		{“invoke”，“AddAsset”, “xiaozhang”，AddAsset{Asset:Issuer=AAA,Code=A1,Amount=100}}

		{“invoke”，“AddAsset”, “xiaozhang”，AddAsset{Asset:Issuer=BBB,Code=B1,Amount=200}}


3. 小张欲将AAA机构的A1数字资产转移50股给小王。
		
		{“invoke”，“TransferAsset”,“xiaozhang”, TransferAsset {AccountId =xiaowang, Asset{Asset:Issuer=AAA,Code=A1,Amount=50}}

