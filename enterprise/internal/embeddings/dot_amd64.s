#include "textflag.h"

TEXT ·dotAVX2(SB), NOSPLIT, $0-52
	// Offsets based on slice header offsets.
	// To check, use `GOARCH=amd64 go vet`
	MOVQ a_base+0(FP), AX
	MOVQ b_base+24(FP), BX
	MOVQ a_len+8(FP), DX

	XORQ R8, R8 // return sum

	// Zero Y0, which will store 8 packed 32-bit sums
	VPXOR Y0, Y0, Y0

// In blockloop, we calculate the dot product 16 at a time
blockloop:
	CMPQ DX, $16
	JB reduce

	// Sign-extend 16 bytes into 16 int16s
	VPMOVSXBW (AX), Y1
	VPMOVSXBW (BX), Y2

	// Multiply words vertically to form doubleword intermediates,
	// then add adjacent doublewords.
	VPMADDWD Y1, Y2, Y1

	// Add results to the running sum
	VPADDD Y0, Y1, Y0

	ADDQ $16, AX
	ADDQ $16, BX
	SUBQ $16, DX
	JMP blockloop

reduce:
	// X0 is the low bits of Y0.
	// Extract the high bits into X1, fold in half, add, repeat.
	VEXTRACTI128 $1, Y0, X1
	VPADDD X0, X1, X0

	VPSRLDQ $8, X0, X1
	VPADDD X0, X1, X0

	VPSRLDQ $4, X0, X1
	VPADDD X0, X1, X0

	// Store the reduced sum
	VMOVD X0, R8

	// For the remainder of the function, we do not use vector instructions.
	// Zero the registers to avoid the SSE penalty.
	VZEROALL

end:
	MOVD R8, ret+48(FP)
	RET

TEXT ·dotAVX512(SB), NOSPLIT, $0-52
	// Offsets based on slice header offsets.
	// To check, use `GOARCH=amd64 go vet`
	MOVQ a_base+0(FP), AX
	MOVQ b_base+24(FP), BX
	MOVQ a_len+8(FP), DX

    ADDQ AX, DX // end pointer

	// Zero our accumulators
	VPXORQ Z0, Z0, Z0 // positive
	VPXORQ Z1, Z1, Z1 // negative

	// Fill Z3 with 128
	MOVD $0x80808080, R9
	VPBROADCASTD R9, Z2

blockloop:
	CMPQ AX, DX
	JE reduce

	VMOVDQU8 (AX), Z3
	VMOVDQU8 (BX), Z4

	// The VPDPBUSD instruction calculates of the dot product of 4 lanes at a
	// time, accumulating into an i32 vector. The problem is it expects one
	// vector to be unsigned bytes and one to be signed bytes. To make this
	// work, we make one of our vectors unsigned by adding 128 to each element.
	// This causes us to overshoot, so we keep track of the amount we need
	// to compensate by so we can subtract it from the sum at the end.
	//
	// Effectively, we are calculating SUM((Z3 + 128) * Z4) - SUM(128 * Z4).

	VPADDB Z3, Z2, Z3
	VPDPBUSD Z3, Z4, Z0
	VPDPBUSD Z2, Z4, Z1

	ADDQ $64, AX
	ADDQ $64, BX
	JMP blockloop

reduce:
	VPSUBD Z1, Z0, Z0

    VEXTRACTI64X4 $1, Z0, Y1
    VPADDD Y0, Y1, Y0

	VEXTRACTI128 $1, Y0, X1
	VPADDD X0, X1, X0

	VPSRLDQ $8, X0, X1
	VPADDD X0, X1, X0

	VPSRLDQ $4, X0, X1
	VPADDD X0, X1, X0

	// Store the reduced sum
	VMOVD X0, R8

end:
	MOVD R8, ret+48(FP)
	VZEROALL
	RET
