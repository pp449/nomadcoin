package blockchain

import (
	"errors"
	"time"

	"github.com/nomadcoders/nomadcoin/utils"
)

const (
	minerReward int = 50
)

type mempool struct {
	Txs []*Tx
}

var Mempool *mempool = &mempool{}

type Tx struct {
	ID        string   `json:"id"`
	Timestamp int      `json:"timestamp"`
	TxIns     []*TxIn  `json:"txIns"`
	TxOuts    []*TxOut `json:"txOut"`
}

func (t *Tx) getId() {
	t.ID = utils.Hash(t)
}

type TxIn struct {
	TxID  string `json:"txId"`
	Index int    `json:"index"`
	Owner string `json:"owner"`
}

type TxOut struct {
	Owner  string `json:"owner"`
	Amount int    `json:"amount"`
}

type UTxOut struct {
	TxID   string
	Index  int
	Amount int
}

func isOnMempool(uTxOut *UTxOut) bool {
	for _, tx := range Mempool.Txs {
		for _, input := range tx.TxIns {
			if input.TxID == uTxOut.TxID && input.Index == uTxOut.Index {
				return true
			}
		}
	}
	return false
}

func makeCoinbaseTx(address string) *Tx { // Coinbase 기반으로 채굴보상 transaction을 만듦
	txIns := []*TxIn{
		{"", -1, "COINBASE"},
	}
	txOuts := []*TxOut{
		{address, minerReward},
	}
	tx := Tx{
		ID:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.getId()
	return &tx
}

func makeTx(from, to string, amount int) (*Tx, error) { // 액수가 된다면 송금 transaction을 만듦
	if BalanceByAddress(from, Blockchain()) < amount {
		return nil, errors.New("not enough 돈")
	}
	var txOuts []*TxOut
	var txIns []*TxIn
	total := 0
	uTxOuts := UTxOutsByAddress(from, Blockchain())
	for _, uTxOut := range uTxOuts {
		if total >= amount {
			break
		}
		txIn := &TxIn{uTxOut.TxID, uTxOut.Index, from}
		txIns = append(txIns, txIn)
		total += uTxOut.Amount
	}
	if change := total - amount; change != 0 {
		changeTxOut := &TxOut{from, change}
		txOuts = append(txOuts, changeTxOut)
	}
	txOut := &TxOut{to, amount}
	txOuts = append(txOuts, txOut)
	tx := &Tx{
		ID:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.getId()
	return tx, nil
}

func (m *mempool) AddTx(to string, amount int) error { // 송금 transaction이 성공하면 mempool에 저장
	tx, err := makeTx("nico", to, amount)
	if err != nil {
		return err
	}
	m.Txs = append(m.Txs, tx)
	return nil
}

func (m *mempool) TxToConfirm() []*Tx { // mempool에 있는 transaction을 confirm하는듯. 근데 makeCoinbaseTx는 잘 모르겠네 ㅎㅎㅎ
	coinbase := makeCoinbaseTx("nico")
	txs := m.Txs
	txs = append(txs, coinbase)
	m.Txs = nil
	return txs
}
