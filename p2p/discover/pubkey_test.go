package discover

import (
	"math/big"
	"net"
	"testing"

	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"
)

func findSmallCurvePoint(t *testing.T) (*big.Int, *big.Int) {
	t.Helper()

	params := crypto.S256().Params()
	two256 := new(big.Int).Lsh(big.NewInt(1), 256)
	limit := new(big.Int).Sub(two256, params.P) // need x < 2^256 - P so x+P still fits in 32 bytes

	for i := int64(1); ; i++ {
		x := big.NewInt(i)
		if x.Cmp(limit) >= 0 {
			t.Fatal("failed to find small curve point")
		}
		rhs := new(big.Int).Mul(x, x)
		rhs.Mul(rhs, x)
		rhs.Add(rhs, params.B)
		rhs.Mod(rhs, params.P)

		y := new(big.Int).ModSqrt(rhs, params.P)
		if y != nil {
			return x, y
		}
	}
}

func encodeBadPubkey(x, y *big.Int) (out v4wire.Pubkey) {
	ethmath.ReadBits(x, out[:32])
	ethmath.ReadBits(y, out[32:])
	return out
}

func TestMalformedNeighborPubkeyRejected(t *testing.T) {
	x, y := findSmallCurvePoint(t)
	badX := new(big.Int).Add(x, crypto.S256().Params().P)

	rn := v4wire.Node{
		IP:  net.ParseIP("127.0.0.1").To4(),
		UDP: 30303,
		TCP: 30303,
		ID:  encodeBadPubkey(badX, y),
	}

	if _, err := v4wire.DecodePubkey(crypto.S256(), rn.ID); err == nil {
		t.Fatal("expected non-canonical pubkey to be rejected")
	}

	var udp UDPv4
	sender := &net.UDPAddr{IP: net.ParseIP("127.0.0.1").To4(), Port: 30303}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	if _, err := udp.nodeFromRPC(sender, rn); err == nil {
		t.Fatal("expected nodeFromRPC to fail on bad pubkey")
	}
}
