	.section	__TEXT,__text,regular,pure_instructions
	.build_version macos, 14, 0	sdk_version 14, 4
	.globl	_matrix_multiply_vec4           ## -- Begin function matrix_multiply_vec4
	.p2align	4, 0x90
_matrix_multiply_vec4:                  ## @matrix_multiply_vec4
## %bb.0:
	pushq	%rbp
	movq	%rsp, %rbp
	andq	$-8, %rsp
	testq	%rdx, %rdx
	je	LBB0_2
	.p2align	4, 0x90
LBB0_1:                                 ## =>This Inner Loop Header: Depth=1
	movss	(%rsi), %xmm0                   ## xmm0 = mem[0],zero,zero,zero
	movss	4(%rsi), %xmm1                  ## xmm1 = mem[0],zero,zero,zero
	movss	8(%rsi), %xmm2                  ## xmm2 = mem[0],zero,zero,zero
	movss	12(%rsi), %xmm3                 ## xmm3 = mem[0],zero,zero,zero
	shufps	$0, %xmm0, %xmm0                ## xmm0 = xmm0[0,0,0,0]
	shufps	$0, %xmm1, %xmm1                ## xmm1 = xmm1[0,0,0,0]
	shufps	$0, %xmm2, %xmm2                ## xmm2 = xmm2[0,0,0,0]
	shufps	$0, %xmm3, %xmm3                ## xmm3 = xmm3[0,0,0,0]
	mulps	(%rdi), %xmm0
	mulps	16(%rdi), %xmm1
	mulps	32(%rdi), %xmm2
	addps	%xmm0, %xmm1
	mulps	48(%rdi), %xmm3
	addps	%xmm2, %xmm3
	addps	%xmm1, %xmm3
	movups	%xmm3, (%rsi)
	addq	$16, %rsi
	decq	%rdx
	jne	LBB0_1
LBB0_2:
	movq	%rbp, %rsp
	popq	%rbp
	retq
                                        ## -- End function
.subsections_via_symbols
