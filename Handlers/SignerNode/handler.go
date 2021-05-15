package SignerNode

import (
	"bytes"
	"fmt"
	"github.com/jffp113/SignerNode_Thesis/client"
	"github.com/jffp113/Thesis_Client/Client"
	"github.com/jffp113/Thesis_Client/Handlers/Algorand"
	"github.com/jffp113/Thesis_Client/conf"
	"github.com/jffp113/go-algorand-sdk/client/algod"
	"github.com/jffp113/go-algorand-sdk/encoding/msgpack"
	"github.com/jffp113/go-algorand-sdk/types"
	"time"
)

type signerNode struct {
	signerNodesURL []string

	//Permissionless Settings
	IsPermissionless       bool
	IsOneTimeKey           bool
	IsGroupRandomGenerated bool
	N                      int
	T                      int
	Scheme                 string

	key *client.Key

	//If incorporating with algo
	clients []algod.Client
	sendToAlgorand bool
}

func (h *signerNode) InitHandler(config conf.Configuration) {
	h.signerNodesURL = config.Conf.SignerNodes
	h.IsPermissionless = config.Conf.IsPermissionless
	h.IsOneTimeKey = config.Conf.IsOneTimeKey
	h.IsGroupRandomGenerated = config.Conf.IsGroupRandomGenerated
	h.N = config.Conf.N
	h.T = config.Conf.T
	h.Scheme = config.Conf.Scheme
	h.sendToAlgorand = config.Conf.SendSignatureToAlgorand


	if h.IsPermissionless && !h.IsOneTimeKey{
		key,err := h.InstallKey()
		h.key = key
		if err != nil {
			panic(err)
		}
	}


	//Parse algorand Validators
	if h.sendToAlgorand {
		for _,v := range config.Conf.ValidatorNodes {
			algodClient, err := algod.MakeClient(fmt.Sprintf("http://%s",v), config.Conf.Token)
			if err != nil {
				panic(err)
			}

			h.clients = append(h.clients,algodClient)
		}
	}
}


func (h *signerNode) InstallKey() (*client.Key,error) {
	cli := client.NewPermissionlessClient()
	var membership []string

	if h.IsGroupRandomGenerated {
		membership = client.GetSubsetMembership(h.signerNodesURL,h.N)
	} else{
		membership = client.GetNNearestNodes(h.signerNodesURL,h.N)
	}

	gen := getKeyGen(h.Scheme)
	pub, priv := gen.Gen(h.N,h.T)

	key := client.Key{
		T:               h.T,
		N:               h.N,
		Scheme:          h.Scheme,
		ValidUntil:      time.Now().Add(60*time.Minute),
		IsOneTimeKey:    h.IsOneTimeKey,
		PubKey:          pub,
		PrivKeys:        priv,
		GroupMembership: membership,
	}

	err := cli.InstallShare(&key)
	return &key,err
}

func (h *signerNode) DoRequest() Client.RequestStatus {
	var stats Client.RequestStatus

	if !h.IsPermissionless{
		err := performPermissionedTransaction(h.signerNodesURL,&stats)
		if err == nil {
			stats.Success = true
		}
	}else{
		err := performPermissionlessTransaction(h,&stats)
		if err == nil {
			stats.Success = true
		}
	}

	return stats
}

func NewHandler() Client.Handler {
	return &signerNode{}
}

func performPermissionedTransaction(urls []string,stats *Client.RequestStatus) error {
	c ,err := client.NewPermissionedClient(client.SetSignerNodeAddresses(urls...))

	if err != nil {
		return err
	}

	stats.StartTime = time.Now()
	_, err = c.SendSignRequest([]byte("Hello"),"intkey")
	stats.EndTime = time.Now()
	return err
}

//performPermissionlessTransaction does to many things.
//If Algorand property is activated it generates a algorand transaction to be sent
//Next if one time key is activated it generates a key on the fly installs
//Only next asks for a group signature over the data, depending on the flags it can happen
//over a algorand transaction or over the bytes "hello".
func performPermissionlessTransaction(h *signerNode,stats *Client.RequestStatus) error {
	c  := client.NewPermissionlessClient()
	var err error
	var tx types.Transaction
	var algoCli algod.Client
	contentToSign := []byte("Hello")
	key := h.key

	stats.StartTime = time.Now()

	//fmt.Println(h.sendToAlgorand)
	//If we are sending to algorand, should change bytes to sign
	if h.sendToAlgorand {
		_,algoCli = Algorand.ChooseOne(h.clients)
		tx,err = Algorand.CreateAlgoTransaction(algoCli)

		if err != nil {
			return err
		}

		encodedTx := msgpack.Encode(tx)
		msgParts := [][]byte{[]byte("TX"), encodedTx}
		contentToSign = bytes.Join(msgParts, nil)
	}

	if h.IsOneTimeKey {
		key, err = h.InstallKey()
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	sig, err := c.SendSignRequest(contentToSign,"intkey",key)

	if err != nil {
		//fmt.Println(err)
		return nil
	}

	//Creating signed msg to send to algorand
	if h.sendToAlgorand {
		pub, _ :=  key.PubKey.MarshalBinary()
		id, b, err := Algorand.SignTransaction(tx,types.GroupEnvelop{
			PublicKey: pub,
			Signature: sig.Signature,
			Scheme:    sig.Scheme,
		})

		if err != nil {
			return err
		}
		err = Algorand.SendTransaction(algoCli,b,id)
	}

	stats.EndTime = time.Now()
	return err
}