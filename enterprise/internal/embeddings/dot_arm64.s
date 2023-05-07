#include "textflag.h"

#define BLOCKSIZE $16

TEXT ·dotSIMD(SB), NOSPLIT, $0-56
	// Offsets based on slice header offsets.
	// To check, use `GOARCH=amd64 go vet`
	MOVD a_base+0(FP), R4
	MOVD b_base+24(FP), R5
	MOVD a_len+8(FP), R6


	// Zero V0, which will store 4 packed 32-bit sums
	VEOR V0.B16, V0.B16, V0.B16

blockloop:
    CMP BLOCKSIZE, R6
    BLT reduce

    VLD1 (R4), [V1.B16]
    VLD1 (R5), [V2.B16]

    // The following instruction is not supported
    // by the go assembler, so use the binary format.
    // VSDOT V1.B16, V2.B16, V0.S4
    WORD $0x4E829420

	ADD BLOCKSIZE, R4
	ADD BLOCKSIZE, R5
	SUB BLOCKSIZE, R6
	JMP blockloop

reduce:
    VADDV V0.S4, V0
    VMOV V0.S[0], R6
    MOVD R6, ret+48(FP)
	RET

