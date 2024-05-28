ASM_CFLAGS = -Wall -O3 -S -mno-red-zone -mstackrealign -fno-asynchronous-unwind-tables -fno-exceptions -fno-rtti

.PHONY: all
all: matrix_amd64.s

matrix_amd64.s:
	clang $(ASM_CFLAGS) -o $@ matrix_amd64.c

main:
	clang -o $@ -Wall -O3 *.c

.PHONY: clean
clean:
	rm -f *.s
	rm -f main
