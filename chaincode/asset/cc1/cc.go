package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode ...
type SimpleChaincode struct {
}

// Asset 资产
type Asset struct {
	Issuer string `json:"issuer"` //资产发行机构
	Code   string `json:"code"`   //资产代码
	Amount int64  `json:"amount"` //资产数量
}

// Account 账户
type Account struct {
	AccountId string   `json:""accountId` //帐户id
	Assets    []*Asset `json:"assets"`    //该帐户的资产列表
}

// Init ...
func (c *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### Init chaincode ###########")

	// Init中可以加一些初始化操作，比如初始化一种资产

	return shim.Success(nil)
}

// Invoke ...
func (c *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### Invoke chaincode ###########")
	_, args := stub.GetFunctionAndParameters()
	// 由于前面PPT中第一个参数总是“invoke”，真正的方法名是第二个参数。其实“invoke”不是必需的
	function := args[0]

	if function == "CreateAccount" {
		return c.createAccount(stub, args[1:])
	} else if function == "AddAsset" {
		return c.addAsset(stub, args[1:])
	} else if function == "TransferAsset" {
		return c.transferAsset(stub, args[1:])
	} else if function == "GetAccount" {
		return c.getAccount(stub, args[1:])
	}

	return shim.Error("Invalid invoke function name.")
}

