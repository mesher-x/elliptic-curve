#pragma once

typedef unsigned char uchar;
typedef unsigned int u32;
typedef unsigned long u64;
typedef __uint64_t uint64_t;

void fst_kernel(uchar *b2, const u32 *scalar, uint64_t r[5], uint64_t a[5]);
