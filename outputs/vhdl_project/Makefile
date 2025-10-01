# Project Makefile for GHDL + GTKWave
# space separated list of source files
SRC       := src/main.vhd
TB_DIR    := tb
WAVE_DIR  := waves
GHDL_STD  := --std=08

# Default simulation time
SIM_TIME  := 500ns

# Default top-level (can be overridden)
TOP       ?= main_tb

# Phony targets
.PHONY: all analyze elab run wave clean

all: run wave

# Analyze (compile) all VHDL files
analyze:
	ghdl -a $(GHDL_STD) --work=work $(SRC) 
	ghdl -a $(GHDL_STD) --work=work $(TB_DIR)/*.vhd

# Elaborate top-level testbench
elab: analyze
	ghdl -e $(GHDL_STD) --work=work $(TOP)

# Run simulation and generate waveform
run: elab
	mkdir -p $(WAVE_DIR)
	ghdl -r $(GHDL_STD) --work=work $(TOP) --stop-time=$(SIM_TIME) --wave=$(WAVE_DIR)/$(TOP).ghw

# Launch GTKWave
wave:
	@if [ -f $(WAVE_DIR)/$(TOP).gtkw ]; then \
		gtkwave $(WAVE_DIR)/$(TOP).ghw $(WAVE_DIR)/$(TOP).gtkw & \
	else \
		gtkwave $(WAVE_DIR)/$(TOP).ghw & \
	fi

# Clean all generated files
clean:
	rm -f *.o *.cf $(TB_DIR)/*.o $(TB_DIR)/*.cf $(TOP) $(WAVE_DIR)/*.ghw
# run "make clean" to remove all generated files
# run "make TOP=entityName" to compile, elaborate, run and view waveforms
