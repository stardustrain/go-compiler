package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{
			OpConstant,
			[]int{65534},
			[]byte{byte(OpConstant), 255, 254},
		},
		{
			OpAdd,
			[]int{},
			[]byte{byte(OpAdd)},
		},
	}

	for _, tt := range tests {
		instructions := Make(tt.op, tt.operands...)

		if len(instructions) != len(tt.expected) {
			t.Errorf("Instructions has wrong lengt")
		}

		for i, b := range tt.expected {
			if instructions[i] != tt.expected[i] {
				t.Errorf("wrong byte at pos %d. want=%d, god=%d", i, b, instructions[i])
			}
		}
	}
}

func TestInstructionString(t *testing.T) {
	instructions := []Instructions{
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpConstant 1
0003 OpConstant 2
0006 OpConstant 65535
`
	concatted := Instructions{}

	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted. \nwant=%q\ngot=%q", expected, concatted.String())
	}

	instructions = []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}
	expected = `0000 OpAdd
0001 OpConstant 2
0004 OpConstant 65535
`
}

func TestOpAddInstructionString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}
	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpConstant 65535
`
	concatted := Instructions{}

	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted. \nwant=%q\ngot=%q", expected, concatted.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
	}

	for _, tt := range tests {
		instructions := Make(tt.op, tt.operands...)

		operand, err := Lookup(byte(tt.op))

		if err != nil {
			t.Fatalf("operand not defined: %q\n", err)
		}

		// 복호화된 피 연산자를 반환하고, 복호화된 피연산자를 읽을 때 바이트를 얼마만큼의 크기로 읽어야 하는지 알려줌.
		operandsRead, n := ReadOperands(operand, instructions[1:])

		if n != tt.bytesRead {
			t.Fatalf("n wrong. want=%d, got=%d", tt.bytesRead, n)
		}

		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
			}
		}
	}
}
