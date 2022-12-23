SRC = main.cu kernels.o

OBJDIR = obj

OBJET = $(addprefix $(OBJDIR)/, kernels.o)

CXX        = g++
CUDA       = /usr/local/cuda-12.0
CXXCUDA    = /usr/bin/g++-11
NVCC       = $(CUDA)/bin/nvcc

ifdef dbg
CXXFLAGS   = -DWITHGPU -m64 -mssse3 -Wno-unused-result -Wno-write-strings -g -I. -I$(CUDA)/include
else
CXXFLAGS   = -DWITHGPU -m64 -mssse3 -Wno-unused-result -Wno-write-strings -O2 -I. -I$(CUDA)/include
endif
LFLAGS     = -lpthread -L$(CUDA)/lib64 -lcudart

ifdef dbg
$(OBJDIR)/kernels.o: kernels.cu
	$(NVCC) -G -maxrregcount=0 --ptxas-options=-v --compile --compiler-options -fPIC -ccbin $(CXXCUDA) -m64 -g -I$(CUDA)/include -gencode=arch=compute_$(ccap),code=sm_$(ccap) -o $(OBJDIR)/kernels.o -c kernels.cu
else
$(OBJDIR)/kernels.o: kernels.cu
	$(NVCC) -maxrregcount=0 --ptxas-options=-v --compile --compiler-options -fPIC -ccbin $(CXXCUDA) -m64 -O2 -I$(CUDA)/include -gencode=arch=compute_$(ccap),code=sm_$(ccap) -o $(OBJDIR)/kernels.o -c kernels.cu
endif

# $(OBJDIR)/%.o : %.cpp
# 	$(CXX) $(CXXFLAGS) -o $@ -c $<

main: $(OBJET)
	@echo Making executable...
	$(NVCC) $(OBJET) $(LFLAGS) main.cu -o executable

$(OBJET): | $(OBJDIR)

$(OBJDIR):
	mkdir -p $(OBJDIR)

clean:
	@echo Cleaning...
	@rm -f obj/*.o
