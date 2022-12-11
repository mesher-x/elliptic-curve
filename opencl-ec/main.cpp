#include <cmath>
#include <fstream>
#include <iomanip>
#include <iostream>
#include <sstream>
#include <vector>
#include <cstring>
#include <tuple>
#include <cassert>

#if defined(__APPLE__) || defined(__MACOSX)
#include <OpenCL/opencl.hpp>
#else
#include <CL/opencl.hpp>
#endif

typedef unsigned char uchar;
typedef unsigned int u32;

const uchar scalar_size = 8;
const uchar b2_len = 65;

// BigInt.hpp not needed, was used for debugging

// running:
// linux: g++ main.cpp -DCL_HPP_TARGET_OPENCL_VERSION=120 -DCL_HPP_MINIMUM_OPENCL_VERSION=120 -o executable -lOpenCL && ./executable 045769b6ab9a4719181badde312dafe40af6f1ee9676ad413d5d5f08989ea28f690fb1ce68a891f325303618f7c87bc5719570ff559c00be0503033eff68388768 ffff
// mac: clang++ main.cpp -framework OpenCL -std=c++14 -DCL_HPP_TARGET_OPENCL_VERSION=120 -DCL_HPP_MINIMUM_OPENCL_VERSION=120 -o executable && ./executable 1287427058ffff

void fst(int argc, char *argv[]);
void snd(int argc, char *argv[]);

int main(int argc, char *argv[])
{
    fst(argc, argv);
    //snd(argc, argv);
    return 0;
}

void convert_input_hex_to_scalar(const char *hex, u32 *array);
std::tuple<cl::CommandQueue, cl::Context, cl::Program> setup_opencl();
void set_kernel_args(cl::Kernel &k, const std::vector<cl::Buffer*> args);

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

    cl::CommandQueue queue;
    cl::Context context;
    cl::Program program;
    std::tie(queue, context, program) = setup_opencl();

    cl_int ret = -1;

    cl::Buffer scalar_d(context, CL_MEM_READ_ONLY, scalar_size * sizeof(u32));
    ret = queue.enqueueWriteBuffer(scalar_d, CL_TRUE, 0, scalar_size * sizeof(u32), &scalar);
    assert(ret == CL_SUCCESS);

    uchar b2[b2_len];
    cl::Buffer b2_d(context, CL_MEM_WRITE_ONLY, b2_len * sizeof(uchar));

    cl::Buffer r_d(context, CL_MEM_WRITE_ONLY, b2_len * sizeof(uchar));

    cl::Kernel kernel(program, "fst");

    // carefull here, wrong order of arguments may not cause error in runtime
    std::vector<cl::Buffer*> args{&b2_d, &scalar_d, &r_d};
    set_kernel_args(kernel, args);
    ret = queue.enqueueNDRangeKernel(kernel, cl::NullRange, cl::NDRange(1));
    assert(ret == CL_SUCCESS);
    ret = queue.enqueueReadBuffer(b2_d, CL_TRUE, 0, b2_len * sizeof(uchar), &b2);
    assert(ret == CL_SUCCESS);

    u32 r[8];
    ret = queue.enqueueReadBuffer(r_d, CL_TRUE, 0, 8 * sizeof(u32), &r);
    assert(ret == CL_SUCCESS);
    std::cout << r[1] << r[0] << r[3] << r[2] << r[5] << r[4] << r[7] << r[6] << std::endl;
    std::cout << std::endl;

    std::stringstream ss;
    ss << "0x";
    ss << std::hex << std::setfill('0');
    for (int i = 0; i < b2_len; i++)
    {
        ss << std::setw(2) << (u32)b2[i];
    }
    std::cout << "result:" << std::endl << ss.str() << std::endl;
}


