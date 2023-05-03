#include "textflag.h"

TEXT ·avx2Dot(SB), NOSPLIT, $0-56
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

// In tailloop, we add to the dot product one at a time
tailloop:
	CMPQ DX, $0
	JE end

	// Load values from the input slices
	MOVBQSX (AX), R9
	MOVBQSX (BX), R10

	// Multiply and accumulate
	IMULQ R9, R10
	ADDQ R10, R8

	INCQ AX
	INCQ BX
	DECQ DX
	JMP tailloop

end:
	MOVQ R8, ret+48(FP)
	RET

