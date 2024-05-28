#include <stdio.h>
#include <immintrin.h>

#include "matrix_amd64.h"

int main(void) {
	__m128 mat4[4] = {
		_mm_set_ps(1.0, 0.0, 0.0, 0.0),  // Row 1
		_mm_set_ps(0.0, 2.0, 0.0, 0.0),  // Row 2
		_mm_set_ps(0.0, 0.0, 3.0, 0.0),  // Row 3
		_mm_set_ps(0.0, 0.0, 0.0, 4.0)   // Row 4
	};

	printf("%.0f %.0f %.0f %.0f\n", mat4[0][0], mat4[0][1], mat4[0][2], mat4[0][3]);
	printf("%.0f %.0f %.0f %.0f\n", mat4[1][0], mat4[1][1], mat4[1][2], mat4[1][3]);
	printf("%.0f %.0f %.0f %.0f\n", mat4[2][0], mat4[2][1], mat4[2][2], mat4[2][3]);
	printf("%.0f %.0f %.0f %.0f\n", mat4[3][0], mat4[3][1], mat4[3][2], mat4[3][3]);
	printf("\n");

//	float mat[16] = {
//		1, 0, 0, 1,
//		0, 1, 0, 0,
//		0, 0, 1, 0,
//		0, 0, 0, 1,
//	};
//
//	__m128 *mat4 = (__m128*) mat;

	vec4_t vec_arr[] = {
		{1, 1, 1, 1},
	};

	matrix_multiply_vec4(mat4, vec_arr, 1);
	printf("%f %f %f %f\n", vec_arr[0].x, vec_arr[0].y, vec_arr[0].z, vec_arr[0].w);

	return 0;
}
