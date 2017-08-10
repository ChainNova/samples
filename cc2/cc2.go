package main

import (
	"encoding/json"
	"fmt"
	"strconv"

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
	Amount int    `json:"amount"` //资产数量
}

// Account ...
type Account struct {
	ID      string `json:"id"`      //帐户id
	Balance int    `json:"balance"` //账户余额
}

const (
	AssetObjectType        = "Asset~issuer~code"
	AccountAssetObjectType = "AccountAsset~id~issuer~code"
)

// Init ...
func (c *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### Init chaincode ###########")

	// init asset A1
	a1 := Asset{
		Issuer: "AAA",
		Code:   "A1",
		Amount: 10000,
	}
	_, _, isExist, key, err := c.checkAsset(stub, a1.Issuer, a1.Code)
	if err != nil {
		e := fmt.Sprintf("Check asset=%+v error:%s", a1, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if isExist {
		e := fmt.Sprintf("Asset=%+v already exists.", a1)
		fmt.Println(e)
		return shim.Error(e)
	}

	err = c.save(stub, key, a1)
	if err != nil {
		e := fmt.Sprintf("save asset=%+v error:%s", a1, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	// init asset B1
	b1 := Asset{
		Issuer: "BBB",
		Code:   "B1",
		Amount: 10000,
	}
	_, _, isExist, key, err = c.checkAsset(stub, b1.Issuer, b1.Code)
	if err != nil {
		e := fmt.Sprintf("Check asset=%+v error:%s", b1, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if isExist {
		e := fmt.Sprintf("Asset=%+v already exists.", b1)
		fmt.Println(e)
		return shim.Error(e)
	}

	err = c.save(stub, key, b1)
	if err != nil {
		e := fmt.Sprintf("save asset=%+v error:%s", b1, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(nil)
}

// Invoke ...
func (c *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### Invoke chaincode ###########")
	function, args := stub.GetFunctionAndParameters()

	if function == "CreateAccount" {
		return c.createAccount(stub, args)
	} else if function == "CreateAsset" {
		return c.createAsset(stub, args)
	} else if function == "Buy" {
		return c.buy(stub, args)
	} else if function == "Transfer" {
		return c.transfer(stub, args)
	} else if function == "AccountInfo" {
		return c.accountInfo(stub, args)
	} else if function == "AssetInfo" {
		return c.assetInfo(stub, args)
	} else if function == "MyAssets" {
		return c.myAssets(stub, args)
	} else if function == "IssuerAssets" {
		return c.issuerAssets(stub, args)
	}

	return shim.Error("Invalid invoke function name.")
}

func (c *SimpleChaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== create account ==========")
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 2")
	}

	id := args[0]
	balance, err := strconv.Atoi(args[1])
	if id == "" || err != nil || balance <= 0 {
		fmt.Println("create account arguments error: id can't be nil; balance must be a number and greater than 0.")
		return shim.Error("create account arguments error: id can't be nil; balance must be a number and greater than 0.")
	}

	_, _, isExist, err := c.checkAccout(stub, id)
	if err != nil {
		e := fmt.Sprintf("Check account=%s error:%s", id, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if isExist {
		e := fmt.Sprintf("Account=%s already exists.", id)
		fmt.Println(e)
		return shim.Error(e)
	}

	a := Account{
		ID:      id,
		Balance: balance,
	}
	err = c.save(stub, a.ID, a)
	if err != nil {
		e := fmt.Sprintf("save account=%+v error:%s", a, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(nil)
}

func (c *SimpleChaincode) createAsset(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== create asset ==========")
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 3")
	}

	issuer := args[0]
	code := args[1]
	amount, err := strconv.Atoi(args[2])
	if issuer == "" || code == "" || err != nil || amount <= 0 {
		fmt.Println("create asset arguments error: issuer and code can't be nil; amount must be a number and greater than 0.")
		return shim.Error("create asset arguments error: issuer and code can't be nil; amount must be a number and greater than 0.")
	}

	a := Asset{
		Issuer: issuer,
		Code:   code,
		Amount: amount,
	}

	_, _, isExist, key, err := c.checkAsset(stub, a.Issuer, a.Code)
	if err != nil {
		e := fmt.Sprintf("Check asset=%+v error:%s", a, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if isExist {
		e := fmt.Sprintf("Asset=%+v already exists.", a)
		fmt.Println(e)
		return shim.Error(e)
	}

	err = c.save(stub, key, a)
	if err != nil {
		e := fmt.Sprintf("save asset=%+v error:%s", a, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(nil)
}

func (c *SimpleChaincode) buy(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== buy ==========")
	if len(args) < 4 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 4")
	}

	id := args[0]
	issuer := args[1]
	code := args[2]
	count, err := strconv.Atoi(args[3])
	if id == "" || issuer == "" || code == "" || err != nil || count <= 0 {
		fmt.Println("buy asset arguments error: account, issuer and code can't be nil; count must be a number and greater than 0.")
		return shim.Error("buy asset arguments error: account, issuer and code can't be nil; count must be a number and greater than 0.")
	}

	_, account, isExist, err := c.checkAccout(stub, id)
	if err != nil {
		e := fmt.Sprintf("Check account=%s error:%s", id, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if !isExist {
		e := fmt.Sprintf("Account=%s not exists.", id)
		fmt.Println(e)
		return shim.Error(e)
	}

	_, asset, isExist, key, err := c.checkAsset(stub, issuer, code)
	if err != nil {
		e := fmt.Sprintf("Check asset issuer=%s&code=%s error:%s", issuer, code, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if !isExist {
		e := fmt.Sprintf("Asset issuer=%s&code=%s not exists.", issuer, code)
		fmt.Println(e)
		return shim.Error(e)
	}

	if account.Balance < count {
		e := fmt.Sprintf("Account balance=%v < buy count=%v.", account.Balance, count)
		fmt.Println(e)
		return shim.Error(e)
	}
	if asset.Amount < count {
		e := fmt.Sprintf("Asset amount=%v < buy count=%v.", asset.Amount, count)
		fmt.Println(e)
		return shim.Error(e)
	}

	account.Balance = account.Balance - count
	asset.Amount = asset.Amount - count

	err = c.save(stub, account.ID, account)
	if err != nil {
		e := fmt.Sprintf("save account=%+v error:%s", account, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	err = c.save(stub, key, asset)
	if err != nil {
		e := fmt.Sprintf("save asset=%+v error:%s", asset, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	sum, key, err := c.checkAccoutAsset(stub, account.ID, asset.Issuer, asset.Code)
	if err != nil {
		e := fmt.Sprintf("Check account=%s, asset issuer=%s&code=%s error:%s", account.ID, asset.Issuer, asset.Code, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	err = stub.PutState(key, []byte(strconv.Itoa(sum+count)))
	if err != nil {
		e := fmt.Sprintf("PutState error:%s", err)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(nil)
}

func (c *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== transfer ==========")
	if len(args) < 5 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 5")
	}

	from := args[0]
	to := args[1]
	issuer := args[2]
	code := args[3]
	count, err := strconv.Atoi(args[4])
	if from == "" || to == "" || issuer == "" || code == "" || err != nil || count <= 0 {
		fmt.Println("transfer asset arguments error: account, issuer and code can't be nil; amount must be a number and greater than 0.")
		return shim.Error("transfer asset arguments error: account, issuer and code can't be nil; amount must be a number and greater than 0.")
	}

	_, accountF, isExist, err := c.checkAccout(stub, from)
	if err != nil {
		e := fmt.Sprintf("Check account=%s error:%s", from, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if !isExist {
		e := fmt.Sprintf("Account=%s not exists.", from)
		fmt.Println(e)
		return shim.Error(e)
	}

	_, accountT, isExist, err := c.checkAccout(stub, to)
	if err != nil {
		e := fmt.Sprintf("Check account=%s error:%s", to, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if !isExist {
		e := fmt.Sprintf("Account=%s not exists.", to)
		fmt.Println(e)
		return shim.Error(e)
	}

	sumF, keyF, err := c.checkAccoutAsset(stub, accountF.ID, issuer, code)
	if err != nil {
		e := fmt.Sprintf("Check account=%s, asset issuer=%s&code=%s error:%s", accountF.ID, issuer, code, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	sumT, keyT, err := c.checkAccoutAsset(stub, accountT.ID, issuer, code)
	if err != nil {
		e := fmt.Sprintf("Check account=%s, asset issuer=%s&code=%s error:%s", accountT.ID, issuer, code, err)
		fmt.Println(e)
		return shim.Error(e)
	}

	if sumF < count {
		e := fmt.Sprintf("Account=%s issuer=%s&code=%s&count=%v < buy count=%v.", accountT.ID, issuer, code, sumF, count)
		fmt.Println(e)
		return shim.Error(e)
	}

	err = stub.PutState(keyF, []byte(strconv.Itoa(sumF-count)))
	if err != nil {
		e := fmt.Sprintf("PutState error:%s", err)
		fmt.Println(e)
		return shim.Error(e)
	}
	err = stub.PutState(keyT, []byte(strconv.Itoa(sumT+count)))
	if err != nil {
		e := fmt.Sprintf("PutState error:%s", err)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(nil)
}

func (c *SimpleChaincode) accountInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== accountInfo ==========")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 1")
	}

	id := args[0]
	b, _, isExist, err := c.checkAccout(stub, id)
	if err != nil {
		e := fmt.Sprintf("Check account=%s error:%s", id, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if !isExist {
		e := fmt.Sprintf("Account=%s not exists.", id)
		fmt.Println(e)
		return shim.Error(e)
	}

	return shim.Success(b)
}

func (c *SimpleChaincode) assetInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== assetInfo ==========")
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 2")
	}

	issuer := args[0]
	code := args[1]
	b, _, isExist, _, err := c.checkAsset(stub, issuer, code)
	if err != nil {
		e := fmt.Sprintf("Check asset issuer=%s&code=%s error:%s", issuer, code, err)
		fmt.Println(e)
		return shim.Error(e)
	} else if !isExist {
		e := fmt.Sprintf("Asset issuer=%s&code=%s not exists.", issuer, code)
		fmt.Println(e)
		return shim.Error(e)
	}
	return shim.Success(b)
}

func (c *SimpleChaincode) myAssets(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== myAsset ==========")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 1")
	}

	id := args[0]
	assetsIterator, err := stub.GetStateByPartialCompositeKey(AccountAssetObjectType, []string{id})
	if err != nil {

	}
	defer assetsIterator.Close()

	myAssets := struct {
		ID     string  `json:"id"`
		Assets []Asset `json:"assets"`
	}{ID: id}

	for assetsIterator.HasNext() {
		kv, _ := assetsIterator.Next()
		count, err := strconv.Atoi(string(kv.Value))
		if err != nil {
			fmt.Println("strconv.Atoi error:", err, string(kv.Value))
			continue
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(kv.Key)
		if err != nil {
			fmt.Println("SplitCompositeKey error:", err)
			continue
		}

		asset := Asset{
			Issuer: compositeKeyParts[1],
			Code:   compositeKeyParts[2],
			Amount: count,
		}
		myAssets.Assets = append(myAssets.Assets, asset)
	}

	b, err := json.Marshal(myAssets)
	if err != nil {
		e := fmt.Sprintf("Marshal myAssets error:%s", err)
		fmt.Println(e)
		return shim.Error(e)
	}
	return shim.Success(b)
}

func (c *SimpleChaincode) issuerAssets(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("=========== issuerAsset ==========")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting atleast 1")
	}

	issuer := args[0]
	assetsIterator, err := stub.GetStateByPartialCompositeKey(AssetObjectType, []string{issuer})
	if err != nil {
		fmt.Println("GetStateByPartialCompositeKey err:", err)
	}
	defer assetsIterator.Close()

	issuerAssets := struct {
		Issuer string  `json:"issuer"`
		Assets []Asset `json:"assets"`
	}{Issuer: issuer}

	for assetsIterator.HasNext() {
		kv, _ := assetsIterator.Next()

		var asset Asset
		err := json.Unmarshal(kv.Value, &asset)
		if err != nil {
			fmt.Println("json.Unmarshal error:", err, string(kv.Value))
			continue
		}
		issuerAssets.Assets = append(issuerAssets.Assets, asset)
	}

	b, err := json.Marshal(issuerAssets)
	if err != nil {
		e := fmt.Sprintf("Marshal myAssets error:%s", err)
		fmt.Println(e)
		return shim.Error(e)
	}
	return shim.Success(b)
}

func (c *SimpleChaincode) checkAccout(stub shim.ChaincodeStubInterface, id string) (b []byte, a Account, isExist bool, err error) {

	b, err = stub.GetState(id)
	if err != nil {
		return b, a, a.ID != "", err
	}
	if b != nil && len(b) > 0 {
		err = json.Unmarshal(b, &a)
	}

	return b, a, a.ID != "", err
}

func (c *SimpleChaincode) checkAsset(stub shim.ChaincodeStubInterface, issuer, code string) (b []byte, a Asset, isExist bool, key string, err error) {
	key, err = stub.CreateCompositeKey(AssetObjectType, []string{issuer, code})
	if err != nil {
		return b, a, a.Code != "", key, err
	}
	b, err = stub.GetState(key)
	if err != nil {
		return b, a, a.Code != "", key, err
	}
	if b != nil && len(b) > 0 {
		err = json.Unmarshal(b, &a)
	}
	return b, a, a.Code != "", key, err
}

func (c *SimpleChaincode) checkAccoutAsset(stub shim.ChaincodeStubInterface, id, issuer, code string) (count int, key string, err error) {
	key, err = stub.CreateCompositeKey(AccountAssetObjectType, []string{id, issuer, code})
	b, err := stub.GetState(key)
	if err != nil {
		return count, key, err
	}
	if b != nil && len(b) > 0 {
		count, _ = strconv.Atoi(string(b))
	}
	return count, key, err
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
