//go:build amd64 && !purego

#include "textflag.h"
#include "go_asm.h"

// func _matrixMultiplyVec4SSE(mat *Matrix, vecs []Vec4)
TEXT ·_matrixMultiplyVec4SSE(SB), $0-32
	MOVQ mat+0(FP), SI        // SI = mat
	MOVQ vecs_base+8(FP), DI  // DI = vecs
	MOVQ vecs_len+16(FP), CX  // CX = size
	XORQ AX, AX               // Current index

	// Exit if size is 0
	CMPQ AX, CX
	JE end

	MOVUPS 0(SI), X4       // XMM4 = Matrix[0]
	MOVUPS 16(SI), X5      // XMM5 = Matrix[1]
	MOVUPS 32(SI), X6      // XMM6 = Matrix[2]
	MOVUPS 48(SI), X7      // XMM7 = Matrix[3]

loop:
	MOVSS Vec4_X(DI), X0   // XMM0 = Vec4.X
	MOVSS Vec4_Y(DI), X1   // XMM1 = Vec4.Y
	MOVSS Vec4_Z(DI), X2   // XMM2 = Vec4.Z
	MOVSS Vec4_W(DI), X3   // XMM3 = Vec4.W

	SHUFPS $0x00, X0, X0   // Broadcast vec4.X across XMM0
	SHUFPS $0x00, X1, X1   // Broadcast vec4.Y across XMM1
	SHUFPS $0x00, X2, X2   // Broadcast vec4.Z across XMM2
	SHUFPS $0x00, X3, X3   // Broadcast vec4.W across XMM3

	MULPS X4, X0           // Matrix[0] * Vec4.X
	MULPS X5, X1           // Matrix[1] * Vec4.Y
	MULPS X6, X2           // Matrix[2] * Vec4.Z
	MULPS X7, X3           // Matrix[3] * Vec4.W

	ADDPS X1, X0           // Add XMM1 to XMM0
	ADDPS X2, X0           // Add XMM2 to XMM0
	ADDPS X3, X0           // Add XMM3 to XMM0

	MOVUPS X0, 0(DI)       // arr[i] = XMM0
	ADDQ $Vec4__size, DI   // Move to the next Vec4

	INCQ AX                // i++
	CMPQ AX, CX            // i < size
	JL loop                // continue

end:
	RET
