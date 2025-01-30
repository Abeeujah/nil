package types

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSszBlock(t *testing.T) {
	t.Parallel()

	block := Block{
		BlockData: BlockData{
			Id:                 1,
			PrevBlock:          common.Hash{0x01},
			SmartContractsRoot: common.Hash{0x02},
		},
	}

	encoded, err := block.MarshalSSZ()
	require.NoError(t, err)

	block2 := Block{}
	err = block2.UnmarshalSSZ(encoded)
	require.NoError(t, err)

	require.Equal(t, block2.Id, block.Id)
	require.Equal(t, block2.PrevBlock, block.PrevBlock)
	require.Equal(t, block2.SmartContractsRoot, block.SmartContractsRoot)

	h, err := common.PoseidonSSZ(&block2)
	require.NoError(t, err)

	h2, err := hex.DecodeString("21682bef122af58bd745341d3ece26a7c78a636d459382d5cd1a9da0f7ddc34e")
	require.NoError(t, err)

	require.Equal(t, common.BytesToHash(h2), common.BytesToHash(h[:]))
}

func TestSszTransaction(t *testing.T) {
	t.Parallel()

	transaction := Transaction{
		TransactionDigest: TransactionDigest{
			To:    Address{},
			Data:  Code{0x00000001},
			Seqno: 567,
		},
		From:  Address{},
		Value: NewValueFromUint64(1234),
	}

	encoded, err := transaction.MarshalSSZ()
	require.NoError(t, err)

	transaction2 := Transaction{}
	err = transaction2.UnmarshalSSZ(encoded)
	require.NoError(t, err)

	require.Equal(t, transaction2.From, transaction.From)
	require.Equal(t, transaction2.To, transaction.To)
	require.Equal(t, transaction2.Value, transaction.Value)
	require.Equal(t, transaction2.Data, transaction.Data)
	require.Equal(t, transaction2.Seqno, transaction.Seqno)
	require.True(t, bytes.Equal(transaction2.Signature, transaction.Signature))

	h, err := common.PoseidonSSZ(&transaction2)
	require.NoError(t, err)

	h2 := common.HexToHash("2d3efc5c6f1d6ade476e0ed2641cde7e863434f7eb2429d59cc1844a0144ff38")
	require.Equal(t, h2, h)
}

func TestSszSmc(t *testing.T) {
	t.Parallel()

	smc := SmartContract{
		Address:     HexToAddress("1d9bc16f1a559"),
		Initialised: true,
		Balance:     NewValueFromUint64(1234),
		StorageRoot: common.Hash{0x01},
		CodeHash:    common.Hash{0x02},
		Seqno:       567,
	}

	encoded, err := smc.MarshalSSZ()
	require.NoError(t, err)

	smc2 := SmartContract{}
	err = smc2.UnmarshalSSZ(encoded)
	require.NoError(t, err)

	require.Equal(t, smc.Address, smc2.Address)
	require.Equal(t, smc.Initialised, smc2.Initialised)
	require.Equal(t, smc.Balance, smc2.Balance)
	require.Equal(t, smc.StorageRoot, smc2.StorageRoot)
	require.Equal(t, smc.CodeHash, smc2.CodeHash)
	require.Equal(t, smc.Seqno, smc2.Seqno)

	h, err := common.PoseidonSSZ(&smc2)
	require.NoError(t, err)

	h2 := common.HexToHash("0x1101c3ea5b9afdc740f1f37823a9bb5209276e1581237909ddba9836ad010237")
	require.Equal(t, h2, common.BytesToHash(h[:]))
}

type testcase struct {
	val string
	exp string
}

