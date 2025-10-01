library ieee;
use ieee.std_logic_1164.all;

entity fa_tb is
end fa_tb;

architecture TB_ARCHITECTURE of fa_tb is

	-- Component declaration of the tested unit
	component fa
		port(
			A    : in  STD_LOGIC;
			B    : in  STD_LOGIC;
			Cin  : in  STD_LOGIC;
			Sum  : out STD_LOGIC;
			Cout : out STD_LOGIC
		);
	end component;

	-- Stimulus signals
	signal A    : STD_LOGIC := '0';
	signal B    : STD_LOGIC := '0';
	signal Cin  : STD_LOGIC := '0';
	signal Sum  : STD_LOGIC;
	signal Cout : STD_LOGIC;

begin

	-- Unit Under Test port map
	UUT : fa
		port map (
			A    => A,
			B    => B,
			Cin  => Cin,
			Sum  => Sum,
			Cout => Cout
		);

	-- Stimulus process
	stim_proc: process
	begin
		-- Try all 8 combinations
		A <= '0'; B <= '0'; Cin <= '0'; wait for 10 ns;
		A <= '0'; B <= '0'; Cin <= '1'; wait for 10 ns;
		A <= '0'; B <= '1'; Cin <= '0'; wait for 10 ns;
		A <= '0'; B <= '1'; Cin <= '1'; wait for 10 ns;
		A <= '1'; B <= '0'; Cin <= '0'; wait for 10 ns;
		A <= '1'; B <= '0'; Cin <= '1'; wait for 10 ns;
		A <= '1'; B <= '1'; Cin <= '0'; wait for 10 ns;
		A <= '1'; B <= '1'; Cin <= '1'; wait for 10 ns;

		-- Stop simulation
		wait;
	end process;

end TB_ARCHITECTURE;

-- Configuration
configuration TESTBENCH_FOR_fa of fa_tb is
	for TB_ARCHITECTURE
		for UUT : fa
			use entity work.fa(behavioral);  
		end for;
	end for;
end TESTBENCH_FOR_fa;
