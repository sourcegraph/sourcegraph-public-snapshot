#include "textflag.h"

TEXT ·avx2Dot(SB), NOSPLIT, $0-24
	MOVQ a+0(FP), SI
	MOVQ b+8(FP), DI
	MOVQ n+16(FP), DX

	XORQ R8, R8 // return sum

	// Zero Y0, which will store 8 packed 32-bit sums
	VPXOR Y0, Y0, Y0

// In blockloop, we calculate the dot product 16 at a time
blockloop:
	CMPQ DX, $16
	JB reduce

	// Sign-extend 16 bytes into 16 int16s
	VPMOVSXBW (SI), Y1
	VPMOVSXBW (DI), Y2

	// Multiply words vertically to form doubleword intermediates,
	// then add adjacent doublewords.
	VPMADDWD Y1, Y2, Y1

	// Add results to the running sum
	VPADDD Y0, Y1, Y0

	ADDQ $16, SI
	ADDQ $16, DI
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

// In tailloop, we add to the dot product one at a time
tailloop:
	CMPQ DX, $0
	JE end

	// Load values from the input slices
	MOVBQSX (SI), R9
	MOVBQSX (DI), R10

	// Multiply and accumulate
	IMULQ R9, R10
	ADDQ R10, R8

	INCQ SI
	INCQ DI
	DECQ DX
	JMP tailloop

end:
	MOVQ R8, ret+24(FP)
	RET

