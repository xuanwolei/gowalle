package gowalle

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// getCommentLength 获取注释的长度
func getCommentLength(file *os.File) (int, error) {
	// End of central directory record (EOCD)
	// Offset    Bytes     Description[23]
	// 0           4       End of central directory signature = 0x06054b50
	// 4           2       Number of this disk
	// 6           2       Disk where central directory starts
	// 8           2       Number of central directory records on this disk
	// 10          2       Total number of central directory records
	// 12          4       Size of central directory (bytes)  // 核心目录大小
	// 16          4       Offset of start of central directory, relative to start of archive
	// 20          2       Comment length (n)
	// 22          n       Comment
	// For a zip with no archive comment, the
	// end-of-central-directory record will be 22 bytes long, so
	// we expect to find the EOCD marker 22 bytes from the end.

	// 获取文件的大小
	archiveSize, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		return 0, err
	}

	if archiveSize < zipEOCDRecMinSize {
		return 0, errors.New("APK too small for ZIP End of Central Directory (EOCD) record")
	}

	// 计算最大可能的注释长度
	maxCommentLength := int64(archiveSize - zipEOCDRecMinSize)
	if maxCommentLength > uint16MaxValue {
		maxCommentLength = uint16MaxValue
	}

	// 获取EOCD结尾位置
	eocdWithEmptyCommentStartPosition := archiveSize - zipEOCDRecMinSize

	// 循环寻找EOCD记录
	for expectedCommentLength := 0; expectedCommentLength <= int(maxCommentLength); expectedCommentLength++ {
		// 计算EOCD开始位置
		eocdStartPos := eocdWithEmptyCommentStartPosition - int64(expectedCommentLength)

		// 读取4个字节，用于判断是否为EOCD记录
		byteBuffer := make([]byte, 4)
		_, err := file.ReadAt(byteBuffer, eocdStartPos)
		if err != nil {
			return 0, err
		}

		if binary.LittleEndian.Uint32(byteBuffer) == zipEOCDRecSignature {
			// 读取comment length（占2个字节）
			commentLengthByteBuffer := make([]byte, zipEOCDCommentFieldSize)
			_, err := file.ReadAt(commentLengthByteBuffer, eocdStartPos+zipEOCDCommentField)
			if err != nil {
				return 0, err
			}

			actualCommentLength := int(binary.LittleEndian.Uint16(commentLengthByteBuffer))
			if actualCommentLength == expectedCommentLength {
				return actualCommentLength, nil
			}
		}
	}

	return 0, errors.New("ZIP End of Central Directory (EOCD) record not found")
}

// findCentralDirStartOffset 查找中央目录开始的偏移量
func findCentralDirStartOffset(file *os.File, commentLength int) (int64, error) {
	// 创建一个字节缓冲区来存储中央目录开始的偏移量
	zipCentralDirectoryStart := make([]byte, zipCentralDirStartOffsetSize)
	file.Seek(-int64(commentLength+zipCentralDirStartOffsetRelativeSize), os.SEEK_END)
	_, err := file.Read(zipCentralDirectoryStart)
	if err != nil {
		return 0, err
	}

	// 将读取的字节缓冲区转换为int32类型，注意采用LittleEndian字节序
	centralDirStartOffset := int64(binary.LittleEndian.Uint32(zipCentralDirectoryStart))
	return centralDirStartOffset, nil
}

// findApkSigningBlock 查找APK签名块
func findApkSigningBlock(file *os.File, centralDirOffset int64) ([]byte, int64, error) {
	if centralDirOffset < apkSigBlockMinSize {
		return nil, 0, fmt.Errorf("APK too small for APK Signing Block. ZIP Central Directory offset: %d", centralDirOffset)
	}

	// 从签名块的footer部分读取魔数和文件偏移量
	footer := make([]byte, 24)
	file.Seek(centralDirOffset-24, os.SEEK_SET)
	_, err := file.Read(footer)
	if err != nil {
		return nil, 0, err
	}

	magicLo := binary.LittleEndian.Uint64(footer[8:16])
	magicHi := binary.LittleEndian.Uint64(footer[16:24])

	if magicLo != uint64(apkSigBlockMagicLo) || magicHi != uint64(apkSigBlockMagicHi) {
		return nil, 0, errors.New("No APK Signing Block before ZIP Central Directory")
	}

	// 读取大小字段并进行比较
	apkSigBlockSizeInFooter := binary.LittleEndian.Uint64(footer[:8])
	if apkSigBlockSizeInFooter < 32 || apkSigBlockSizeInFooter > uint64(^uint32(0))-8 {
		return nil, 0, fmt.Errorf("APK Signing Block size out of range: %d", apkSigBlockSizeInFooter)
	}

	totalSize := int(apkSigBlockSizeInFooter) + 8 // +8 (APK签名块大小字段所占字节数)
	apkSigBlockOffset := centralDirOffset - int64(totalSize)
	if apkSigBlockOffset < 0 {
		return nil, 0, fmt.Errorf("APK Signing Block offset out of range: %d", apkSigBlockOffset)
	}

	// 读取整个APK签名块
	apkSigBlock := make([]byte, totalSize)
	file.Seek(apkSigBlockOffset, os.SEEK_SET)
	_, err = file.Read(apkSigBlock)
	if err != nil {
		return nil, 0, err
	}

	// 检查头部中的签名块大小字段是否与尾部相匹配
	apkSigBlockSizeInHeader := binary.LittleEndian.Uint64(apkSigBlock[:8])
	if apkSigBlockSizeInHeader != apkSigBlockSizeInFooter {
		return nil, 0, fmt.Errorf("APK Signing Block sizes in header and footer do not match: %d vs %d", apkSigBlockSizeInHeader, apkSigBlockSizeInFooter)
	}

	return apkSigBlock, apkSigBlockOffset, nil
}

