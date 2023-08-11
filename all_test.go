package gowalle

import (
	"fmt"
	"testing"
	"time"
)

//func BenchmarkWriteBlockByte(b *testing.B) {
//	now := time.Now()
//	filePath := "./test.apk"
//	err := WriteBlockByte(filePath, []byte("this is custom information 1 2 3 4 5!"))
//	if err != nil {
//		b.Error(err)
//	}
//	CopyFile(filePath, "./test.apk.bak", false)
//	time.Since(now).Milliseconds()
//}

//
func TestWriteBlockByte(t *testing.T) {
	now := time.Now()
	filePath := "./test.apk"
	destinationFile := "./test_channel.apk"
	CopyFile(filePath, destinationFile, false)
	err := WriteBlockByte(destinationFile, []byte("this is custom information 2"))
	if err != nil {
		t.Error(err)
	}
	time.Since(now).Milliseconds()
	fmt.Printf("TestWriteBlockByte:%d\n", time.Since(now).Microseconds())
}

func TestGetBlockByte(t *testing.T) {
	filePath := "./test_channel.apk"
	data, err := GetBlockByte(filePath)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("TestGetBlockByte:%s\n", string(data))
}