var sszUint256TestCases = []testcase{
	{"", "0000000000000000000000000000000000000000000000000000000000000000"},
	{"01", "0100000000000000000000000000000000000000000000000000000000000000"},
	{"02", "0200000000000000000000000000000000000000000000000000000000000000"},
	{"04", "0400000000000000000000000000000000000000000000000000000000000000"},
	{"08", "0800000000000000000000000000000000000000000000000000000000000000"},
	{"10", "1000000000000000000000000000000000000000000000000000000000000000"},
	{"20", "2000000000000000000000000000000000000000000000000000000000000000"},
	{"40", "4000000000000000000000000000000000000000000000000000000000000000"},
	{"80", "8000000000000000000000000000000000000000000000000000000000000000"},
	{"0100", "0001000000000000000000000000000000000000000000000000000000000000"},
	{"0200", "0002000000000000000000000000000000000000000000000000000000000000"},
	{"0400", "0004000000000000000000000000000000000000000000000000000000000000"},
	{"0800", "0008000000000000000000000000000000000000000000000000000000000000"},
	{"1000", "0010000000000000000000000000000000000000000000000000000000000000"},
	{"2000", "0020000000000000000000000000000000000000000000000000000000000000"},
	{"4000", "0040000000000000000000000000000000000000000000000000000000000000"},
	{"8000", "0080000000000000000000000000000000000000000000000000000000000000"},
	{"010000", "0000010000000000000000000000000000000000000000000000000000000000"},
	{"020000", "0000020000000000000000000000000000000000000000000000000000000000"},
	{"040000", "0000040000000000000000000000000000000000000000000000000000000000"},
	{"080000", "0000080000000000000000000000000000000000000000000000000000000000"},
	{"100000", "0000100000000000000000000000000000000000000000000000000000000000"},
	{"200000", "0000200000000000000000000000000000000000000000000000000000000000"},
	{"400000", "0000400000000000000000000000000000000000000000000000000000000000"},
	{"800000", "0000800000000000000000000000000000000000000000000000000000000000"},
	{"01000000", "0000000100000000000000000000000000000000000000000000000000000000"},
	{"02000000", "0000000200000000000000000000000000000000000000000000000000000000"},
	{"04000000", "0000000400000000000000000000000000000000000000000000000000000000"},
	{"08000000", "0000000800000000000000000000000000000000000000000000000000000000"},
	{"10000000", "0000001000000000000000000000000000000000000000000000000000000000"},
	{"20000000", "0000002000000000000000000000000000000000000000000000000000000000"},
	{"40000000", "0000004000000000000000000000000000000000000000000000000000000000"},
	{"80000000", "0000008000000000000000000000000000000000000000000000000000000000"},
	{"0100000000", "0000000001000000000000000000000000000000000000000000000000000000"},
	{"0200000000", "0000000002000000000000000000000000000000000000000000000000000000"},
	{"0400000000", "0000000004000000000000000000000000000000000000000000000000000000"},
	{"0800000000", "0000000008000000000000000000000000000000000000000000000000000000"},
	{"1000000000", "0000000010000000000000000000000000000000000000000000000000000000"},
	{"2000000000", "0000000020000000000000000000000000000000000000000000000000000000"},
	{"4000000000", "0000000040000000000000000000000000000000000000000000000000000000"},
	{"8000000000", "0000000080000000000000000000000000000000000000000000000000000000"},
	{"010000000000", "0000000000010000000000000000000000000000000000000000000000000000"},
	{"020000000000", "0000000000020000000000000000000000000000000000000000000000000000"},
	{"040000000000", "0000000000040000000000000000000000000000000000000000000000000000"},
	{"080000000000", "0000000000080000000000000000000000000000000000000000000000000000"},
	{"100000000000", "0000000000100000000000000000000000000000000000000000000000000000"},
	{"200000000000", "0000000000200000000000000000000000000000000000000000000000000000"},
	{"400000000000", "0000000000400000000000000000000000000000000000000000000000000000"},
	{"800000000000", "0000000000800000000000000000000000000000000000000000000000000000"},
	{"01000000000000", "0000000000000100000000000000000000000000000000000000000000000000"},
	{"02000000000000", "0000000000000200000000000000000000000000000000000000000000000000"},
	{"04000000000000", "0000000000000400000000000000000000000000000000000000000000000000"},
	{"08000000000000", "0000000000000800000000000000000000000000000000000000000000000000"},
	{"10000000000000", "0000000000001000000000000000000000000000000000000000000000000000"},
	{"20000000000000", "0000000000002000000000000000000000000000000000000000000000000000"},
	{"40000000000000", "0000000000004000000000000000000000000000000000000000000000000000"},
	{"80000000000000", "0000000000008000000000000000000000000000000000000000000000000000"},
	{"0100000000000000", "0000000000000001000000000000000000000000000000000000000000000000"},
	{"0200000000000000", "0000000000000002000000000000000000000000000000000000000000000000"},
	{"0400000000000000", "0000000000000004000000000000000000000000000000000000000000000000"},
	{"0800000000000000", "0000000000000008000000000000000000000000000000000000000000000000"},
	{"1000000000000000", "0000000000000010000000000000000000000000000000000000000000000000"},
	{"2000000000000000", "0000000000000020000000000000000000000000000000000000000000000000"},
	{"4000000000000000", "0000000000000040000000000000000000000000000000000000000000000000"},
	{"8000000000000000", "0000000000000080000000000000000000000000000000000000000000000000"},
	{"010000000000000000", "0000000000000000010000000000000000000000000000000000000000000000"},
	{"020000000000000000", "0000000000000000020000000000000000000000000000000000000000000000"},
	{"040000000000000000", "0000000000000000040000000000000000000000000000000000000000000000"},
	{"080000000000000000", "0000000000000000080000000000000000000000000000000000000000000000"},
	{"100000000000000000", "0000000000000000100000000000000000000000000000000000000000000000"},
	{"200000000000000000", "0000000000000000200000000000000000000000000000000000000000000000"},
	{"400000000000000000", "0000000000000000400000000000000000000000000000000000000000000000"},
	{"800000000000000000", "0000000000000000800000000000000000000000000000000000000000000000"},
	{"01000000000000000000", "0000000000000000000100000000000000000000000000000000000000000000"},
	{"02000000000000000000", "0000000000000000000200000000000000000000000000000000000000000000"},
	{"04000000000000000000", "0000000000000000000400000000000000000000000000000000000000000000"},
	{"08000000000000000000", "0000000000000000000800000000000000000000000000000000000000000000"},
	{"10000000000000000000", "0000000000000000001000000000000000000000000000000000000000000000"},
	{"20000000000000000000", "0000000000000000002000000000000000000000000000000000000000000000"},
	{"40000000000000000000", "0000000000000000004000000000000000000000000000000000000000000000"},
	{"80000000000000000000", "0000000000000000008000000000000000000000000000000000000000000000"},
	{"0100000000000000000000", "0000000000000000000001000000000000000000000000000000000000000000"},
	{"0200000000000000000000", "0000000000000000000002000000000000000000000000000000000000000000"},
	{"0400000000000000000000", "0000000000000000000004000000000000000000000000000000000000000000"},
	{"0800000000000000000000", "0000000000000000000008000000000000000000000000000000000000000000"},
	{"1000000000000000000000", "0000000000000000000010000000000000000000000000000000000000000000"},
	{"2000000000000000000000", "0000000000000000000020000000000000000000000000000000000000000000"},
	{"4000000000000000000000", "0000000000000000000040000000000000000000000000000000000000000000"},
	{"8000000000000000000000", "0000000000000000000080000000000000000000000000000000000000000000"},
	{"010000000000000000000000", "0000000000000000000000010000000000000000000000000000000000000000"},
	{"020000000000000000000000", "0000000000000000000000020000000000000000000000000000000000000000"},
	{"040000000000000000000000", "0000000000000000000000040000000000000000000000000000000000000000"},
	{"080000000000000000000000", "0000000000000000000000080000000000000000000000000000000000000000"},
	{"100000000000000000000000", "0000000000000000000000100000000000000000000000000000000000000000"},
	{"200000000000000000000000", "0000000000000000000000200000000000000000000000000000000000000000"},
	{"400000000000000000000000", "0000000000000000000000400000000000000000000000000000000000000000"},
	{"800000000000000000000000", "0000000000000000000000800000000000000000000000000000000000000000"},
	{"01000000000000000000000000", "0000000000000000000000000100000000000000000000000000000000000000"},
	{"02000000000000000000000000", "0000000000000000000000000200000000000000000000000000000000000000"},
	{"04000000000000000000000000", "0000000000000000000000000400000000000000000000000000000000000000"},
	{"08000000000000000000000000", "0000000000000000000000000800000000000000000000000000000000000000"},
	{"10000000000000000000000000", "0000000000000000000000001000000000000000000000000000000000000000"},
	{"20000000000000000000000000", "0000000000000000000000002000000000000000000000000000000000000000"},
	{"40000000000000000000000000", "0000000000000000000000004000000000000000000000000000000000000000"},
	{"80000000000000000000000000", "0000000000000000000000008000000000000000000000000000000000000000"},
	{"0100000000000000000000000000", "0000000000000000000000000001000000000000000000000000000000000000"},
	{"0200000000000000000000000000", "0000000000000000000000000002000000000000000000000000000000000000"},
	{"0400000000000000000000000000", "0000000000000000000000000004000000000000000000000000000000000000"},
	{"0800000000000000000000000000", "0000000000000000000000000008000000000000000000000000000000000000"},
	{"1000000000000000000000000000", "0000000000000000000000000010000000000000000000000000000000000000"},
	{"2000000000000000000000000000", "0000000000000000000000000020000000000000000000000000000000000000"},
	{"4000000000000000000000000000", "0000000000000000000000000040000000000000000000000000000000000000"},
	{"8000000000000000000000000000", "0000000000000000000000000080000000000000000000000000000000000000"},
	{"010000000000000000000000000000", "0000000000000000000000000000010000000000000000000000000000000000"},
	{"020000000000000000000000000000", "0000000000000000000000000000020000000000000000000000000000000000"},
	{"040000000000000000000000000000", "0000000000000000000000000000040000000000000000000000000000000000"},
	{"080000000000000000000000000000", "0000000000000000000000000000080000000000000000000000000000000000"},
	{"100000000000000000000000000000", "0000000000000000000000000000100000000000000000000000000000000000"},
	{"200000000000000000000000000000", "0000000000000000000000000000200000000000000000000000000000000000"},
	{"400000000000000000000000000000", "0000000000000000000000000000400000000000000000000000000000000000"},
	{"800000000000000000000000000000", "0000000000000000000000000000800000000000000000000000000000000000"},
	{"01000000000000000000000000000000", "0000000000000000000000000000000100000000000000000000000000000000"},
	{"02000000000000000000000000000000", "0000000000000000000000000000000200000000000000000000000000000000"},
	{"04000000000000000000000000000000", "0000000000000000000000000000000400000000000000000000000000000000"},
	{"08000000000000000000000000000000", "0000000000000000000000000000000800000000000000000000000000000000"},
	{"10000000000000000000000000000000", "0000000000000000000000000000001000000000000000000000000000000000"},
	{"20000000000000000000000000000000", "0000000000000000000000000000002000000000000000000000000000000000"},
	{"40000000000000000000000000000000", "0000000000000000000000000000004000000000000000000000000000000000"},
	{"80000000000000000000000000000000", "0000000000000000000000000000008000000000000000000000000000000000"},
	{"0100000000000000000000000000000000", "0000000000000000000000000000000001000000000000000000000000000000"},
	{"0200000000000000000000000000000000", "0000000000000000000000000000000002000000000000000000000000000000"},
	{"0400000000000000000000000000000000", "0000000000000000000000000000000004000000000000000000000000000000"},
	{"0800000000000000000000000000000000", "0000000000000000000000000000000008000000000000000000000000000000"},
	{"1000000000000000000000000000000000", "0000000000000000000000000000000010000000000000000000000000000000"},
	{"2000000000000000000000000000000000", "0000000000000000000000000000000020000000000000000000000000000000"},
	{"4000000000000000000000000000000000", "0000000000000000000000000000000040000000000000000000000000000000"},
	{"8000000000000000000000000000000000", "0000000000000000000000000000000080000000000000000000000000000000"},
	{"010000000000000000000000000000000000", "0000000000000000000000000000000000010000000000000000000000000000"},
	{"020000000000000000000000000000000000", "0000000000000000000000000000000000020000000000000000000000000000"},
	{"040000000000000000000000000000000000", "0000000000000000000000000000000000040000000000000000000000000000"},
	{"080000000000000000000000000000000000", "0000000000000000000000000000000000080000000000000000000000000000"},
	{"100000000000000000000000000000000000", "0000000000000000000000000000000000100000000000000000000000000000"},
	{"200000000000000000000000000000000000", "0000000000000000000000000000000000200000000000000000000000000000"},
	{"400000000000000000000000000000000000", "0000000000000000000000000000000000400000000000000000000000000000"},
	{"800000000000000000000000000000000000", "0000000000000000000000000000000000800000000000000000000000000000"},
	{"01000000000000000000000000000000000000", "0000000000000000000000000000000000000100000000000000000000000000"},
	{"02000000000000000000000000000000000000", "0000000000000000000000000000000000000200000000000000000000000000"},
	{"04000000000000000000000000000000000000", "0000000000000000000000000000000000000400000000000000000000000000"},
	{"08000000000000000000000000000000000000", "0000000000000000000000000000000000000800000000000000000000000000"},
	{"10000000000000000000000000000000000000", "0000000000000000000000000000000000001000000000000000000000000000"},
	{"20000000000000000000000000000000000000", "0000000000000000000000000000000000002000000000000000000000000000"},
	{"40000000000000000000000000000000000000", "0000000000000000000000000000000000004000000000000000000000000000"},
	{"80000000000000000000000000000000000000", "0000000000000000000000000000000000008000000000000000000000000000"},
	{"0100000000000000000000000000000000000000", "0000000000000000000000000000000000000001000000000000000000000000"},
	{"0200000000000000000000000000000000000000", "0000000000000000000000000000000000000002000000000000000000000000"},
	{"0400000000000000000000000000000000000000", "0000000000000000000000000000000000000004000000000000000000000000"},
	{"0800000000000000000000000000000000000000", "0000000000000000000000000000000000000008000000000000000000000000"},
	{"1000000000000000000000000000000000000000", "0000000000000000000000000000000000000010000000000000000000000000"},
	{"2000000000000000000000000000000000000000", "0000000000000000000000000000000000000020000000000000000000000000"},
	{"4000000000000000000000000000000000000000", "0000000000000000000000000000000000000040000000000000000000000000"},
	{"8000000000000000000000000000000000000000", "0000000000000000000000000000000000000080000000000000000000000000"},
	{"010000000000000000000000000000000000000000", "0000000000000000000000000000000000000000010000000000000000000000"},
	{"020000000000000000000000000000000000000000", "0000000000000000000000000000000000000000020000000000000000000000"},
	{"040000000000000000000000000000000000000000", "0000000000000000000000000000000000000000040000000000000000000000"},
	{"080000000000000000000000000000000000000000", "0000000000000000000000000000000000000000080000000000000000000000"},
	{"100000000000000000000000000000000000000000", "0000000000000000000000000000000000000000100000000000000000000000"},
	{"200000000000000000000000000000000000000000", "0000000000000000000000000000000000000000200000000000000000000000"},
	{"400000000000000000000000000000000000000000", "0000000000000000000000000000000000000000400000000000000000000000"},
	{"800000000000000000000000000000000000000000", "0000000000000000000000000000000000000000800000000000000000000000"},
	{"01000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000100000000000000000000"},
	{"02000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000200000000000000000000"},
	{"04000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000400000000000000000000"},
	{"08000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000800000000000000000000"},
	{"10000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000001000000000000000000000"},
	{"20000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000002000000000000000000000"},
	{"40000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000004000000000000000000000"},
	{"80000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000008000000000000000000000"},
	{"0100000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000001000000000000000000"},
	{"0200000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000002000000000000000000"},
	{"0400000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000004000000000000000000"},
	{"0800000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000008000000000000000000"},
	{"1000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000010000000000000000000"},
	{"2000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000020000000000000000000"},
	{"4000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000040000000000000000000"},
	{"8000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000080000000000000000000"},
	{"010000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000010000000000000000"},
	{"020000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000020000000000000000"},
	{"040000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000040000000000000000"},
	{"080000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000080000000000000000"},
	{"100000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000100000000000000000"},
	{"200000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000200000000000000000"},
	{"400000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000400000000000000000"},
	{"800000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000800000000000000000"},
	{"01000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000100000000000000"},
	{"02000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000200000000000000"},
	{"04000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000400000000000000"},
	{"08000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000800000000000000"},
	{"10000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000001000000000000000"},
	{"20000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000002000000000000000"},
	{"40000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000004000000000000000"},
	{"80000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000008000000000000000"},
	{"0100000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000001000000000000"},
	{"0200000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000002000000000000"},
	{"0400000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000004000000000000"},
	{"0800000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000008000000000000"},
	{"1000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000010000000000000"},
	{"2000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000020000000000000"},
	{"4000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000040000000000000"},
	{"8000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000080000000000000"},
	{"010000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000010000000000"},
	{"020000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000020000000000"},
	{"040000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000040000000000"},
	{"080000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000080000000000"},
	{"100000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000100000000000"},
	{"200000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000200000000000"},
	{"400000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000400000000000"},
	{"800000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000800000000000"},
	{"01000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000100000000"},
	{"02000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000200000000"},
	{"04000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000400000000"},
	{"08000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000800000000"},
	{"10000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000001000000000"},
	{"20000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000002000000000"},
	{"40000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000004000000000"},
	{"80000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000008000000000"},
	{"0100000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000001000000"},
	{"0200000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000002000000"},
	{"0400000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000004000000"},
	{"0800000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000008000000"},
	{"1000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000010000000"},
	{"2000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000020000000"},
	{"4000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000040000000"},
	{"8000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000080000000"},
	{"010000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000010000"},
	{"020000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000020000"},
	{"040000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000040000"},
	{"080000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000080000"},
	{"100000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000100000"},
	{"200000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000200000"},
	{"400000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000400000"},
	{"800000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000800000"},
	{"01000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000100"},
	{"02000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000200"},
	{"04000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000400"},
	{"08000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000800"},
	{"10000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000001000"},
	{"20000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000002000"},
	{"40000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000004000"},
	{"80000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000008000"},
	{"0100000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001"},
	{"0200000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000002"},
	{"0400000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000004"},
	{"0800000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000008"},
	{"1000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000010"},
	{"2000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000020"},
	{"4000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000040"},
	{"8000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000080"},
}

func TestSszUint256(t *testing.T) {
	t.Parallel()

	for _, tt := range sszUint256TestCases {
		z := NewUint256FromBytes(hexutil.FromHex(tt.val))
		assert.Equal(t, 32, z.SizeSSZ())

		b, err := z.MarshalSSZ()
		require.NoError(t, err)
		assert.Equal(t, hexutil.FromHex(tt.exp), b)

		z2 := &Uint256{}
		require.NoError(t, z2.UnmarshalSSZ(b))
		assert.Equal(t, z, z2)

		r, err := z.HashTreeRoot()
		require.NoError(t, err)
		assert.Equal(t, hexutil.FromHex(tt.exp), r[:])
	}
}
