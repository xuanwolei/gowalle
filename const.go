package gowalle

const (
	apkSigBlockMagicHi          int64 = 0x3234206b636f6c42 // LITTLE_ENDIAN, High
	apkSigBlockMagicLo          int64 = 0x20676953204b5041 // LITTLE_ENDIAN, Low
	apkSigBlockMinSize                = 32
	apkSignatureSchemeV2BlockID       = 0x7109871a
	apkChannelBlockId                 = 0x71777777

	zipEOCDRecSignature     = 0x06054b50
	zipEOCDRecMinSize       = 22
	zipEOCDCommentField     = 20
	zipEOCDCommentFieldSize = 2
	uint16MaxValue          = 65535

	zipCentralDirStartOffsetSize         = 4
	zipCentralDirStartOffsetRelativeSize = 6
)
