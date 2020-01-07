package main

import (
	"encoding/binary"
	"math"
)

//Converts bytes into a 64-bit float
func byte2Float64(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

//Converts Floats into a 8 bytes
func float642Byte(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}
