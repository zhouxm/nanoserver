package algoutil

import (
	"testing"
)

func TestGenRSAKey(t *testing.T) {
	ts := []string{"hello world", "miss right"}
	priv, pub, err := GenRSAKey()
	if err != nil {
		t.Fatal(err)
	}

	for _, tstr := range ts {
		c, err := RSAEncrypt([]byte(tstr), pub)
		if err != nil {
			t.Fatal(err)
		}
		d, err := RSADecrypt(c, priv)
		if string(d) != tstr {
			t.Fail()
		}
	}
}

func BenchmarkGenRSAKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenRSAKey()
	}
}


