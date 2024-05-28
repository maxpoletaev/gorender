#include <immintrin.h>
#include "matrix_amd64.h"

void matrix_multiply_vec4(__m128* mat4, vec4_t* vec_arr, size_t size) {
    for (int i = 0; i < size; ++i) {
        __m128 x = _mm_set1_ps(vec_arr[i].x);
        __m128 y = _mm_set1_ps(vec_arr[i].y);
        __m128 z = _mm_set1_ps(vec_arr[i].z);
        __m128 w = _mm_set1_ps(vec_arr[i].w);

        __m128 p1 = _mm_mul_ps(x, mat4[0]);
        __m128 p2 = _mm_mul_ps(y, mat4[1]);
        __m128 p3 = _mm_mul_ps(z, mat4[2]);
        __m128 p4 = _mm_mul_ps(w, mat4[3]);

        __m128 result = _mm_add_ps(_mm_add_ps(p1, p2), _mm_add_ps(p3, p4));

        _mm_storeu_ps((float*)&vec_arr[i], result);
    }
}