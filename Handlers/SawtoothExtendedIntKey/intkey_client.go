/**
 * Copyright 2018 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * ------------------------------------------------------------------------------
 */

package SawoothExtendedIntKey

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	cbor "github.com/brianolson/cbor_go"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"github.com/jffp113/CryptoProviderSDK/keychain"
	"github.com/jffp113/Thesis_Client/Handlers/SawtoothBaseIntKey/protobuf/transaction_pb2"
	"github.com/jffp113/Thesis_Client/Handlers/SawtoothExtendedIntKey/pb"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// String literals
	FAMILY_NAME       string = "intkey"
	FAMILY_VERSION    string = "1.0"
	DISTRIBUTION_NAME string = "sawtooth-intkey"
	DEFAULT_URL       string = "http://127.0.0.1:8008"
	// Verbs
	VERB_SET string = "set"
	VERB_INC string = "inc"
	VERB_DEC string = "dec"
	// APIs
	BATCH_SUBMIT_API string = "batches"
	BATCH_STATUS_API string = "batch_statuses"
	STATE_API        string = "state"
	// Content types
	CONTENT_TYPE_OCTET_STREAM string = "application/octet-stream"
	// Integer literals
	FAMILY_NAMESPACE_ADDRESS_LENGTH uint = 6
	FAMILY_VERB_ADDRESS_LENGTH      uint = 64
)

type IntkeyClient struct {
	signer *signing.Signer
}

func NewIntkeyClient(keyfile string) (IntkeyClient, error) {

	var privateKey signing.PrivateKey
	if keyfile != "" {
		// Read private key file
		privateKeyStr, err := ioutil.ReadFile(keyfile)
		if err != nil {
			return IntkeyClient{},
				errors.New(fmt.Sprintf("Failed to read private key: %v", err))
		}
		// Get private key object
		privateKey = signing.NewSecp256k1PrivateKey(privateKeyStr)
	} else {
		privateKey = signing.NewSecp256k1Context().NewRandomPrivateKey()
	}
	cryptoFactory := signing.NewCryptoFactory(signing.NewSecp256k1Context())
	signer := cryptoFactory.NewSigner(privateKey)
	return IntkeyClient{signer}, nil
}

func (intkeyClient IntkeyClient) Set(
	name string, value uint, wait uint, validatorURL string, signerURL string) (string, error) {
	return intkeyClient.sendTransaction(VERB_SET, name, value, wait, validatorURL, signerURL)
}

func (intkeyClient IntkeyClient) Inc(
	name string, value uint, wait uint, validatorURL string, signerURL string) (string, error) {
	return intkeyClient.sendTransaction(VERB_INC, name, value, wait, validatorURL, signerURL)
}

func (intkeyClient IntkeyClient) Dec(
	name string, value uint, wait uint, validatorURL string, signerURL string) (string, error) {
	return intkeyClient.sendTransaction(VERB_DEC, name, value, wait, validatorURL, signerURL)
}

func (intkeyClient IntkeyClient) List(validatorURL string) ([]map[interface{}]interface{}, error) {

	// API to call
	apiSuffix := fmt.Sprintf("%s?address=%s",
		STATE_API, intkeyClient.getPrefix())
	response, err := intkeyClient.sendRequest(validatorURL, apiSuffix, []byte{}, "", "")
	if err != nil {
		return []map[interface{}]interface{}{}, err
	}

	var toReturn []map[interface{}]interface{}
	responseMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return []map[interface{}]interface{}{},
			errors.New(fmt.Sprintf("Error reading response: %v", err))
	}
	encodedEntries := responseMap["data"].([]interface{})
	for _, entry := range encodedEntries {
		entryData, ok := entry.(map[interface{}]interface{})
		if !ok {
			return []map[interface{}]interface{}{},
				errors.New("Error reading entry data")
		}
		stringData, ok := entryData["data"].(string)
		if !ok {
			return []map[interface{}]interface{}{},
				errors.New("Error reading string data")
		}
		decodedBytes, err := base64.StdEncoding.DecodeString(stringData)
		if err != nil {
			return []map[interface{}]interface{}{},
				errors.New(fmt.Sprint("Error decoding: %v", err))
		}
		foundMap := make(map[interface{}]interface{})
		err = cbor.Loads(decodedBytes, &foundMap)
		if err != nil {
			return []map[interface{}]interface{}{},
				errors.New(fmt.Sprint("Error binary decoding: %v", err))
		}
		toReturn = append(toReturn, foundMap)
	}
	return toReturn, nil
}

