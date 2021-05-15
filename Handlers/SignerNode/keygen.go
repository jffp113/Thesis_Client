package SignerNode

import (
	"github.com/jffp113/CryptoProviderSDK/crypto"
	"github.com/jffp113/CryptoProviderSDK/example/handlers/tbls"
	"github.com/jffp113/CryptoProviderSDK/example/handlers/trsa"
)

//Used to get the correct key generator
func getKeyGen(scheme string) crypto.KeyShareGenerator {
	switch scheme {
	case "TBLS256Optimistic":
		fallthrough
	case "TBLS256Pessimistic":
		fallthrough
	case "TBLS256":
		return tbls.NewTBLS256KeyGenerator()
	case "TRSA1024Optimistic":
		fallthrough
	case "TRSA1024Pessimistic":
		fallthrough
	case "TRSA1024":
		return trsa.NewTRSAKeyGenerator(1024)
	case "TRSA2048Optimistic":
		fallthrough
	case "TRSA2048Pessimistic":
		fallthrough
	case "TRSA2048":
		return trsa.NewTRSAKeyGenerator(2048)
	default:
		return nil
	}
}