// findIdValues 查找APK签名块中的id值
func findIdValues(apkSigningBlock []byte) (map[int][]byte, error) {
	// 检查字节序是否为LittleEndian
	//if binary.LittleEndian.Uint64(apkSigningBlock[16:24]) != apkSigBlockMagicLo || binary.LittleEndian.Uint64(apkSigningBlock[8:16]) != apkSigBlockMagicHi {
	//	return nil, errors.New("Invalid APK Signing Block magic")
	//}

	// 获取pairs的字节缓冲区
	pairs := apkSigningBlock[8 : len(apkSigningBlock)-24]

	idValues := make(map[int][]byte)
	entryCount := 0

	for len(pairs) >= 8 {
		entryCount++

		lenLong := int64(binary.LittleEndian.Uint64(pairs[:8]))
		pairs = pairs[8:]

		if lenLong < 4 || lenLong > int64(^uint32(0)) {
			return nil, fmt.Errorf("APK Signing Block entry #%d size out of range: %d", entryCount, lenLong)
		}

		length := int(lenLong)
		if length > len(pairs) {
			return nil, fmt.Errorf("APK Signing Block entry #%d size out of range: %d, available: %d", entryCount, length, len(pairs))
		}

		id := int(binary.LittleEndian.Uint32(pairs[:4]))
		idValues[id] = append([]byte(nil), pairs[4:length]...)

		pairs = pairs[length:]
	}

	return idValues, nil
}

func genApkSigningBlock(idValues map[int][]byte, originIdValues map[int][]byte) *ApkSigningBlock {
	// 把已有的和新增的 ID-value 添加到 originIdValues
	if idValues != nil {
		for id, value := range idValues {
			originIdValues[id] = value
		}
	}

	apkSigningBlock := &ApkSigningBlock{}
	var ids []int
	for id, value := range originIdValues {
		ids = append(ids, id)
		payload := &ApkSigningPayload{
			id:    id,
			value: value,
		}
		apkSigningBlock.AddPayload(payload)
	}
	//按照id降序排列
	//sort.Sort(sort.Reverse(sort.IntSlice(ids)))
	//for _, id := range ids {
	//	payload := &ApkSigningPayload{
	//		id:    id,
	//		value: originIdValues[id],
	//	}
	//	apkSigningBlock.AddPayload(payload)
	//}

	return apkSigningBlock
}

func updateApkSigning(fIn *os.File, centralDirStartOffset int64, apkSigBlockOffset int64, commentLength int64, apkSigningBlock *ApkSigningBlock) error {
	// 读取核心目录的内容
	fIn.Seek(centralDirStartOffset, os.SEEK_SET)
	// 更新文件的总长度
	fileInfo, err := fIn.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()
	centralDirBytes := make([]byte, fileSize-centralDirStartOffset)
	if _, err := fIn.Read(centralDirBytes); err != nil {
		return err
	}

	// 更新签名块
	fIn.Seek(apkSigBlockOffset, os.SEEK_SET)
	fIn.Truncate(apkSigBlockOffset)

	// 写入新的签名块，返回的长度是不包含签名块头部的 Size of block（8字节）
	lengthExcludeHSOB, err := apkSigningBlock.WriteApkSigningBlock(fIn)
	if err != nil {
		return err
	}

	// 更新核心目录
	fIn.Write(centralDirBytes)
	// 更新文件的总长度
	fileInfo, err = fIn.Stat()
	if err != nil {
		return err
	}
	fileSize = fileInfo.Size()
	fIn.Truncate(fileSize)
	// 更新 EOCD 所记录的核心目录的偏移
	fIn.Seek(fileSize-commentLength-6, os.SEEK_SET)
	// 6 = 2(Comment length) + 4 (Offset of start of central directory, relative to start of archive)
	temp := make([]byte, 4)
	oldSignBlockLength := centralDirStartOffset - apkSigBlockOffset // 旧签名块字节数
	newSignBlockLength := lengthExcludeHSOB + 8                     // 新签名块字节数, 8 = size of block in bytes (excluding this field) (uint64)
	extraLength := newSignBlockLength - oldSignBlockLength
	binary.LittleEndian.PutUint32(temp, uint32(int32(centralDirStartOffset+extraLength)))
	fIn.Write(temp)

	return nil
}

func CopyFile(sourceFile, destinationFile string, sync bool) error {
	// 打开源文件
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	// 创建目标文件
	destination, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer destination.Close()

	// 复制文件内容
	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	// 可选：如果你想确保数据被写入磁盘，可以使用Sync方法
	if sync {
		err = destination.Sync()
		if err != nil {
			return err
		}
	}

	return nil
}
