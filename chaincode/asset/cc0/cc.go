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
