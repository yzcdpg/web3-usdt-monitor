package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// User 存储用户信息
type User struct {
	Address string `json:"address"`
	Balance int64  `json:"balance"`
}

// Monitor 监控系统
type Monitor struct {
	users       map[string]*User // 地址到用户的映射
	mutex       sync.RWMutex
	httpClient  *http.Client
	apiKey      string // TronGrid API密钥
	usdtAddress string // USDT合约地址
	lastBlock   int64
}

// NewMonitor 创建新的监控实例
func NewMonitor(apiKey, usdtAddress string) *Monitor {
	return &Monitor{
		users:       make(map[string]*User),
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		apiKey:      apiKey,
		usdtAddress: usdtAddress,
	}
}

// AddUser 添加新用户地址
func (m *Monitor) AddUser(address string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.users[address]; !exists {
		m.users[address] = &User{
			Address: address,
			Balance: 0,
		}
		log.Printf("Added user: %s", address)
	}
}

// HexToBase58 将十六进制地址转换为 TRON Base58Check 地址
func HexToBase58(hexAddr string) (string, error) {
	if len(hexAddr) >= 2 && hexAddr[:2] == "41" {
		hexAddr = hexAddr[2:]
	}

	addrBytes, err := hex.DecodeString(hexAddr)
	if err != nil {
		return "", fmt.Errorf("invalid hex address: %v", err)
	}

	addrWithPrefix := append([]byte{0x41}, addrBytes...)
	hash1 := sha256.Sum256(addrWithPrefix)
	hash2 := sha256.Sum256(hash1[:])
	checksum := hash2[:4]

	fullBytes := append(addrWithPrefix, checksum...)
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	result := make([]byte, 0, 34)
	num := new(big.Int).SetBytes(fullBytes)
	zero := big.NewInt(0)
	base := big.NewInt(58)

	for num.Cmp(zero) > 0 {
		mod := new(big.Int)
		num.DivMod(num, base, mod)
		result = append([]byte{alphabet[mod.Int64()]}, result...)
	}

	for _, b := range fullBytes {
		if b != 0 {
			break
		}
		result = append([]byte{'1'}, result...)
	}

	return string(result), nil
}

// Block 区块结构
type Block struct {
	BlockHeader struct {
		RawData struct {
			Number int64 `json:"number"`
		} `json:"raw_data"`
	} `json:"block_header"`
	Transactions []struct {
		TxID    string `json:"txID"`
		RawData struct {
			Contract []struct {
				Type      string `json:"type"`
				Parameter struct {
					Value struct {
						OwnerAddress    string `json:"owner_address"`
						ContractAddress string `json:"contract_address"`
						Data            string `json:"data"`
					} `json:"value"`
				} `json:"parameter"`
			} `json:"contract"`
		} `json:"raw_data"`
	} `json:"transactions"`
}

// ProcessTransaction 解析 TriggerSmartContract 交易的合约数据
func (m *Monitor) ProcessTransaction(blockNum int64, tx struct {
	TxID    string `json:"txID"`
	RawData struct {
		Contract []struct {
			Type      string `json:"type"`
			Parameter struct {
				Value struct {
					OwnerAddress    string `json:"owner_address"`
					ContractAddress string `json:"contract_address"`
					Data            string `json:"data"`
				} `json:"value"`
			} `json:"parameter"`
		} `json:"contract"`
	} `json:"raw_data"`
}) {
	for _, contract := range tx.RawData.Contract {
		if contract.Type != "TriggerSmartContract" {
			log.Println("Not trigger contract")
			continue
		}

		data := contract.Parameter.Value.Data
		if len(data) < 8 || strings.ToLower(data[:8]) != "a9059cbb" {
			log.Println("Not Transfer")
			continue // 不是 transfer(address,uint256) 函数
		}

		// 解析 transfer(address,uint256) 参数
		if len(data) < 136 { // 8 (selector) + 64 (to) + 64 (value)
			log.Printf("Invalid data length for tx %s: %s", tx.TxID, data)
			continue
		}

		// 提取 from 地址
		fromAddress := contract.Parameter.Value.OwnerAddress
		fromBase58, err := HexToBase58(fromAddress)
		if err != nil {
			log.Printf("Failed to decode from address for tx %s: %v", tx.TxID, err)
			continue
		}

		// 提取 to 地址（32字节，填充后取后20字节）
		toAddress := data[16:80][24:] // 跳过前12字节填充
		toBase58, err := HexToBase58("41" + toAddress)
		if err != nil {
			log.Printf("Failed to decode to address for tx %s: %v", tx.TxID, err)
			continue
		}

		// 提取金额（uint256，32字节）
		amountHex := data[72:136]
		amountBytes, err := hex.DecodeString(amountHex)
		if err != nil {
			log.Printf("Failed to decode amount for tx %s: %v", tx.TxID, err)
			continue
		}
		amount := new(big.Int).SetBytes(amountBytes)
		fmt.Printf("block: %v tx: %v   %v ==> %v   %v\n", blockNum, tx.TxID, fromBase58, toBase58, amount)
		m.mutex.RLock()
		user, exists := m.users[toBase58]
		m.mutex.RUnlock()

		if exists {
			m.mutex.Lock()
			user.Balance += amount.Int64()
			m.mutex.Unlock()
			log.Printf("Updated balance for %s: +%d USDT from %s (tx: %s)", toBase58, amount.Int64(), fromBase58, tx.TxID)
		}
	}
}

