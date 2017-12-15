#include "textflag.h"

TEXT ·bytesToLowerASCII(SB),NOSPLIT,$0
	// use the smaller of the two lengths to avoid out-of-bounds writes
	MOVQ	dst_len+8(FP), BX
	MOVQ	src_len+32(FP), DX
	CMPQ	DX, BX
	JLT	2(PC)
	MOVQ	DX, BX
	// BX contains length, DX is now a scratch reg
	MOVQ	dst_base+0(FP), AX
	MOVQ	src_base+24(FP), CX

	// Handle small amounts of work with purely scalar code.
	CMPQ	BX, $16
	JLT	small

	// broadcast fills register x with bytes b.
	// It is a bit awkward, but it uses only <= SSE2 instructions
	// for maximum compatibility. It is cribbed from the Go core runtime/indexbyte.
#define broadcast(b, x)	MOVQ b, DX; MOVD DX, x; PUNPCKLBW x, x; PUNPCKLBW x, x; PSHUFL $0, x, x
	broadcast($193, X0)  //  193 == 'A'+128
	broadcast($-103, X2) // -103 == -128 + 25
	broadcast($32, X3)   // 32 == 'A'-'a'

loop:
	// process 16 bytes
	MOVOU	(CX), X1
	PSUBB	X0, X1
	PCMPGTB X2, X1
	// X1 contains 0xff for all capital letters
	PANDN	X3, X1
	// X1 contains 32 for all capital letters
	MOVOU	(CX), X4
	PADDB	X1, X4
	MOVOU	X4, (AX)

	// reduce work remaining by 16, return if appropriate
	SUBQ	$16, BX
	JLE	done // SUBQ sets flags; if BX-16 is <= 0, we are done

	// dx will hold how big a step to take forward.
	// if there are less than 16 bytes, take a smaller step and reprocess some data.
	// TODO: if dx >= 16, but AX/CX is not aligned, take a smaller step to get into alignment
	MOVQ	$16, DX
	CMPQ	BX, $16
	JGE	2(PC)
	MOVQ	BX, DX

	// adjust pointers
	ADDQ	DX, CX
	ADDQ	DX, AX
	JMP loop

small:
	CMPQ	BX, $0
	JE	done
	// if R9 >= 'A' && R9 <= 'Z' { R9 -= 'A'-'a' }
	MOVB	(CX), R9
	CMPB	R9, $65 // 65 == 'A'
	JLT	ignore
	CMPB	R9, $90 // 90 == 'Z'
	JGT	ignore
	ADDB	$32, R9 // 32 == 'A'-'a'
ignore:
	MOVB	R9, (AX)
	// advance pointers, decrement work
	INCQ	AX
	INCQ	CX
	DECQ	BX
	JMP	small

done:
	RET

