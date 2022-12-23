#!/bin/sh
g++-11 -m64 -mssse3 -Wno-unused-result -Wno-write-strings -O2 -I. -I/usr/local/cuda-12.0/include -o obj/main.o -c main.cpp
g++-11 obj/main.o -lpthread -o main
