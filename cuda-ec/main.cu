#include "kernels.cuh"

#include <cassert>
#include <cmath>
#include <cstring>
#include <cuda_runtime.h>
#include <fstream>
#include <iomanip>
#include <iostream>
#include <sstream>
#include <tuple>
#include <vector>

const uchar scalar_size = 8;
const uchar b2_len = 65;

void fst(int argc, char *argv[]);
//void snd(int argc, char *argv[]);

int main(int argc, char *argv[])
{
    fst(argc, argv);
    //snd(argc, argv);
    return 0;
}

void convert_input_hex_to_scalar(const char *hex, u32 *array);

void fst(int argc, char *argv[])
{
    if (argc < 2) {
        std::cout << "Please provide a hex string as a command line argument." << std::endl;
        return;
    }

    if (argc > 2) {
        std::cout << "command line argument must be single" << std::endl;
        return;
    }

    if (strlen(argv[1]) > 64) {
        std::cout << "hex string must be 64 chars or shorter" << std::endl;
        return;
    }

    u32 scalar[scalar_size];
    convert_input_hex_to_scalar(argv[1], scalar);

    u32* scalar_d;
    cudaError_t e;
    e = cudaMalloc((void**)&scalar_d, scalar_size * sizeof(u32));
    assert(e == cudaSuccess);
    e = cudaMemcpy(scalar_d, scalar, scalar_size * sizeof(u32), cudaMemcpyHostToDevice);
    assert(e == cudaSuccess);

    uchar* b2_d;
    e = cudaMalloc((void**)&b2_d, b2_len * sizeof(uchar));
    assert(e == cudaSuccess);
    
    uint64_t* r_d;
    e = cudaMalloc((void**)&r_d, 5 * sizeof(uint64_t));
    assert(e == cudaSuccess);

    uint64_t* a_d;
    e = cudaMalloc((void**)&a_d, 5 * sizeof(uint64_t));
    assert(e == cudaSuccess);

    fst_kernel(b2_d, scalar_d, r_d, a_d);

    uchar b2[b2_len];
    e = cudaMemcpy(b2, b2_d, b2_len * sizeof(uchar), cudaMemcpyDeviceToHost);
    assert(e == cudaSuccess);

    uint64_t r[5];
    e = cudaMemcpy(r, r_d, 5 * sizeof(uint64_t), cudaMemcpyDeviceToHost);
    assert(e == cudaSuccess);

    for (int i = 0; i < 5; i++)
    {
        std::cout << ((r[i] >> 32) & 0xFFFFFFFF) << (r[i] & 0xFFFFFFFF);
    }
    std::cout << std::endl;

    e = cudaFree(scalar_d);
    assert(e == cudaSuccess);
    e = cudaFree(b2_d);
    assert(e == cudaSuccess);
    e = cudaFree(r_d);
    assert(e == cudaSuccess);
    e = cudaFree(a_d);
    assert(e == cudaSuccess);
    
    std::stringstream ss;
    ss << "0x";
    ss << std::hex << std::setfill('0');
    for (int i = 0; i < b2_len; i++)
    {
        ss << std::setw(2) << (u32)b2[i];
    }
    std::cout << "result:" << std::endl << ss.str() << std::endl;
}

void convert_input_hex_to_scalar(const char *hex, u32 *array) {
    std::string hex_str(hex);
    assert(hex_str.length() <= 64);
    std::string paddedHex = std::string(64 - hex_str.length(), '0') + hex_str;

    for (int i = 0; i < 8; i++)
    {
        std::string sub = paddedHex.substr(i * 8, 8);
        array[7 - i] = std::stoul(sub, nullptr, 16);
    }
}

// int main()
// {
//     int a = 3, b = 4, c;
//     int* d_c;

//     // Allocate memory on the device
//     cudaMalloc((void**)&d_c, sizeof(int));

//     // Launch the kernel on the device
//     add_kernel(a, b, d_c);

//     // Copy the result back to the host
//     cudaMemcpy(&c, d_c, sizeof(int), cudaMemcpyDeviceToHost);

//     // Free the device memory
//     cudaFree(d_c);

//     // Print the result
//     std::cout << "The sum of " << a << " and " << b << " is " << c << std::endl;

//     return 0;
// }
