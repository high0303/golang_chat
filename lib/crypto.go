package lib

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	//"encoding/base64"
	"encoding/hex"
	"log"
	"math/big"
)

type ECDSASignature struct {
	Hash [64]byte
	R    *big.Int
	S    *big.Int
}

type CTRCipher struct {
}

func GenerateSHA512Hash(in_str string) string {
	hasher := sha512.New()
	hasher.Write([]byte(in_str))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash
}

func GenerateECDSAKeys() *ecdsa.PrivateKey {
	log.Print("Generating ECDSA keypair.")
	ecdsaPrivateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		log.Fatal("Failed to generate ECDSA key.")
	}

	log.Print("Generated ECDSA keypair.")
	log.Printf("Public key: %s", ecdsaPrivateKey.PublicKey)

	return ecdsaPrivateKey
}

func CreateSignature(priv *ecdsa.PrivateKey, data []byte) (signature *ECDSASignature, err error) {
	//hash := sha512.Sum512(data)
	hash := data
	// TODO: I think one shouldn't use new?!
	signature = new(ECDSASignature)
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
	//signature.Hash = hash
	signature.R = r
	signature.S = s

	return signature, err
}

func VerifySignature(pub *ecdsa.PublicKey, sig *ECDSASignature) bool {
	return ecdsa.Verify(pub, sig.Hash[:], sig.R, sig.S)
}

func CreateCTRCipher() {
	// Generate the IV
	iv := make([]byte, 16)
	_, err := rand.Read(iv)
	if err != nil {
		log.Fatal("Failed to create IV for CTR cipher.")
	}

	log.Print(iv)

	/*
		iv := {1,2,3,4,5,... }
		block, err := aes.NewCipher(key)
		aes := cipher.NewCBCEncrypter(block, iv)
		aes.CryptBlocks(out, in)
	*/
}

func GetRandomHash() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic("Failed to get 32 random bytes!")
	}

	return GenerateSHA512Hash(hex.EncodeToString(b))
}

