package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const allPartnersKey = "all-partners"
const earnPointsTransactionsKey = "earn-points-transactions"
const usePointsTransactionsKey = "use-points-transactions"

type CustomerLoyaltyContract struct {
	contractapi.Contract
}

func (c *CustomerLoyaltyContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	err := ctx.GetStub().PutState("instantiate", []byte("INIT-LEDGER"))
	if err != nil {
		return NiceErrorf("instantiate contract", err)
	}

	emptyPartner, _ := json.Marshal(make([]Partner, 0))
	ctx.GetStub().PutState(allPartnersKey, emptyPartner)
	if err != nil {
		return NiceErrorf(
			fmt.Sprintf("put %s", allPartnersKey),
			err,
		)
	}

	emptyPoints, _ := json.Marshal(make([]PointTransaction, 0))
	ctx.GetStub().PutState(earnPointsTransactionsKey, emptyPoints)
	if err != nil {
		return NiceErrorf(
			fmt.Sprintf("put %s", earnPointsTransactionsKey),
			err,
		)
	}
	ctx.GetStub().PutState(usePointsTransactionsKey, emptyPoints)
	if err != nil {
		return NiceErrorf(
			fmt.Sprintf("put %s", usePointsTransactionsKey),
			err,
		)
	}

	return nil
}

// MEMBER RELATE FUNCTION
type Member struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phoneNumber"`
	AccountNumber string `json:"accountNumber"`
	CardId        string `json:"cardId"`
	Points        int    `json:"points"`
}

func (c *CustomerLoyaltyContract) CreateMember(
	ctx contractapi.TransactionContextInterface,
	memberStr string,
) (*Member, error) {

	member := Member{}
	err := json.Unmarshal([]byte(memberStr), &member)
	if err != nil {
		return nil, NiceErrorf("Wrong member format", err)
	}
	memberJson, _ := json.Marshal(member)
	err = ctx.GetStub().PutState(member.AccountNumber, memberJson)
	if err != nil {
		return nil, NiceErrorf("create member %s", err)
	}

	return &member, nil
}

type Partner struct {
	Name   string `json:"name"`
	Id     string `json:"partnerId"`
	CardId string `json:"cardId"`
}

func (c *CustomerLoyaltyContract) CreatePartner(
	ctx contractapi.TransactionContextInterface,
	partnerStr string,
) (*Partner, error) {
	partner := Partner{}
	err := json.Unmarshal([]byte(partnerStr), &partner)
	if err != nil {
		return nil, NiceErrorf("Wrong partner format", err)
	}

	partnerJson, err := json.Marshal(partner)
	if err != nil {
		return nil, NiceErrorf("parse partner to json", err)
	}
	ctx.GetStub().PutState(partner.Id, partnerJson)

	rawAllPartners, err := ctx.GetStub().GetState(allPartnersKey)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("get %s", allPartnersKey), err)
	}

	allPartner := []Partner{}
	err = json.Unmarshal(rawAllPartners, &allPartner)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("parse %s json | %s", allPartnersKey, string(rawAllPartners)), err)
	}
	allPartner = append(allPartner, partner)
	fmt.Printf("new all partner is %v\n", allPartner)

	allPartnerJson, _ := json.Marshal(allPartner)
	err = ctx.GetStub().PutState(allPartnersKey, allPartnerJson)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("put new %s", allPartnersKey), err)
	}

	return &partner, nil
}

func (c *CustomerLoyaltyContract) GetState(
	ctx contractapi.TransactionContextInterface,
	key string,
) (string, error) {
	raw, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", NiceErrorf(fmt.Sprintf("can't get %s", key), err)
	}

	return string(raw), nil
}

type PointTransaction struct {
	TransactionId string    `json:"transactionId"`
	Member        string    `json:"member" binding:"required"`
	Partner       string    `json:"partner" binding:"required"`
	Points        int       `json:"points" binding:"required"`
	Timestamp     time.Time `json:"timestamps"`
}

func (c *CustomerLoyaltyContract) EarnPoints(
	ctx contractapi.TransactionContextInterface,
	earnPointsStr string,
) (*PointTransaction, error) {
	earnPoints := PointTransaction{}
	err := json.Unmarshal([]byte(earnPointsStr), &earnPoints)
	if err != nil {
		return nil, NiceErrorf("wrong earn points format", err)
	}

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return nil, NiceErrorf("get TX timestamp", err)
	}

	earnPoints.Timestamp = timestamp.AsTime()
	earnPoints.TransactionId = ctx.GetStub().GetTxID()

	rawMember, err := ctx.GetStub().GetState(earnPoints.Member)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("get member %s", earnPoints.Member), err)
	}

	member := Member{}
	json.Unmarshal(rawMember, &member)
	member.Points += earnPoints.Points
	memberJson, _ := json.Marshal(member)
	err = ctx.GetStub().PutState(earnPoints.Member, memberJson)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("update member %s in EarnPoints", earnPoints.Member), err)
	}

	rawEarnPointsTransactions, err := ctx.GetStub().GetState(earnPointsTransactionsKey)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("get %s", earnPointsTransactionsKey), err)
	}
	earnPointsTransactions := []PointTransaction{}
	json.Unmarshal(rawEarnPointsTransactions, &earnPointsTransactions)

	earnPointsTransactions = append(earnPointsTransactions, earnPoints)
	earnPointsTransactionsJson, _ := json.Marshal(earnPointsTransactions)

	err = ctx.GetStub().PutState(earnPointsTransactionsKey, earnPointsTransactionsJson)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("update %s", earnPointsTransactionsKey), err)
	}

	return &earnPoints, nil
}

