package lib

import "testing"

func TestGenerateECDSAKeys(t *testing.T) {
	privkey1 := GenerateECDSAKeys()
	privkey2 := GenerateECDSAKeys()

	if(privkey1.D == privkey2.D) {
		t.Error("Two generated private keys are the same!")
	}
}

func TestSignature(t *testing.T) {
	privkey1 := GenerateECDSAKeys()
	privkey2 := GenerateECDSAKeys()

	sig1, err := CreateSignature(privkey1, []byte("foobar"))
	if(err != nil) {
		t.Error("Failed to create signature.")
	}
	sig2, err := CreateSignature(privkey2, []byte("foobar"))
	if(err != nil) {
		t.Error("Failed to create signature.")
	}
	sig3, err := CreateSignature(privkey1, []byte("foobar2"))
	if(err != nil) {
		t.Error("Failed to create signature.")
	}

	if(!VerifySignature(&privkey1.PublicKey, sig1)) {
		t.Error("Failed to verify signature.")
	}
	if(!VerifySignature(&privkey2.PublicKey, sig2)) {
		t.Error("Failed to verify signature.")
	}
	if(VerifySignature(&privkey2.PublicKey, sig1)) {
		t.Error("Signature verification verified faulty signature.")
	}
	if(VerifySignature(&privkey1.PublicKey, sig2)) {
		t.Error("Signature verification verified faulty signature.")
	}

	if(sig1.hash != sig2.hash) {
		t.Error("Hash function is broken!")
	}
	if(sig3.hash == sig1.hash) {
		t.Error("Hash function is broken!")
	}
}

func TestCTRCipher(t *testing.T) {
	CreateCTRCipher()
}
