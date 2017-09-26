package gripcrypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"hash"
	"io"
	"log"
	"math/big"
	"os"
)

//MODELEN is the number of bytes in the version
const MODELEN int = 8

//ECDSAMODEP521 Indicates signature is generated with ECDSA
const ECDSAMODEP521 uint64 = 1

//CURRENTMODE Indicates the signature method used for all new signatures
const CURRENTMODE uint64 = ECDSAMODEP521

//DigestInf digests a struct with the interface
type DigestInf interface {
	Digest() []byte
	GetDig() []byte
}

//SignInf signs a struct that implements Digest as well
type SignInf interface {
	DigestInf
	GetNodeID() []byte
	SetNodeID(id []byte)
	GetSig() []byte
	SetSig(b []byte)
}

//HashBool writes a bool to a hash
func HashBool(h hash.Hash, b bool) {
	if b {
		h.Write([]byte{1})
	} else {
		h.Write([]byte{0})
	}
}

//HashUint8 writes a uint8 to a hash
func HashUint8(h hash.Hash, v uint8) {
	h.Write([]byte{v})
}

//HashUint32 writes a uint32 to a hash
func HashUint32(h hash.Hash, v uint32) {
	a := make([]byte, 4)
	binary.BigEndian.PutUint32(a, v)
	h.Write(a)
}

//HashUint64 writes a uint32 to a hash
func HashUint64(h hash.Hash, v uint64) {
	a := make([]byte, 8)
	binary.BigEndian.PutUint64(a, v)
	h.Write(a)
}

//HashString writes a string to a hash
func HashString(h hash.Hash, s string) {
	h.Write([]byte(s))
}

//HashBytes writes a byte slice to a hash
func HashBytes(h hash.Hash, b []byte) {
	if b != nil {
		h.Write(b)
	}
}

//HashFile hashes the contents of a file.
func HashFile(h hash.Hash, f string) {
	buf := make([]byte, 32*1024)
	file, err := os.Open(f)
	if err != nil {
		log.Print(err)
		h.Write([]byte{0})
	} else {
		defer file.Close()
		for {
			n, err := file.Read(buf)
			if n > 0 {
				h.Write(buf[:n])
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("read %d bytes: %v", n, err)
				break
			}
		}
	}
}

//ReadInt read int from slice
func ReadInt(s []byte) (int, int) {
	bl := int(binary.BigEndian.Uint64(s[0:8]))
	return bl, 8
}

//ReadBytes read a slice from a slice
func ReadBytes(s []byte) ([]byte, int) {
	l, il := ReadInt(s)
	return s[il:(l + il)], l + il
}

//ReadBigInt read big int from a slice
func ReadBigInt(s []byte) (*big.Int, int) {
	bs, bl := ReadBytes(s)
	bo := big.NewInt(0)
	bo = bo.SetBytes(bs)
	return bo, bl
}

//PrepareBigInt prepare to read big int from slice
func PrepareBigInt(b *big.Int) ([]byte, int) {
	bo := b.Bytes()
	return bo, len(bo) + 8
}

//WriteBigInt write big int slice to a slice
func WriteBigInt(s []byte, b []byte) int {
	bl := len(b)
	binary.BigEndian.PutUint64(s, uint64(bl))
	copy(s[8:8+bl], b)
	return 8 + bl
}

//GenerateECDSAKeyPair generates ecdsa key pair
func GenerateECDSAKeyPair() (prv, pub []byte, e error) {
	prvk, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	xb, xblen := PrepareBigInt(prvk.PublicKey.X)
	yb, yblen := PrepareBigInt(prvk.PublicKey.Y)
	db, dblen := PrepareBigInt(prvk.D)

	privatekey := make([]byte, MODELEN+dblen+xblen+yblen)
	binary.BigEndian.PutUint64(privatekey[0:MODELEN], ECDSAMODEP521)
	nv := MODELEN
	nv += WriteBigInt(privatekey[nv:], xb)
	nv += WriteBigInt(privatekey[nv:], yb)
	nv += WriteBigInt(privatekey[nv:], db)

	pubkey := make([]byte, MODELEN+xblen+yblen)
	binary.BigEndian.PutUint64(pubkey[0:MODELEN], ECDSAMODEP521)
	nv = MODELEN
	nv += WriteBigInt(pubkey[nv:], xb)
	nv += WriteBigInt(pubkey[nv:], yb)

	return privatekey, pubkey, nil
}

//Verify verifies a signature
func Verify(g SignInf, pk []byte) bool {
	idx := 0
	ek := binary.BigEndian.Uint64(pk[idx:MODELEN])
	idx += MODELEN
	sig := g.GetSig()
	ek2 := binary.BigEndian.Uint64(sig[0:MODELEN])
	if ek == ECDSAMODEP521 && ek2 == ECDSAMODEP521 {
		//This is NOT on safecurve list.  FIXME!  Implement something better
		var x, y *big.Int
		var nx int
		x, nx = ReadBigInt(pk[idx:])
		idx += nx
		y, nx = ReadBigInt(pk[idx:])
		cv := elliptic.P521()
		p := ecdsa.PublicKey{Curve: cv, X: x, Y: y}
		//Get sig big.Ints
		idx = MODELEN
		var r, s *big.Int
		r, nx = ReadBigInt(sig[idx:])
		idx += nx
		s, nx = ReadBigInt(sig[idx:])
		return ecdsa.Verify(&p, g.Digest(), r, s)
	}
	return false
}

//Sign signs a sign interface
func Sign(g SignInf, pk []byte) error {
	idx := 0
	ek := binary.BigEndian.Uint64(pk[idx:MODELEN])
	idx += MODELEN
	if ek == ECDSAMODEP521 {
		//This is NOT on safecurve list.  FIXME!  Implement something better
		var x, y, d *big.Int
		var nx int
		x, nx = ReadBigInt(pk[idx:])
		idx += nx
		y, nx = ReadBigInt(pk[idx:])
		idx += nx
		d, nx = ReadBigInt(pk[idx:])
		cv := elliptic.P521()
		p := ecdsa.PublicKey{Curve: cv, X: x, Y: y}
		pv := ecdsa.PrivateKey{PublicKey: p, D: d}
		r, s, err := ecdsa.Sign(rand.Reader, &pv, g.Digest())
		if err != nil {
			return err
		}
		rb, rblen := PrepareBigInt(r)
		sb, sblen := PrepareBigInt(s)
		if rblen < 0 || sblen < 0 {
			return errors.New("Signature byte array less than 0")
		}
		totallen := MODELEN + rblen + sblen
		ob := make([]byte, totallen)
		idx := 0
		binary.BigEndian.PutUint64(ob[0:MODELEN], ECDSAMODEP521)
		idx += MODELEN
		idx += WriteBigInt(ob[idx:idx+rblen], rb)
		WriteBigInt(ob[idx:idx+sblen], sb)
		g.SetSig(ob)
		return nil
	}
	return errors.New("Unknown key type")
}
