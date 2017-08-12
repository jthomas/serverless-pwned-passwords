package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/willf/bloom"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const BloomFileDirectory = "/bloom_filters/"

func HashString(source string) (hash string) {
	hash_bytes := sha1.Sum([]byte(source))
	hex_string := hex.EncodeToString(hash_bytes[:])
	return strings.ToUpper(hex_string)
}

func BucketForHash(hash string) string {
	if len(hash) != 40 {
		log.Fatal("invalid hash length, must be 40 characters")
	}
	return strings.ToLower(hash[0:2])
}

func BloomFileForBucket(bucket string) string {
	bloomFile := fmt.Sprintf("%s.dat", bucket)
	return filepath.Join(BloomFileDirectory, bloomFile)
}

func DecodedBloomFilter(b []byte) *bloom.BloomFilter {
	log.Println("decoding bloom filter from", len(b), "bytes")
	filter := bloom.New(1, 1)
	start := time.Now()
	err := filter.GobDecode(b)
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)
	log.Println("decoded bloom filter parameters -> m:", filter.Cap(), "k:", filter.K())
	log.Printf("decoding bloom filter took %s", elapsed)

	return filter
}

func TestBloomForHash(bucket string, hash string) bool {
	file := BloomFileForBucket(bucket)
	bloom := DecodedBloomFilter(ReadFile(file))
	return bloom.Test([]byte(hash))
}

func ReadFile(path string) []byte {
	log.Println("reading file @", path)

	start := time.Now()
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)

	log.Println("file contained", len(b), "bytes")
	log.Printf("elapsed time: %s", elapsed)

	return b
}

func PrintResponse(found bool) {
	msg := map[string]bool{"found": found}
	res, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(res))
}

func ParseInput() string {
	if len(os.Args) < 2 {
		log.Fatal("Missing programme argument with input parameters.")
	}
	arg := os.Args[1]

	var obj map[string]interface{}
	json.Unmarshal([]byte(arg), &obj)
	password, ok := obj["password"].(string)
	if !ok {
		log.Fatal("Unable to parse input parameters for password.")
	}

	return password
}

func main() {
	password := ParseInput()
	log.Println("checking password:", password)

	hash := HashString(password)
	bucket := BucketForHash(hash)
	log.Println("hash:", hash)

	found := TestBloomForHash(bucket, hash)
	log.Println("found password in bloom filter:", found)
	PrintResponse(found)
}
