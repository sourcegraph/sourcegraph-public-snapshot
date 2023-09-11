#include "textflag.h"

// dotSIMD uses NEON instructions and the SDOT instruction
// (required after Armv8.4) to compute the dot product of
// two int8 vectors.
//
// The vectors must be of the same length, and that length
// must be a multiple of 16.
TEXT ·dotSIMD(SB), NOSPLIT, $0-52
	// Offsets based on slice header offsets.
	// To check, use `GOARCH=arm64 go vet`
	MOVD a_base+0(FP), R4
	MOVD b_base+24(FP), R5
	MOVD a_len+8(FP), R6

	ADD R4, R6, R6 // end pointer

	// Zero V0, which will store 4 packed 32-bit sums
	VEOR V0.B16, V0.B16, V0.B16

blockloop:
	CMP R4, R6
	BEQ reduce

	// Load 16 bytes from each slice, post-incrementing the pointers
	VLD1.P 16(R4), [V1.B16]
	VLD1.P 16(R5), [V2.B16]

	// The following instruction is not supported by the go assembler, so use
	// the binary format. It would be the equivalent of the following instruction:
	//
    // VSDOT V1.B16, V2.B16, V0.S4
	//
	// I generated the binary form of the instruction using this godbolt setup:
	// https://godbolt.org/z/3jPohn4dn
	WORD $0x4E829420

	JMP blockloop

reduce:
	VADDV V0.S4, V0
	VMOV V0.S[0], R6
	MOVD R6, ret+48(FP)
	RET