// 创建账户
// 参数：账户信息（ID）
func (c *SimpleChaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== create account ==========")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 1")
	}

	var prarm struct {
		AccountId string `json:"accountId"` //帐户id
	}
	// 解析参数
	err := json.Unmarshal([]byte(args[0]), &prarm)
	if prarm.AccountId == "" || err != nil {
		fmt.Println("create account arguments error: AccountId can't be nil.")
		return shim.Error("create account arguments error: AccountId can't be nil.")
	}

	// 校验账户信息
	_, _, isExist, err := c.checkAccout(stub, prarm.AccountId)
	if err != nil {
		e := fmt.Sprintf("Check account=%s error:%s", prarm.AccountId, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if isExist {
		e := fmt.Sprintf("Account=%s already exists.", prarm.AccountId)
		fmt.Println(e)
		return shim.Error(e)
	}

	a := Account{
		AccountId: prarm.AccountId,
		Assets:    []*Asset{},
	}
	// 保存账户信息
	err = c.save(stub, a.AccountId, a)
	if err != nil {
		e := fmt.Sprintf("save account=%+v error:%s", a, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(nil)
}

// 添加资产
// 参数：添加资产信息
func (c *SimpleChaincode) addAsset(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== addAsset asset ==========")
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 2")
	}

	accountId := args[0]

	var addAsset struct {
		Asset *Asset `json:"asset"` //资产
	}
	// 解析参数
	err := json.Unmarshal([]byte(args[1]), &addAsset)
	if accountId == "" || addAsset.Asset.Issuer == "" || addAsset.Asset.Code == "" || err != nil || addAsset.Asset.Amount <= 0 {
		fmt.Println("add asset arguments error: accountId, issuer and code can't be nil; amount must be a number and greater than 0.")
		return shim.Error("add asset arguments error: accountId, issuer and code can't be nil; amount must be a number and greater than 0.")
	}

	// 获取并校验账户资产信息
	_, account, isExist, err := c.checkAccout(stub, accountId)
	if err != nil {
		e := fmt.Sprintf("Check account=%s error:%s", accountId, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if !isExist {
		e := fmt.Sprintf("Account=%s not exists.", accountId)
		fmt.Println(e)
		return shim.Error(e)
	}

	find := false
	// 判断是否存在该资产
	// 如果已有该资产，则数值增加
	// 如果没有，则加入该资产
	for k, v := range account.Assets {
		if v.Issuer == addAsset.Asset.Issuer && v.Code == addAsset.Asset.Code {
			account.Assets[k].Amount = v.Amount + addAsset.Asset.Amount
			find = true
			break
		}
	}
	if !find {
		account.Assets = append(account.Assets, addAsset.Asset)
	}

	// 保存账户资产
	err = c.save(stub, account.AccountId, account)
	if err != nil {
		e := fmt.Sprintf("save account=%+v error:%s", account, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(nil)
}

// 转移资产
// 参数：账户ID
//		转移资产信息（包括接收账户）
func (c *SimpleChaincode) transferAsset(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== transferAsset ==========")
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 2")
	}

	fromID := args[0]
	var transferAsset struct {
		AccountId string `json:"accountId"` //转移目的帐号
		Asset     *Asset `json:"asset"`     //欲转移的资产
	}
	// 解析参数
	err := json.Unmarshal([]byte(args[1]), &transferAsset)
	if fromID == "" || transferAsset.AccountId == "" || transferAsset.Asset.Issuer == "" || transferAsset.Asset.Code == "" || err != nil || transferAsset.Asset.Amount <= 0 {
		fmt.Println("transfer asset arguments error: account, issuer and code can't be nil; amount must be a number and greater than 0.")
		return shim.Error("transfer asset arguments error: account, issuer and code can't be nil; amount must be a number and greater than 0.")
	}

	// 获取并校验账户信息
	_, accountF, isExist, err := c.checkAccout(stub, fromID)
	if err != nil {
		e := fmt.Sprintf("Check account=%s error:%s", fromID, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if !isExist {
		e := fmt.Sprintf("Account=%s not exists.", fromID)
		fmt.Println(e)
		return shim.Error(e)
	}

	// 获取并校验接收账户信息
	_, accountT, isExist, err := c.checkAccout(stub, transferAsset.AccountId)
	if err != nil {
		e := fmt.Sprintf("Check account=%s error:%s", transferAsset.AccountId, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if !isExist {
		e := fmt.Sprintf("Account=%s not exists.", transferAsset.AccountId)
		fmt.Println(e)
		return shim.Error(e)
	}

	find := false
	// 检测账户资产
	// 如果存在，则减去转移量（必须确保转移量小于账户对应资产数量）
	// 如果不存在，则返回错误
	for k, v := range accountF.Assets {
		if v.Issuer == transferAsset.Asset.Issuer && v.Code == transferAsset.Asset.Code {
			if v.Amount < transferAsset.Asset.Amount {
				e := fmt.Sprintf("Account=%s issuer=%s&code=%s&count=%v < transfer count=%v.", accountF.AccountId, v.Issuer, v.Code, v.Amount, transferAsset.Asset.Amount)
				fmt.Println(e)
				return shim.Error(e)
			}
			accountF.Assets[k].Amount = v.Amount - transferAsset.Asset.Amount
			find = true
		}
	}
	if !find {
		e := fmt.Sprintf("Asset issuer=%s&code=%s of Account=%s not exists.", transferAsset.Asset.Issuer, transferAsset.Asset.Code, accountF.AccountId)
		fmt.Println(e)
		return shim.Error(e)
	}

	find = false
	// 判断接收账户资产
	// 如果存在该资产，则数量增加
	// 如果不存在该资产，则新增该资产
	for k, v := range accountT.Assets {
		if v.Issuer == transferAsset.Asset.Issuer && v.Code == transferAsset.Asset.Code {
			accountF.Assets[k].Amount = v.Amount + transferAsset.Asset.Amount
			find = true
		}
	}
	if !find {
		accountT.Assets = append(accountT.Assets, transferAsset.Asset)
	}

	// 保存账户信息
	err = c.save(stub, accountF.AccountId, accountF)
	if err != nil {
		e := fmt.Sprintf("save account=%+v error:%s", accountF, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	// 保存接收账户信息
	err = c.save(stub, accountT.AccountId, accountT)
	if err != nil {
		e := fmt.Sprintf("save account=%+v error:%s", accountT, err)
		fmt.Println(e)
		return shim.Error(e)
	}
	return shim.Success(nil)
}

// 获取用户信息
// 参数：查询账户信息
func (c *SimpleChaincode) getAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== getAccount ==========")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 1")
	}

	var prarm struct {
		AccountId string `json:"accountId"` //帐户id
	}
	// 解析参数
	err := json.Unmarshal([]byte(args[0]), &prarm)
	if prarm.AccountId == "" || err != nil {
		fmt.Println("create account arguments error: AccountId can't be nil.")
		return shim.Error("create account arguments error: AccountId can't be nil.")
	}

	// 获取账户信息
	b, _, isExist, err := c.checkAccout(stub, prarm.AccountId)
	if err != nil {
		e := fmt.Sprintf("Check account=%s error:%s", prarm.AccountId, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if !isExist {
		e := fmt.Sprintf("Account=%s not exists.", prarm.AccountId)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(b)
}

// 获取账户信息，并判断是否存在
func (c *SimpleChaincode) checkAccout(stub shim.ChaincodeStubInterface, id string) (b []byte, a Account, isExist bool, err error) {

	b, err = stub.GetState(id)
	if err != nil {
		return b, a, a.AccountId != "", err
	}
	if b != nil && len(b) > 0 {
		err = json.Unmarshal(b, &a)
	}

	return b, a, a.AccountId != "", err
}

// 保存state
func (c *SimpleChaincode) save(stub shim.ChaincodeStubInterface, k string, v interface{}) error {
	val, err := json.Marshal(v)
	if err != nil {
		return err
	}
	err = stub.PutState(k, val)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
