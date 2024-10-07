package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Struct untuk merepresentasikan item menu
type MenuItem struct {
	Name     string
	Price    float64
	Quantity int
}

// Interface untuk mendefinisikan metode umum pesanan
type OrderProcessor interface {
	ProcessOrder(order Order) error
}

// Struct untuk pesanan
type Order struct {
	ID         int
	ItemName   string
	Quantity   int
	Price      float64
	TotalPrice float64
}

// Interface kosong untuk menangani berbagai tipe data
type Any interface{}

// Mutex untuk menghindari race condition saat mengakses menu
var menuMutex sync.Mutex

// Menu slice untuk menyimpan item menu
var menu = []MenuItem{
	{Name: "Nasi Goreng", Price: 15000, Quantity: 10},
	{Name: "Mie Ayam", Price: 12000, Quantity: 8},
	{Name: "Sate Ayam", Price: 20000, Quantity: 5},
	{Name: "Es Teh", Price: 5000, Quantity: 20},
}

// WaitGroup untuk menunggu semua goroutine selesai
var wg sync.WaitGroup

// Channel untuk komunikasi antara goroutine dan proses utama
var orderChan = make(chan Order, 10)

// Timeout duration untuk pemrosesan pesanan
const timeoutDuration = 5 * time.Second

// Variable untuk menyimpan total semua pesanan
var totalAllOrders float64
var totalMutex sync.Mutex

func main() {
	// Defer untuk mencetak pesan "Program selesai" selalu
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from error:", r)
		}
		fmt.Println("Program selesai")
	}()

	// Mulai pemrosesan pesanan
	go processOrders()

	reader := bufio.NewReader(os.Stdin)
	orderID := 1

	for {
		fmt.Println("\n===== Sistem Manajemen Pesanan Restoran =====")
		fmt.Println("1. Tampilkan Menu")
		fmt.Println("2. Buat Pesanan")
		fmt.Println("3. Tampilkan Total Semua Pesanan")
		fmt.Println("4. Keluar")
		fmt.Print("Pilih opsi: ")

		input, _ := reader.ReadString('\n')
		option := strings.TrimSpace(input)

		switch option {
		case "1":
			displayMenu()
		case "2":
			order := createOrder(reader, orderID)
			if order != nil {
				wg.Add(1)
				orderChan <- *order
				orderID++
			}
		case "3":
			displayTotalAllOrders()
		case "4":
			close(orderChan)
			wg.Wait()
			return
		default:
			fmt.Println("Opsi tidak valid. Silakan coba lagi.")
		}
	}
}

// Fungsi untuk menampilkan menu
func displayMenu() {
	menuMutex.Lock()
	defer menuMutex.Unlock()

	if len(menu) == 0 {
		fmt.Println("Menu kosong.")
		return
	}

	fmt.Println("\n===== Menu =====")
	for _, item := range menu {
		fmt.Printf("Nama: %s | Harga: %.2f | Stok: %d\n", item.Name, item.Price, item.Quantity)
	}
}

// Fungsi untuk membuat pesanan
func createOrder(reader *bufio.Reader, orderID int) *Order {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi kesalahan:", r)
		}
	}()

	fmt.Print("Masukkan nama item yang dipesan: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	menuMutex.Lock()
	defer menuMutex.Unlock()

	var selectedItem *MenuItem
	for i, item := range menu {
		if strings.EqualFold(item.Name, name) {
			selectedItem = &menu[i]
			break
		}
	}

	if selectedItem == nil {
		fmt.Println("Item tidak ditemukan.")
		return nil
	}

	fmt.Print("Masukkan jumlah: ")
	quantityInput, _ := reader.ReadString('\n')
	quantityInput = strings.TrimSpace(quantityInput)

	// Validasi jumlah menggunakan regexp
	matched, err := regexp.MatchString(`^\d+$`, quantityInput)
	if err != nil {
		panic(err)
	}
	if !matched {
		panic("Jumlah harus berupa angka")
	}

	quantity, err := strconv.Atoi(quantityInput)
	if err != nil || quantity <= 0 {
		panic("Jumlah harus berupa angka positif")
	}

	if quantity > selectedItem.Quantity {
		fmt.Println("Jumlah melebihi stok yang tersedia.")
		return nil
	}

	selectedItem.Quantity -= quantity

	order := &Order{
		ID:         orderID,
		ItemName:   selectedItem.Name,
		Quantity:   quantity,
		Price:      selectedItem.Price,
		TotalPrice: float64(quantity) * selectedItem.Price,
	}

	return order
}

// Fungsi untuk memproses pesanan menggunakan goroutine dan channel
func processOrders() {
	for order := range orderChan {
		go func(ord Order) {
			defer wg.Done()
			processOrder(ord)
		}(order)
	}
}

// Implementasi dari interface OrderProcessor
func (op *OrderProcessorImpl) ProcessOrder(order Order) error {
	// Simulasi pemrosesan pesanan
	time.Sleep(2 * time.Second)
	fmt.Printf("Pesanan ID %d: %s x%d telah diproses.\n", order.ID, order.ItemName, order.Quantity)
	return nil
}

// Struct implementasi OrderProcessor
type OrderProcessorImpl struct{}

// Fungsi untuk memproses satu pesanan
func processOrder(order Order) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in goroutine:", r)
		} 
	}()

	op := &OrderProcessorImpl{}
	err := op.ProcessOrder(order)
	if err != nil {
		panic(err)
	}

	// Encode detail pesanan menggunakan base64
	orderDetails := fmt.Sprintf("ID:%d,Item:%s,Quantity:%d,TotalPrice:%.2f", order.ID, order.ItemName, order.Quantity, order.TotalPrice)
	encoded := base64.StdEncoding.EncodeToString([]byte(orderDetails))
	fmt.Printf("Detail Pesanan Terencode: %s\n", encoded)

	// Update total semua pesanan
	totalMutex.Lock()
	totalAllOrders += order.TotalPrice
	totalMutex.Unlock()
}

// Fungsi untuk menampilkan total semua pesanan
func displayTotalAllOrders() {
	totalMutex.Lock()
	defer totalMutex.Unlock()

	fmt.Printf("Total Semua Pesanan: %.2f\n", totalAllOrders)
}