func (c *CustomerLoyaltyContract) UsePoints(
	ctx contractapi.TransactionContextInterface,
	usePointsStr string,
) (*PointTransaction, error) {
	usePoints := PointTransaction{}
	err := json.Unmarshal([]byte(usePointsStr), &usePoints)
	if err != nil {
		return nil, NiceErrorf("wrong earn points format", err)
	}

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("get member %s", usePoints.Member), err)
	}

	usePoints.Timestamp = timestamp.AsTime()
	usePoints.TransactionId = ctx.GetStub().GetTxID()

	rawMember, err := ctx.GetStub().GetState(usePoints.Member)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("parse member %s", usePoints.Member), err)
	}
	member := Member{}
	json.Unmarshal(rawMember, &member)

	if member.Points < usePoints.Points {
		return nil, fmt.Errorf("member does not have sufficient points")
	}

	member.Points -= usePoints.Points
	memberJson, _ := json.Marshal(member)
	err = ctx.GetStub().PutState(usePoints.Member, memberJson)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("get member %s", usePoints.Member), err)
	}

	rawUsePointsTransactions, err := ctx.GetStub().GetState(usePointsTransactionsKey)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("get %s", usePointsTransactionsKey), err)
	}

	usePointsTransactions := []PointTransaction{}
	json.Unmarshal(rawUsePointsTransactions, &usePointsTransactions)

	usePointsTransactions = append(usePointsTransactions, usePoints)
	usePointsTransactionsJson, _ := json.Marshal(usePointsTransactions)

	err = ctx.GetStub().PutState(usePointsTransactionsKey, usePointsTransactionsJson)
	if err != nil {
		return nil, NiceErrorf(fmt.Sprintf("put new value for %s", usePointsTransactionsKey), err)
	}

	return &usePoints, nil
}

func (c *CustomerLoyaltyContract) EarnPointsTransactionsInfo(
	ctx contractapi.TransactionContextInterface,
	userType string,
	userId string,
) ([]PointTransaction, error) {
	userTransactions := []PointTransaction{}
	rawEarnTransactions, err := ctx.GetStub().GetState(earnPointsTransactionsKey)
	if err != nil {
		return []PointTransaction{}, NiceErrorf(fmt.Sprintf("get %s", earnPointsTransactionsKey), err)
	}

	earnTransactions := []PointTransaction{}
	err = json.Unmarshal(rawEarnTransactions, &earnTransactions)
	if err != nil {
		return []PointTransaction{}, NiceErrorf("parse earn points", err)
	}

	for _, transaction := range earnTransactions {
		if (userType == "member" && transaction.Member == userId) ||
			(userType == "partner" && transaction.Partner == userId) {
			userTransactions = append(userTransactions, transaction)
			continue
		}
	}

	return userTransactions, nil
}

func (c *CustomerLoyaltyContract) UsePointsTransactionsInfo(
	ctx contractapi.TransactionContextInterface,
	userType string,
	userId string,
) ([]PointTransaction, error) {
	userTransactions := []PointTransaction{}
	rawEarnTransactions, err := ctx.GetStub().GetState(usePointsTransactionsKey)
	if err != nil {
		return []PointTransaction{}, NiceErrorf(fmt.Sprintf("get %s", earnPointsTransactionsKey), err)
	}

	earnTransactions := []PointTransaction{}
	err = json.Unmarshal(rawEarnTransactions, &earnTransactions)
	if err != nil {
		return []PointTransaction{}, NiceErrorf("parse earn points", err)
	}

	for _, transaction := range earnTransactions {
		if (userType == "member" && transaction.Member == userId) ||
			(userType == "partner" && transaction.Partner == userId) {
			userTransactions = append(userTransactions, transaction)
			continue
		}
	}

	return userTransactions, nil
}

func NiceErrorf(Problem string, Cause error) error {
	errMessage := ""
	errMessage += fmt.Sprintf("failed to %s", Problem)
	if Cause != nil {
		errMessage += fmt.Sprintf("\n---> because %s\n", Cause)
	}
	return fmt.Errorf(errMessage)
}