// Start 启动监控
func (m *Monitor) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			url := fmt.Sprintf("https://api.trongrid.io/wallet/getblockbynum?num=%d", m.lastBlock)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Printf("Error creating block request: %v", err)
				continue
			}
			req.Header.Set("TRON-PRO-API-KEY", m.apiKey)
			req.Header.Set("Accept", "application/json")

			resp, err := m.httpClient.Do(req)
			if err != nil {
				log.Printf("Error fetching block %d: %v", m.lastBlock, err)
				continue
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Printf("Error reading block response: %v", err)
				continue
			}

			var block Block
			if err := json.Unmarshal(body, &block); err != nil {
				log.Printf("Error parsing block %d: %v", m.lastBlock, err)
				continue
			}
			log.Println(block.BlockHeader.RawData.Number)
			for _, tx := range block.Transactions {
				if tx.RawData.Contract[0].Type == "TriggerSmartContract" && tx.RawData.Contract[0].Parameter.Value.ContractAddress == "41a614f803b6fd780986a42c78ec9c7f77e6ded13c" {
					//log.Println(tx.TxID)
					go m.ProcessTransaction(m.lastBlock, tx)
				}
			}
			m.mutex.Lock()
			m.lastBlock++
			m.mutex.Unlock()
		}
	}
}

// REST API handlers
func (m *Monitor) handleAddUser(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	m.AddUser(user.Address)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User added successfully"})
}

func (m *Monitor) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	m.mutex.RLock()
	user, exists := m.users[address]
	m.mutex.RUnlock()

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (m *Monitor) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	users := make([]*User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}

	json.NewEncoder(w).Encode(users)
}

// loadLastBlock 从文件加载 lastBlock
func (m *Monitor) loadLastBlock() error {
	data, err := os.ReadFile("last_block.json")
	if os.IsNotExist(err) {
		log.Println("No last_block.json found, starting from latest block")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read last_block.json: %v", err)
	}

	var state struct {
		LastBlock int64 `json:"last_block"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse last_block.json: %v", err)
	}

	m.mutex.Lock()
	m.lastBlock = state.LastBlock
	m.mutex.Unlock()
	log.Printf("Loaded lastBlock: %d", m.lastBlock)
	return nil
}

// saveLastBlock 保存 lastBlock 到文件
func (m *Monitor) saveLastBlock() error {
	m.mutex.RLock()
	state := struct {
		LastBlock int64 `json:"last_block"`
	}{LastBlock: m.lastBlock}
	m.mutex.RUnlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lastBlock: %v", err)
	}

	if err := os.WriteFile("last_block.json", data, 0644); err != nil {
		return fmt.Errorf("failed to write last_block.json: %v", err)
	}
	log.Printf("Saved lastBlock: %d", state.LastBlock)
	return nil
}

func main() {
	// 配置
	apiKey := ""                                        // 从 TronGrid 获取
	usdtAddress := "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" // USDT 合约地址

	// 初始化监控
	monitor := NewMonitor(apiKey, usdtAddress)

	err := monitor.loadLastBlock()
	if err != nil {
		log.Fatalln(err)
	}

	// 启动监控
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go monitor.Start(ctx)

	// 设置 REST API
	router := mux.NewRouter()
	router.HandleFunc("/api/users", monitor.handleAddUser).Methods("POST")
	router.HandleFunc("/api/users/{address}", monitor.handleGetUser).Methods("GET")
	router.HandleFunc("/api/users", monitor.handleGetAllUsers).Methods("GET")

	// 捕获退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, saving lastBlock...")
		if err := monitor.saveLastBlock(); err != nil {
			log.Printf("Error saving lastBlock: %v", err)
		}
		cancel()
		os.Exit(0)
	}()

	// 启动服务器
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