func (intkeyClient IntkeyClient) Show(name string, validatorURL string) (string, error) {
	apiSuffix := fmt.Sprintf("%s/%s", STATE_API, intkeyClient.getAddress(name))
	response, err := intkeyClient.sendRequest(validatorURL, apiSuffix, []byte{}, "", name)
	if err != nil {
		return "", err
	}
	responseMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return "", errors.New(fmt.Sprint("Error reading response: %v", err))
	}
	data, ok := responseMap["data"].(string)
	if !ok {
		return "", errors.New("Error reading as string")
	}
	responseData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", errors.New(fmt.Sprint("Error decoding response: %v", err))
	}
	responseFinal := make(map[interface{}]interface{})
	err = cbor.Loads(responseData, &responseFinal)
	if err != nil {
		return "", errors.New(fmt.Sprint("Error binary decoding: %v", err))
	}
	return fmt.Sprintf("%v", responseFinal[name]), nil
}

func (intkeyClient IntkeyClient) getStatus(
	batchId string, wait uint, validatorURL string) (string, error) {

	// API to call
	apiSuffix := fmt.Sprintf("%s?id=%s&wait=%d",
		BATCH_STATUS_API, batchId, wait)
	response, err := intkeyClient.sendRequest(validatorURL, apiSuffix, []byte{}, "", "")
	if err != nil {
		return "", err
	}

	responseMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading response: %v", err))
	}
	entry :=
		responseMap["data"].([]interface{})[0].(map[interface{}]interface{})
	return fmt.Sprint(entry["status"]), nil
}

func (intkeyClient IntkeyClient) sendRequest(
	url string,
	apiSuffix string,
	data []byte,
	contentType string,
	name string) (string, error) {

	// Construct URL
	//var url string
	if strings.HasPrefix(url, "http://") {
		url = fmt.Sprintf("%s/%s", url, apiSuffix)
	} else {
		url = fmt.Sprintf("http://%s/%s", url, apiSuffix)
	}

	// Send request to validator URL
	var response *http.Response
	var err error
	if len(data) > 0 {
		response, err = http.Post(url, contentType, bytes.NewBuffer(data))
	} else {
		response, err = http.Get(url)
	}
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Failed to connect to REST API: %v", err))
	}
	if response.StatusCode == 404 {
		return "", errors.New(fmt.Sprintf("No such key: %s", name))
	} else if response.StatusCode >= 400 {
		return "", errors.New(
			fmt.Sprintf("Error %d: %s", response.StatusCode, response.Status))
	}
	defer response.Body.Close()
	reponseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading response: %v", err))
	}
	return string(reponseBody), nil
}

func (intkeyClient IntkeyClient) sendTransaction(
	verb string, name string, value uint, wait uint, validatorURL string, signerURL string) (string, error) {

	// construct the payload information in CBOR format
	payloadData := make(map[string]interface{})
	payloadData["Verb"] = verb
	payloadData["Name"] = name
	payloadData["Value"] = value
	payload, err := cbor.Dumps(payloadData)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to construct CBOR: %v", err))
	}

	// construct the address
	address := intkeyClient.getAddress(name)

	// Construct TransactionHeader
	rawTransactionHeader := transaction_pb2.TransactionHeader{
		SignerPublicKey:  intkeyClient.signer.GetPublicKey().AsHex(),
		FamilyName:       FAMILY_NAME,
		FamilyVersion:    FAMILY_VERSION,
		Dependencies:     []string{}, // empty dependency list
		Nonce:            strconv.Itoa(rand.Int()),
		BatcherPublicKey: intkeyClient.signer.GetPublicKey().AsHex(),
		Inputs:           []string{address},
		Outputs:          []string{address},
		PayloadSha512:    Sha512HashValue(string(payload)),
	}
	transactionHeader, err := proto.Marshal(&rawTransactionHeader)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to serialize transaction header: %v", err))
	}

	// Signature of TransactionHeader
	transactionHeaderSignature := hex.EncodeToString(
		intkeyClient.signer.Sign(transactionHeader))

	// Construct Transaction
	transaction := pb.Transaction{
		Header:          transactionHeader,
		HeaderSignature: transactionHeaderSignature,
		Payload:         []byte(payload),
	}

	// Get BatchList
	rawBatchList, err := intkeyClient.createBatchList(
		[]*pb.Transaction{&transaction}, signerURL)

	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to construct batch list: %v", err))
	}

	batchId := rawBatchList.Batches[0].HeaderSignature
	batchList, err := proto.Marshal(&rawBatchList)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to serialize batch list: %v", err))
	}

	if wait > 0 {
		waitTime := uint(0)
		startTime := time.Now()
		response, err := intkeyClient.sendRequest(
			validatorURL, BATCH_SUBMIT_API, batchList, CONTENT_TYPE_OCTET_STREAM, name)
		if err != nil {
			return "", err
		}
		for waitTime < wait {
			status, err := intkeyClient.getStatus(batchId, wait-waitTime, validatorURL)
			if err != nil {
				return "", err
			}
			waitTime = uint(time.Now().Sub(startTime))
			if status != "PENDING" {
				return response, nil
			}
		}
		return response, nil
	}

	return intkeyClient.sendRequest(
		validatorURL, BATCH_SUBMIT_API, batchList, CONTENT_TYPE_OCTET_STREAM, name)
}

