#include <immintrin.h>

typedef struct {
	float x, y, z, w;
} vec4_t;

void matrix_multiply_vec4(__m128* mat4, vec4_t* vec_arr, size_t size);
