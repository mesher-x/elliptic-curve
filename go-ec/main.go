package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

func main() {
	//fst()
	//snd()
	trd()
}

func fst() {
	Pub := "ffff"
	by := hexutil.MustDecode("0x" + Pub)
	// kernel start
	curve := secp256k1.S256()
	x, y := curve.ScalarBaseMult(by)
	x2, y2 := curve.ScalarBaseMult(by)
	x3, y3 := curve.Add(x, y, x2, y2)
	b2 := curve.Marshal(x3, y3)
	// kernel end
	c := hexutil.Encode(b2)
	println("c:")
	println(c)
}

func snd() {
	Pub := "045769b6ab9a4719181badde312dafe40af6f1ee9676ad413d5d5f08989ea28f690fb1ce68a891f325303618f7c87bc5719570ff559c00be0503033eff68388768"
	s := "ffff11a4"

	PubB := hexutil.MustDecode("0x" + Pub)
	sB := hexutil.MustDecode("0x" + s)
	// kernel start
	curve := secp256k1.S256()
	pX, pY := curve.Unmarshal(PubB)
	sX, sY := curve.ScalarBaseMult(sB)
	x, y := Subtract(pX, pY, sX, sY)
	b2 := curve.Marshal(x, y)
	// kernel end
	c := hexutil.Encode(b2)
	println("c:")
	println(c)
}

type KernelResults struct {
	pubKey      [12]byte
	word1Offset int
}

// - [ ]  Загрузка файла в память cpp
// - [ ]  Загрузка файла в память gpu
// - [ ]  Хорошо бы видеть общую скорость переборов в секунду, но если реализация долгая, то позднее
// - [ ]  Поиск правильного word1Offset и word4Offset
func trd() {
	filename12b := "12b_1k.txt" // можно константой
	filename8b := "8b_1k.txt"   // можно константой

	// // Для Pub=045769b6ab9a4719181badde312dafe40af6f1ee9676ad413d5d5f08989ea28f690fb1ce68a891f325303618f7c87bc5719570ff559c00be0503033eff68388768
	// // Нужно найти 0xb9faffb5adab083e1d01d0a8 в файле 12b_1k.txt (1000 строка)
	// // FOUND: 3000110 990123 basePub: 0xb9faffb5adab083e1d01d0a8 inputPub: 045769b6ab9a4719181badde312dafe40af6f1ee9676ad413d5d5f08989ea28f690fb1ce68a891f325303618f7c87bc5719570ff559c00be0503033eff68388768
	// // для этого подходят такие параметры запуска:
	word1Start := 3000110
	word1End := 3000112
	word4Start := 990123
	word4End := 990125

	// // Для большой нагрузки:
	//word1Start := 0
	//word1End := 4000000
	//word4Start := 0
	//word4End := 1000000

	Pub := "045769b6ab9a4719181badde312dafe40af6f1ee9676ad413d5d5f08989ea28f690fb1ce68a891f325303618f7c87bc5719570ff559c00be0503033eff68388768"
	PubB := hexutil.MustDecode("0x" + Pub)

	arr12b := ReadFile12b(filename12b)
	arr8b := ReadFile8b(filename8b)

	// kernel global start
	curve := secp256k1.S256()
	pX, pY := curve.Unmarshal(PubB)
	baseWord1, _ := new(big.Int).SetString("1000000000000000000000000000000000000000000000000", 16) // this number need multiply by word1Offset

	// Константы, чтобы правильно вырезать pubKey и поискать внутри кернела
	// 0x045769b6ab9a4719181badde312dafe40af6f1ee9676ad413d5d5f08989ea28f690fb1ce68a891f325303618f7c87bc5719570ff559c00be0503033eff68388768
	//     5769b6ab9a4719181badde31  // 12bytes
	//    0 1 2 3 4 5 6 7 8 9101112
	//     5769b6ab9a471918          // 8bytes
	//    0 1 2 3 4 5 6 7 8
	PubLeft8b := 1
	PubRight8b := 1 + 8
	PubLeft12b := 1
	PubRight12b := 1 + 12
	// kernel global end

	// Количество word4Offset < word1Offset, поэтому кажется оптимальнее искать так: word4Offset -> kernel(word1Offset)
	for word4Offset := word4Start; word4Offset < word4End; word4Offset++ { // Чтобы загрузить видеокарту полностью, создаем много kernel
		kernelResults := make([]*KernelResults, word4End)

		// kernel start
		word4 := new(big.Int).SetUint64(uint64(word4Offset))

		for word1Offset := word1Start; word1Offset < word1End; word1Offset++ {
			// 1stword with three word zeroes
			word1 := new(big.Int).Mul(new(big.Int).SetInt64(int64(word1Offset)), baseWord1)
			offset := new(big.Int).Add(word1, word4)
			offsetX, offsetY := secp256k1.S256().ScalarBaseMult(offset.Bytes())

			pubX14word, pubY14word := Subtract(pX, pY, offsetX, offsetY)
			PubBytes := secp256k1.S256().Marshal(pubX14word, pubY14word)[PubLeft8b:PubRight8b] // skip 1st byte which always 04
			//p2 := hexutil.Encode(secp256k1.S256().Marshal(pubX14word, pubY14word)[PubLeft12b:PubRight12b])
			//println(p2)
			index := BinarySearch8b(*arr8b, PubBytes)
			if index == -1 {
				continue
			}

			p := secp256k1.S256().Marshal(pubX14word, pubY14word)[PubLeft12b:PubRight12b] // skip 1st byte which always 04

			// Сохраняем результат итерации кернела
			kernelResults[word4Offset] = &KernelResults{
				pubKey:      [12]byte{p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7], p[8], p[9], p[10], p[11]},
				word1Offset: word1Offset,
			}
		}
		// kernel end

		for word4Offset, kernelResult := range kernelResults {
			if kernelResult == nil {
				continue
			}

			index := BinarySearch12b(*arr12b, kernelResult.pubKey[:])
			if index == -1 {
				continue
			}
			println(index)
			// Может быть несколько FOUND
			println("FOUND:", kernelResult.word1Offset, word4Offset, "basePub:", hexutil.Encode(kernelResult.pubKey[:]), "inputPub:", Pub)
		}
	}
}

