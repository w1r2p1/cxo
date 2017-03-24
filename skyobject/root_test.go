package skyobject

import (
	"bytes"
	"testing"

	"github.com/skycoin/skycoin/src/cipher"
)

func TestRoot_Touch(t *testing.T) {
	//
}

func TestRoot_Inject(t *testing.T) {
	//
}

func Test_signature(t *testing.T) {
	pub, sec := cipher.GenerateKeyPair()
	b := []byte("hello")
	hash := cipher.SumSHA256(b)
	sig := cipher.SignHash(hash, sec)
	if err := cipher.VerifySignature(pub, sig, hash); err != nil {
		t.Error(err)
	}
}

func Test_encodeEqual(t *testing.T) {
	c := getCont()
	r := c.NewRoot(pubKey())
	if bytes.Compare(r.Encode(), r.Encode()) != 0 {
		t.Error("not equal")
	}
}

func TestRoot_Encode(t *testing.T) {
	pub, sec := cipher.GenerateKeyPair()
	// encode
	c1 := getCont()
	r1 := c1.NewRoot(pub)
	r1.Register("User", User{})
	r1.SaveSchema(Group{})
	r1.Sign(sec)
	p := r1.Encode()
	// decode
	c2 := getCont()
	if ok, err := c2.SetEncodedRoot(p, r1.Pub, r1.Sig); err != nil {
		t.Error(err)
	} else if !ok {
		t.Error("can't set encoded root")
	} else if len(c2.reg.reg) != len(c1.reg.reg) {
		t.Error("wrong registry")
	}
}

func TestRoot_SchemaByReference(t *testing.T) {
	//
}

func TestRoot_Save(t *testing.T) {
	//
}

func TestRoot_SaveArray(t *testing.T) {
	//
}

func TestRoot_SaveSchema(t *testing.T) {
	//
}

func TestRoot_Dynamic(t *testing.T) {
	//
}

func TestRoot_Register(t *testing.T) {
	//
}

func TestRoot_Values(t *testing.T) {
	//
}
