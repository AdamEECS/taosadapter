package ctools

import (
	"unsafe"

	"github.com/taosdata/driver-go/v2/common"
	"github.com/taosdata/driver-go/v2/wrapper"
	"github.com/taosdata/taosadapter/tools/jsonbuilder"
)

func IsVarDataType(colType uint8) bool {
	return colType == common.TSDB_DATA_TYPE_BINARY || colType == common.TSDB_DATA_TYPE_NCHAR
}

func BitmapLen(n int) int {
	return ((n) + ((1 << 3) - 1)) >> 3
}

func BitPos(n int) int {
	return n & ((1 << 3) - 1)
}

func CharOffset(n int) int {
	return n >> 3
}

func BMIsNull(c byte, n int) bool {
	return c&(1<<(7-BitPos(n))) == (1 << (7 - BitPos(n)))
}

func ItemIsNull(pHeader uintptr, row int) bool {
	offset := CharOffset(row)
	c := *((*byte)(unsafe.Pointer(pHeader + uintptr(offset))))
	if BMIsNull(c, row) {
		return true
	}
	return false
}

func WriteRawJsonBool(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	if (*((*byte)(unsafe.Pointer(pStart + uintptr(row)*1)))) != 0 {
		builder.WriteTrue()
	} else {
		builder.WriteFalse()
	}
}

func WriteRawJsonTinyint(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	builder.WriteInt8(*((*int8)(unsafe.Pointer(pStart + uintptr(row)*1))))
}

func WriteRawJsonSmallint(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	builder.WriteInt16(*((*int16)(unsafe.Pointer(pStart + uintptr(row)*2))))
}

func WriteRawJsonInt(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	builder.WriteInt32(*((*int32)(unsafe.Pointer(pStart + uintptr(row)*4))))
}

func WriteRawJsonBigint(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	builder.WriteInt64(*((*int64)(unsafe.Pointer(pStart + uintptr(row)*8))))
}

func WriteRawJsonUTinyint(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	builder.WriteUint8(*((*uint8)(unsafe.Pointer(pStart + uintptr(row)*1))))
}

func WriteRawJsonUSmallint(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	builder.WriteUint16(*((*uint16)(unsafe.Pointer(pStart + uintptr(row)*2))))
}

func WriteRawJsonUInt(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	builder.WriteUint32(*((*uint32)(unsafe.Pointer(pStart + uintptr(row)*4))))
}

func WriteRawJsonUBigint(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	builder.WriteUint64(*((*uint64)(unsafe.Pointer(pStart + uintptr(row)*8))))
}

func WriteRawJsonFloat(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	builder.WriteFloat32(*((*float32)(unsafe.Pointer(pStart + uintptr(row)*4))))
}

func WriteRawJsonDouble(builder *jsonbuilder.Stream, pStart uintptr, row int) {
	builder.WriteFloat64(*((*float64)(unsafe.Pointer(pStart + uintptr(row)*8))))
}

func WriteRawJsonTime(builder *jsonbuilder.Stream, pStart uintptr, row int, precision int, timeFormat FormatTimeFunc) {
	value := *((*int64)(unsafe.Pointer(pStart + uintptr(row)*8)))
	if value == wrapper.CTimestampNull {
		builder.WriteNil()
	} else {
		timeFormat(builder, value, precision)
	}
}

func WriteRawJsonBinary(builder *jsonbuilder.Stream, pHeader, pStart uintptr, row int) {
	offset := *((*int32)(unsafe.Pointer(pHeader + uintptr(row*4))))
	if offset == -1 {
		builder.WriteNil()
		return
	}
	currentRow := unsafe.Pointer(pStart + uintptr(offset))
	clen := *((*int16)(currentRow))
	currentRow = unsafe.Pointer(uintptr(currentRow) + 2)

	builder.WriteByte('"')
	for index := int16(0); index < clen; index++ {
		builder.WriteStringByte(*((*byte)(unsafe.Pointer(uintptr(currentRow) + uintptr(index)))))
	}
	builder.WriteByte('"')
}

func WriteRawJsonNchar(builder *jsonbuilder.Stream, pHeader, pStart uintptr, row int) {
	offset := *((*int32)(unsafe.Pointer(pHeader + uintptr(row*4))))
	if offset == -1 {
		builder.WriteNil()
		return
	}
	currentRow := unsafe.Pointer(pStart + uintptr(offset))
	clen := *((*int16)(currentRow)) / 4
	currentRow = unsafe.Pointer(uintptr(currentRow) + 2)

	for index := int16(0); index < clen; index++ {
		builder.WriteRuneString(*((*rune)(unsafe.Pointer(uintptr(currentRow) + uintptr(index*4)))))
	}
}

// just like nchar
func WriteRawJsonJson(builder *jsonbuilder.Stream, pHeader, pStart uintptr, row int) {
	offset := *((*int32)(unsafe.Pointer(pHeader + uintptr(row*4))))
	if offset == -1 {
		builder.WriteNil()
		return
	}
	currentRow := unsafe.Pointer(pStart + uintptr(offset))
	clen := *((*int16)(currentRow)) / 4
	currentRow = unsafe.Pointer(uintptr(currentRow) + 2)

	for index := int16(0); index < clen; index++ {
		builder.WriteRune(*((*rune)(unsafe.Pointer(uintptr(currentRow) + uintptr(index*4)))))
	}
}

// ReadBlock in-place
func JsonWriteRawBlock(builder *jsonbuilder.Stream, colType uint8, pHeader, pStart uintptr, row int, precision int, timeFormat FormatTimeFunc) {
	if IsVarDataType(colType) {
		switch colType {
		case uint8(common.TSDB_DATA_TYPE_BINARY):
			WriteRawJsonBinary(builder, pHeader, pStart, row)
		case uint8(common.TSDB_DATA_TYPE_NCHAR):
			WriteRawJsonNchar(builder, pHeader, pStart, row)
		case uint8(common.TSDB_DATA_TYPE_JSON):
			WriteRawJsonJson(builder, pHeader, pStart, row)
		}
	} else {
		if ItemIsNull(pHeader, row) {
			builder.WriteNil()
		} else {
			switch colType {
			case uint8(common.TSDB_DATA_TYPE_BOOL):
				WriteRawJsonBool(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_TINYINT):
				WriteRawJsonTinyint(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_SMALLINT):
				WriteRawJsonSmallint(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_INT):
				WriteRawJsonInt(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_BIGINT):
				WriteRawJsonBigint(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_UTINYINT):
				WriteRawJsonUTinyint(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_USMALLINT):
				WriteRawJsonUSmallint(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_UINT):
				WriteRawJsonUInt(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_UBIGINT):
				WriteRawJsonUBigint(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_FLOAT):
				WriteRawJsonFloat(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_DOUBLE):
				WriteRawJsonDouble(builder, pStart, row)
			case uint8(common.TSDB_DATA_TYPE_TIMESTAMP):
				WriteRawJsonTime(builder, pStart, row, precision, timeFormat)
			}
		}
	}
}
