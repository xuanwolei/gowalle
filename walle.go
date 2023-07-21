package gowalle

import (
	"fmt"
	"os"
)

type ApkAttribute struct {
	CommentLength         int64
	CentralDirStartOffset int64
	ApkSigBlock           []byte
	ApkSigBlockOffset     int64
	OriginIdValues        map[int][]byte
}

func GetBlockByte(filePath string) ([]byte, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("Error opening file:%v", err)
	}
	defer file.Close()
	attr, err := parseApkAttribute(file)
	if err != nil {
		return nil, err
	}
	return attr.OriginIdValues[apkChannelBlockId], nil
}

func WriteBlockByte(filePath string, data []byte) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("Error opening file:%v", err)
	}
	defer file.Close()
	attr, err := parseApkAttribute(file)
	if err != nil {
		return err
	}

	idValues := map[int][]byte{
		apkChannelBlockId: data,
	}
	apkSigningBlock := genApkSigningBlock(idValues, attr.OriginIdValues)
	err = updateApkSigning(file, attr.CentralDirStartOffset, attr.ApkSigBlockOffset, attr.CommentLength, apkSigningBlock)
	if err != nil {
		return fmt.Errorf("Error updateApkSigning:%v", err)
	}
	return nil
}

func parseApkAttribute(file *os.File) (*ApkAttribute, error) {

	commentLength, err := getCommentLength(file)
	if err != nil {
		return nil, fmt.Errorf("Error getting comment length:%v", err)
	}

	centralDirStartOffset, err := findCentralDirStartOffset(file, commentLength)
	if err != nil {
		return nil, fmt.Errorf("Error finding central directory start offset:%v", err)
	}

	apkSigBlock, apkSigBlockOffset, err := findApkSigningBlock(file, centralDirStartOffset)
	if err != nil {
		return nil, fmt.Errorf("Error finding APK Signing Block:%v", err)
	}

	originIdValues, err := findIdValues(apkSigBlock)
	if err != nil {
		return nil, fmt.Errorf("Error finding id values:%v", err)
	}

	return &ApkAttribute{
		CommentLength:         int64(commentLength),
		CentralDirStartOffset: centralDirStartOffset,
		ApkSigBlock:           apkSigBlock,
		ApkSigBlockOffset:     apkSigBlockOffset,
		OriginIdValues:        originIdValues,
	}, nil
}