func Subtract(x1, y1, x2, y2 *big.Int) (x, y *big.Int) {
	// new Point(this.x, mod(-this.y));
	y2Temp := new(big.Int).Neg(y2)                        // -this.y
	y2Temp = y2Temp.Mod(y2Temp, crypto.S256().Params().P) // mod

	return secp256k1.S256().Add(x1, y1, x2, y2Temp)
}

func ReadFile12b(filepath string) *[][12]byte {
	var index uint64 = 0
	fmt.Println("Loading 12b into memory", filepath)

	//var maxSize = 3311521968
	//arr := make([][12]byte{}, maxSize) // Количество строк известно заранее
	arr := [][12]byte{}

	readFile, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Buffer([]byte{}, 10000000)
	fileScanner.Split(bufio.ScanLines)

	s := time.Now()
	s0 := time.Now()
	for fileScanner.Scan() {
		text := fileScanner.Text()
		if text == "" {
			continue
		}
		data := hexutil.MustDecode("0x" + text)

		if index%10000000 == 0 {
			fmt.Println(index, time.Since(s))
			s = time.Now()
		}

		// 0x045769b6ab9a4719181badde312dafe40af6f1ee9676ad413d5d5f08989ea28f690fb1ce68a891f325303618f7c87bc5719570ff559c00be0503033eff68388768
		//     5769b6ab9a4719181badde31  // text, 12bytes
		arr = append(arr, [12]byte{data[0], data[1], data[2], data[3], data[4], data[5], data[6], data[7], data[8], data[9], data[10], data[11]})
		index++

		if index >= 4294967296 {
			fmt.Println("next index will out of bounds 4294967296")
			break
		}
	}
	fmt.Println("Loaded in ", time.Since(s0), "index", index)

	if err := fileScanner.Err(); err != nil {
		log.Fatal(err)
	}

	ind := BinarySearch12b(arr, hexutil.MustDecode("0x"+"b9faffb5adab083e1d01d0a8"))
	fmt.Println("0000000438ada1501aa58612 index is", ind)

	return &arr
}

func ReadFile8b(filepath string) *[][8]byte {
	var index uint64 = 0
	fmt.Println("Loading 8b into gpu memory", filepath)

	//var maxSize = 3311521968
	//arr := make([][12]byte{}, maxSize) // Количество строк известно заранее
	arr := [][8]byte{}

	readFile, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Buffer([]byte{}, 10000000)
	fileScanner.Split(bufio.ScanLines)

	s := time.Now()
	s0 := time.Now()
	for fileScanner.Scan() {
		text := fileScanner.Text()
		if text == "" {
			continue
		}
		data := hexutil.MustDecode("0x" + text)

		if index%10000000 == 0 {
			fmt.Println(index, time.Since(s))
			s = time.Now()
		}

		// 0x045769b6ab9a4719181badde312dafe40af6f1ee9676ad413d5d5f08989ea28f690fb1ce68a891f325303618f7c87bc5719570ff559c00be0503033eff68388768
		//     5769b6ab9a4719181badde31  // text, 12bytes
		arr = append(arr, [8]byte{data[0], data[1], data[2], data[3], data[4], data[5], data[6], data[7]})
		index++

		if index >= 4294967296 {
			fmt.Println("next index will out of bounds 4294967296")
			break
		}
	}
	fmt.Println("Loaded in ", time.Since(s0), "index", index)

	if err := fileScanner.Err(); err != nil {
		log.Fatal(err)
	}

	ind := BinarySearch8b(arr, hexutil.MustDecode("0x"+"b9faffb5adab083e"))
	fmt.Println("0000000438ada150 index is", ind)

	return &arr
}

// -1 if not found
func BinarySearch12b(arr [][12]byte, target []byte) int {
	startIndex := 0
	endIndex := len(arr) - 1
	midIndex := len(arr) / 2
	for startIndex <= endIndex {
		value := arr[midIndex]

		cmp := bytes.Compare(value[:], target) // 0 if a == b, -1 if a < b, and +1 if a > b.
		if cmp == 0 {                          // value == target
			return midIndex
		}

		if cmp == 1 { // value > target
			endIndex = midIndex - 1
			midIndex = (startIndex + endIndex) / 2
			continue
		}

		startIndex = midIndex + 1
		midIndex = (startIndex + endIndex) / 2
	}

	return -1
}

// -1 if not found
func BinarySearch8b(arr [][8]byte, target []byte) int {
	startIndex := 0
	endIndex := len(arr) - 1
	midIndex := len(arr) / 2
	for startIndex <= endIndex {
		value := arr[midIndex]

		cmp := bytes.Compare(value[:], target) // 0 if a == b, -1 if a < b, and +1 if a > b.
		if cmp == 0 {                          // value == target
			return midIndex
		}

		if cmp == 1 { // value > target
			endIndex = midIndex - 1
			midIndex = (startIndex + endIndex) / 2
			continue
		}

		startIndex = midIndex + 1
		midIndex = (startIndex + endIndex) / 2
	}

	return -1
}
