package Algorand

import (
	"errors"
	"fmt"
	"github.com/jffp113/Thesis_Client/Client"
	"github.com/jffp113/Thesis_Client/conf"
	"github.com/jffp113/go-algorand-sdk/client/algod"
	"github.com/jffp113/go-algorand-sdk/client/algod/models"
	"github.com/jffp113/go-algorand-sdk/crypto"
	"github.com/jffp113/go-algorand-sdk/future"
	"github.com/jffp113/go-algorand-sdk/mnemonic"
	"github.com/jffp113/go-algorand-sdk/types"
	"github.com/satori/go.uuid"
	"math/rand"
	"strings"
	"time"
)

const Memmonic = "fiscal interest rare inhale fiscal apple piano body cricket will citizen emerge need view goose solve level obey trip asthma hero range call ability fruit"
const FROM = "UOYMQREGR5RMAI3QBIJLUCFVNVFDH7CW66ITCOSKVUP5XGWIXCYNBVQ76A"
const TO = "XLEXJVZ525GL7X24FD4DMWSERHG3EQ3MLF4SDGIGXAE7TPZTX4AC3PGVRE"
const AMOUNT = 1000000

type AlgorandHandler struct {
	// Create an algod client
	clients []algod.Client
}

func (h *AlgorandHandler) InitHandler(config conf.Configuration) {
	for _,v := range config.Conf.ValidatorNodes {
		algodClient, err := algod.MakeClient(fmt.Sprintf("http://%s",v), config.Conf.Token)
		if err != nil {
			panic(err)
		}

		h.clients = append(h.clients,algodClient)
	}
}

func (h AlgorandHandler) DoRequest() Client.RequestStatus {
	var stats Client.RequestStatus

	_,cli := ChooseOne(h.clients)

	stats.StartTime = time.Now()
	err := performTransaction(cli)
	stats.EndTime = time.Now()

	if err == nil {
		stats.Success = true
	}

	if err != nil {
		fmt.Println(err)
	}

	//fmt.Println(err)

	return stats
}

func NewHandler() Client.Handler {
	return &AlgorandHandler{}
}

func performTransaction(c algod.Client) error{

	tx,err := CreateAlgoTransaction(c)

	if err != nil {
		return err
	}

	id,b,err :=SignTransaction(tx,types.GroupEnvelop{})

	return SendTransaction(c,b,id)
}

func CreateAlgoTransaction(c algod.Client) (types.Transaction, error){
	params,err := c.BuildSuggestedParams()
	if err != nil {
		return types.Transaction{},err
	}

	uuid := uuid.NewV4()
	return future.MakePaymentTxn(FROM, TO,AMOUNT,uuid.Bytes(),"",params)
}

func SignTransaction(tx types.Transaction, g types.GroupEnvelop) (string, []byte, error){
	pk,_ := mnemonic.ToPrivateKey(Memmonic)
	return crypto.SignTransactionWithGroupSignature(pk,tx,g)
}

func SendTransaction(c algod.Client,tx []byte, id string) error {
	_, err := c.SendRawTransaction(tx)

	if err != nil {
		return err
	}

	_,err = waitForConfirmation(id,&c,100000)
	return err
}

func ChooseOne(v []algod.Client) (int, algod.Client) {
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	pos := rnd.Intn(len(v))
	return pos, v[pos]
}

// Function that waits for a given txId to be confirmed by the network
func waitForConfirmation(txID string, client *algod.Client, timeout uint64) (models.Transaction, error) {
	var pt models.Transaction
	if client == nil || txID == "" || timeout < 0 {
		fmt.Printf("Bad arguments for waitForConfirmation")
		var msg = errors.New("Bad arguments for waitForConfirmation")
		return pt, msg

	}

	status, err := client.Status()
	if err != nil {
		fmt.Printf("error getting algod status: %s\n", err)
		var msg = errors.New(strings.Join([]string{"error getting algod status: "}, err.Error()))
		return pt, msg
	}
	startRound := status.LastRound + 1
	currentRound := startRound

	for currentRound < (startRound + timeout) {

		pt, err := client.PendingTransactionInformation(txID)
		if err != nil {
			fmt.Printf("error getting pending transaction: %s\n", err)
			var msg = errors.New(strings.Join([]string{"error getting pending transaction: "}, err.Error()))
			return pt, msg
		}
		if pt.ConfirmedRound > 0 {
			//fmt.Printf("Transaction "+txID+" confirmed in round %d\n", pt.ConfirmedRound)
			return pt, nil
		}
		if pt.PoolError != "" {
			fmt.Printf("There was a pool error, then the transaction has been rejected!")
			var msg = errors.New("There was a pool error, then the transaction has been rejected")
			return pt, msg
		}
		//fmt.Printf("waiting for confirmation\n")
		status, err = client.StatusAfterBlock(currentRound)
		currentRound++
	}
	msg := errors.New("Tx not found in round range")
	return pt, msg
}