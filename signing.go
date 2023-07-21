package gowalle

import (
	"encoding/binary"
	"os"
)

type ApkSigningPayload struct {
	id    int
	value []byte
}

type ApkSigningBlock struct {
	payloads []*ApkSigningPayload
}

func (block *ApkSigningBlock) AddPayload(payload *ApkSigningPayload) {
	block.payloads = append(block.payloads, payload)
}

func (block *ApkSigningBlock) WriteApkSigningBlock(dataOutput *os.File) (int64, error) {
	var length int64 = 24 // 24 = 8(size of block in bytes—same as the very first field (uint64)) + 16 (magic “APK Sig Block 42” (16 bytes))

	// 计算block总长度
	for _, payload := range block.payloads {
		length += 12 + int64(len(payload.value)) // 12 = 8(uint64-length-prefixed) + 4 (ID (uint32))
	}

	// 写入block总长度
	if err := binary.Write(dataOutput, binary.LittleEndian, length); err != nil {
		return 0, err
	}

	// 写入ID-value
	for _, payload := range block.payloads {
		payloadLength := int64(len(payload.value) + 4) // 4 (ID占4个字节，所以长度需要加上4)
		if err := binary.Write(dataOutput, binary.LittleEndian, payloadLength); err != nil {
			return 0, err
		}

		if err := binary.Write(dataOutput, binary.LittleEndian, int32(payload.id)); err != nil {
			return 0, err
		}

		if _, err := dataOutput.Write(payload.value); err != nil {
			return 0, err
		}
	}

	// 再次写入block总长度
	if err := binary.Write(dataOutput, binary.LittleEndian, length); err != nil {
		return 0, err
	}

	// 写入魔数
	if err := binary.Write(dataOutput, binary.LittleEndian, apkSigBlockMagicLo); err != nil {
		return 0, err
	}

	if err := binary.Write(dataOutput, binary.LittleEndian, apkSigBlockMagicHi); err != nil {
		return 0, err
	}

	return length, nil
}