func (intkeyClient IntkeyClient) getPrefix() string {
	return Sha512HashValue(FAMILY_NAME)[:FAMILY_NAMESPACE_ADDRESS_LENGTH]
}

func (intkeyClient IntkeyClient) getAddress(name string) string {
	prefix := intkeyClient.getPrefix()
	nameAddress := Sha512HashValue(name)[FAMILY_VERB_ADDRESS_LENGTH:]
	return prefix + nameAddress
}

func (intkeyClient IntkeyClient) createBatchList(
	transactions []*pb.Transaction, signerURL string) (pb.BatchList, error) {

	// Get list of TransactionHeader signatures
	transactionSignatures := []string{}
	for _, transaction := range transactions {
		transactionSignatures =
			append(transactionSignatures, transaction.HeaderSignature)
	}

	kc := keychain.NewKeyChain("./resources/keys/1/")
	pubKey, err := kc.LoadPublicKey("TBLS256_5_3")

	if err != nil {
		return pb.BatchList{}, err
	}

	pubKeyBytes, _ := pubKey.MarshalBinary()

	transactionList := pb.TransactionList{}
	transactionList.Transactions = transactions
	toSignBytes, _ := proto.Marshal(&transactionList)

	sig, scheme, err := performGroupSignature(toSignBytes, signerURL)

	if err != nil {
		return pb.BatchList{}, err
	}

	groupEnvelop := pb.GroupEnvelop{}
	groupEnvelop.PublicKey = pubKeyBytes
	groupEnvelop.Scheme = scheme
	groupEnvelop.Signature = sig
	groupEnvelopBytes, err := proto.Marshal(&groupEnvelop)

	// Construct BatchHeader
	rawBatchHeader := pb.BatchHeader{
		SignerPublicKey: intkeyClient.signer.GetPublicKey().AsHex(),
		TransactionIds:  transactionSignatures,
		GroupEnvelop:    groupEnvelopBytes,
	}
	batchHeader, err := proto.Marshal(&rawBatchHeader)
	if err != nil {
		return pb.BatchList{}, errors.New(
			fmt.Sprintf("Unable to serialize batch header: %v", err))
	}

	// Signature of BatchHeader
	batchHeaderSignature := hex.EncodeToString(
		intkeyClient.signer.Sign(batchHeader))

	// Construct Batch
	batch := pb.Batch{
		Header:          batchHeader,
		HeaderSignature: batchHeaderSignature,
		Transactions:    transactions,
		Trace:           true,
	}

	// Construct BatchList
	return pb.BatchList{
		Batches: []*pb.Batch{&batch},
	}, nil
}

func performGroupSignature(content []byte, signerURL string) ([]byte, string, error) {

	uuid := fmt.Sprint(uuid.NewV4())
	msg := pb.ClientSignMessage{
		UUID:                 fmt.Sprint(uuid),
		Content:              content,
		SmartContractAddress: "intkey",
	}

	b, err := proto.Marshal(&msg)

	reader := bytes.NewReader(b)

	//TODO change
	resp, err := http.Post(fmt.Sprintf("http://%v/sign", signerURL), "application/protobuf", reader)

	if err != nil {
		return nil, "", err
	}

	//fmt.Println(resp)
	body, _ := ioutil.ReadAll(resp.Body)

	respProto := pb.ClientSignResponse{}
	proto.Unmarshal(body, &respProto)

	return respProto.Signature, respProto.Scheme, err
}

func Sha512HashValue(value string) string {
	hashHandler := sha512.New()
	hashHandler.Write([]byte(value))
	return strings.ToLower(hex.EncodeToString(hashHandler.Sum(nil)))
}
