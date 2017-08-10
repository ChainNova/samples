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

// Asset ...
type Asset struct {
	Issuer string `json:"issuer"` //资产发行机构
	Code   string `json:"code"`   //资产代码
	Amount int64  `json:"amount"` //资产数量
}

// Account ...
type Account struct {
	AccountId string   `json:""accountId` //帐户id
	Assets    []*Asset `json:"assets"`    //该帐户的资产列表
}

// Init ...
func (c *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### Init chaincode ###########")

	return shim.Success(nil)
}

// Invoke ...
func (c *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### Invoke chaincode ###########")
	_, args := stub.GetFunctionAndParameters()
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

func (c *SimpleChaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== create account ==========")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 1")
	}

	var prarm struct {
		AccountId string `json:"accountId"` //帐户id
	}

	err := json.Unmarshal([]byte(args[0]), &prarm)
	if prarm.AccountId == "" || err != nil {
		fmt.Println("create account arguments error: AccountId can't be nil.")
		return shim.Error("create account arguments error: AccountId can't be nil.")
	}

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
	err = c.save(stub, a.AccountId, a)
	if err != nil {
		e := fmt.Sprintf("save account=%+v error:%s", a, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(nil)
}

func (c *SimpleChaincode) addAsset(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== addAsset asset ==========")
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 2")
	}

	accountId := args[0]

	var addAsset struct {
		Asset *Asset `json:"asset"` //资产
	}
	err := json.Unmarshal([]byte(args[1]), &addAsset)
	if accountId == "" || addAsset.Asset.Issuer == "" || addAsset.Asset.Code == "" || err != nil || addAsset.Asset.Amount <= 0 {
		fmt.Println("add asset arguments error: accountId, issuer and code can't be nil; amount must be a number and greater than 0.")
		return shim.Error("add asset arguments error: accountId, issuer and code can't be nil; amount must be a number and greater than 0.")
	}

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

	err = c.save(stub, account.AccountId, account)
	if err != nil {
		e := fmt.Sprintf("save account=%+v error:%s", account, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(nil)
}

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
	err := json.Unmarshal([]byte(args[1]), &transferAsset)
	if fromID == "" || transferAsset.AccountId == "" || transferAsset.Asset.Issuer == "" || transferAsset.Asset.Code == "" || err != nil || transferAsset.Asset.Amount <= 0 {
		fmt.Println("transfer asset arguments error: account, issuer and code can't be nil; amount must be a number and greater than 0.")
		return shim.Error("transfer asset arguments error: account, issuer and code can't be nil; amount must be a number and greater than 0.")
	}

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
	for k, v := range accountT.Assets {
		if v.Issuer == transferAsset.Asset.Issuer && v.Code == transferAsset.Asset.Code {
			accountF.Assets[k].Amount = v.Amount + transferAsset.Asset.Amount
			find = true
		}
	}
	if !find {
		accountT.Assets = append(accountT.Assets, transferAsset.Asset)
	}

	err = c.save(stub, accountF.AccountId, accountF)
	if err != nil {
		e := fmt.Sprintf("save account=%+v error:%s", accountF, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	err = c.save(stub, accountT.AccountId, accountT)
	if err != nil {
		e := fmt.Sprintf("save account=%+v error:%s", accountT, err)
		fmt.Println(e)
		return shim.Error(e)
	}
	return shim.Success(nil)
}

func (c *SimpleChaincode) getAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== getAccount ==========")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 1")
	}

	var prarm struct {
		AccountId string `json:"accountId"` //帐户id
	}

	err := json.Unmarshal([]byte(args[0]), &prarm)
	if prarm.AccountId == "" || err != nil {
		fmt.Println("create account arguments error: AccountId can't be nil.")
		return shim.Error("create account arguments error: AccountId can't be nil.")
	}

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
