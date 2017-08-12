package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/willf/bloom"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const N = uint(17971985)
const K = uint(10)

func BucketForHash(hash string) (bucket string, err error) {
	if len(hash) != 40 {
		return "", errors.New("invalid hash length, must be 40 characters")
	}
	return strings.ToLower(hash[0:2]), nil
}

func GenerateBloomFilters(buckets int) map[string]*bloom.BloomFilter {
	log.Println("Initialising", buckets, "bloom filters...")
	filters := make(map[string]*bloom.BloomFilter)

	for i := 0; i < buckets; i++ {
		bucket := fmt.Sprintf("%02x", i)
		filter := bloom.New(N, K)
		filters[bucket] = filter
	}

	log.Println("Bloom filter buckets:", len(filters))

	return filters
}

func AddHashesToBloomFilters(file string, filters map[string]*bloom.BloomFilter) {
	log.Println("Reading hashes from file: ", file)
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		hash := scanner.Text()
		bucket, err := BucketForHash(hash)
		if err != nil {
			log.Fatal(err)
		}

		filters[bucket].AddString(hash)
		count++
		if (count % 1000000) == 0 {
			log.Println("Processed", count, "hashes...")
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input file:", err)
	}
	log.Println("Added", count, "hashes to bloom filters")
}

func SerialiseBloomFilters(filters map[string]*bloom.BloomFilter) {
	log.Println("Serialising", len(filters), "bloom filters...")

	if _, err := os.Stat("bloom_filters"); os.IsNotExist(err) {
		os.Mkdir("bloom_filters", 0755)
	}

	for bucket, filter := range filters {
		encoded, err := filter.GobEncode()
		log.Println("bucket:", bucket, "encoded bytes:", len(encoded))

		bloomFile := fmt.Sprintf("bloom_filters/%s.dat", bucket)
		err = ioutil.WriteFile(bloomFile, encoded, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing mandatory hash file argument.")
	}
	log.Println("Creating Bloom filters with parameters --> n:", N, "k:", K)
	filters := GenerateBloomFilters(256)

	for _, file := range os.Args[1:] {
		AddHashesToBloomFilters(file, filters)
	}

	SerialiseBloomFilters(filters)
}
