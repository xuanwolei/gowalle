package gowalle

import (
	"fmt"
	"testing"
)

func TestWriteBlockByte(t *testing.T) {
	filePath := "./test.apk"
	err := WriteBlockByte(filePath, []byte("this is custom information 1 2 3 4 5!"))
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("TestWriteBlockByte\n")
}

func TestGetBlockByte(t *testing.T) {
	filePath := "./test.apk"
	data, err := GetBlockByte(filePath)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("TestGetBlockByte:%s\n", string(data))
}