void snd(int argc, char* argv[]) {
    if (argc < 3) {
        std::cout << "Please provide 2 hex strings via command line arguments." << std::endl;
        return;
    }

    if (!(argv[1][0] == '0' && argv[1][1] == '4')) {
        std::cout << "first hex string must start with 04" << std::endl;
        return;
    }

    uint l = strlen(argv[1]);
    if (l != 130) {
        std::cout << "first hex string must be 130 characters long, current length=" << l << std::endl;
        return;
    }

    const uchar point_size = 16;
    u32 point[point_size]; // contains x and y, 8 per coordinate
    std::string hex_point(argv[1] + 2);
    for (int i = 0; i < point_size; i++)
    {
        std::string sub = hex_point.substr(i * 8, 8);
        point[i] = std::stoul(sub, nullptr, 16);
    }

    u32 scalar[scalar_size];
    convert_input_hex_to_scalar(argv[2], scalar);

    cl::CommandQueue queue;
    cl::Context context;
    cl::Program program;
    std::tie(queue, context, program) = setup_opencl();

    cl_int ret = -1;

    cl::Buffer scalar_d(context, CL_MEM_READ_ONLY, scalar_size * sizeof(u32));
    ret = queue.enqueueWriteBuffer(scalar_d, CL_TRUE, 0, scalar_size * sizeof(u32), &scalar);
    assert(ret == CL_SUCCESS);

    cl::Buffer point_d(context, CL_MEM_READ_ONLY, point_size * sizeof(u32));
    ret = queue.enqueueWriteBuffer(point_d, CL_TRUE, 0, point_size * sizeof(u32), &point);
    assert(ret == CL_SUCCESS);

    uchar b2[b2_len];
    cl::Buffer b2_d(context, CL_MEM_WRITE_ONLY, b2_len * sizeof(uchar));

    cl::Kernel kernel(program, "snd");
    
    // carefull here, wrong order of arguments may not cause error in runtime
    std::vector<cl::Buffer*> args{&b2_d, &point_d, &scalar_d};
    set_kernel_args(kernel, args);

    ret = queue.enqueueNDRangeKernel(kernel, cl::NullRange, cl::NDRange(1));
    assert(ret == CL_SUCCESS);

    ret = queue.enqueueReadBuffer(b2_d, CL_TRUE, 0, b2_len * sizeof(uchar), &b2);
    assert(ret == CL_SUCCESS);

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

std::tuple<cl::CommandQueue, cl::Context, cl::Program> setup_opencl() {
    std::vector<cl::Platform> platforms;
    cl_int ret = -1;
    ret = cl::Platform::get(&platforms);
    assert(ret == CL_SUCCESS);
    if (platforms.size() == 0) {
        std::cout << "No OpenCL platforms found" << std::endl;
        exit(0);
    }
    cl::Platform platform = platforms[0];
    std::vector<cl::Device> devices;
    ret = platform.getDevices(CL_DEVICE_TYPE_GPU, &devices);
    assert(ret == CL_SUCCESS);
    cl::Device device = devices[0];
    cl::Context context(device);
    std::ifstream kernelFile("kernels.cl");
    std::string kernelSource((std::istreambuf_iterator<char>(kernelFile)), std::istreambuf_iterator<char>());
    bool build = false;
    cl::Program program(context, kernelSource, build, &ret);
    if (ret != CL_SUCCESS) {
        std::cout << "failure to create program, ret=" << ret << std::endl;
        exit(0);
    }
    ret = program.build(device);
    if (ret != CL_SUCCESS) {
        std::cout << "program build failure, ret=" << ret << std::endl;
        exit(0);
    }
    cl::CommandQueue queue(context, device);
    return std::make_tuple(queue, context, program);
}

void set_kernel_args(cl::Kernel& k, const std::vector<cl::Buffer*> args) {
    int ret = -1;
    for (int i = 0; i < args.size(); i++)
    {
        ret = k.setArg(i, *args[i]);
        if (ret != CL_SUCCESS)
        {
            std::cout << "failed to set kernel argument, index=" << i << ", ret=" << ret << std::endl;
            exit(0);
        }
    }
}

//debug
//#include "BigInt.hpp"
// BigInt accum;
// std::cout << "x as array of 32 bytes:";
// int power = 31;
// for (int i = 7; i >= 0; i--)
// {
//     uchar* x_bytes = (uchar*)&x[i];
//     for (int j = 3; j >= 0; j--)
//     {
//         std::cout << (u32)x_bytes[j] << " ";
//         accum += BigInt((u32)x_bytes[j]) * pow(256, power);
//         power -= 1;
//     }
// }
// std::cout << std::endl;
// std::cout << "x as decimal=" << accum << std::endl;