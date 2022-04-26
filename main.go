package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Block structure
type Block struct {
	Pos       int
	Data      JacketCheckout
	Timestamp string
	Hash      string
	PrevHash  string
}

type JacketCheckout struct {
	JacketID     string `json:"jacket_id"`
	User         string `json:"user"`
	CheckoutDate string `json:"checkout_date"`
	IsGenesis    bool   `json:"is_genesis"`
}

type Jacket struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Author          string `json:"author"`
	ElaborationDate string `json:"Elaboration_date"`
	Serial          string `json:"serial"`
}

func (b *Block) generateHash() {
	// get string val of the Data

	bytes, _ := json.Marshal(b.Data)
	// concatenate the dataset
	data := string(b.Pos) + b.Timestamp + string(bytes) + b.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

func CreateBlock(prevBlock *Block, checkoutItem JacketCheckout) *Block {

	block := &Block{}
	block.Pos = prevBlock.Pos + 1
	block.Timestamp = time.Now().String()
	block.Data = checkoutItem
	block.PrevHash = prevBlock.Hash
	block.generateHash()

	return block
}

// Blockchain structure .. slices of pointer to blocks
type Blockchain struct {
	blocks []*Block
}

// basic variable pointer to the blockchain structure
var BlockChain *Blockchain

func (bc *Blockchain) AddBlock(data JacketCheckout) {

	prevBlock := bc.blocks[len(bc.blocks)-1]

	block := CreateBlock(prevBlock, data)

	if validBlock(block, prevBlock) {
		bc.blocks = append(bc.blocks, block)
	}
}

func GenesisBlock() *Block {
	return CreateBlock(&Block{}, JacketCheckout{IsGenesis: true})
}

func NewBlockchain() *Blockchain {

	return &Blockchain{[]*Block{GenesisBlock()}}
}

func validBlock(block, prevBlock *Block) bool {

	if prevBlock.Hash != block.PrevHash {
		return false
	}

	if !block.validateHash(block.Hash) {
		return false
	}

	if prevBlock.Pos+1 != block.Pos {
		return false
	}
	return true
}

func (b *Block) validateHash(hash string) bool {
	b.generateHash()
	//if b.Hash != hash {
	//	return false
	//}
	//return true
	return b.Hash == hash
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	jbytes, err := json.MarshalIndent(BlockChain.blocks, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	io.WriteString(w, string(jbytes))
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var checkoutItem JacketCheckout
	if err := json.NewDecoder(r.Body).Decode(&checkoutItem); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not write Block: %v", err)
		w.Write([]byte("could not write block"))
		return
	}

	BlockChain.AddBlock(checkoutItem)
	resp, err := json.MarshalIndent(checkoutItem, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal payload: %v", err)
		w.Write([]byte("could not write block"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// Creating a book
func newJacket(w http.ResponseWriter, r *http.Request) {
	var jacket Jacket
	if err := json.NewDecoder(r.Body).Decode(&jacket); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not create: %v", err)
		w.Write([]byte("could not create new Jacket"))
		return
	}

	h := md5.New()
	io.WriteString(h, jacket.Serial+jacket.ElaborationDate)
	jacket.ID = fmt.Sprintf("%x", h.Sum(nil))

	resp, err := json.MarshalIndent(jacket, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal payload: %v", err)
		w.Write([]byte("could not save jacket data"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func main() {

	BlockChain = NewBlockchain()

	r := mux.NewRouter()
	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newJacket).Methods("POST")

	go func() {

		for _, block := range BlockChain.blocks {
			fmt.Printf("Prev. hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data: %v\n", string(bytes))
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Println()
		}

	}()
	log.Println("Listening on port 3000")

	log.Fatal(http.ListenAndServe(":3000", r))
}
