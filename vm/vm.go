package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object
	// 언제나 스택의 다음 슬롯을 가리킨다. 따라서 스택의 최 상단은 stackPointer - 1
	// 스택에 들어있는 요소가 하나이고 0번째 인덱스에 위치한다면, 이 값은 1이고 존재하는 요소에 접근하려면 stack[stackPointer - 1]로 접근
	// 새로운 요소 저장 시 stack[stackPointer]에 저장하고 값을 1 증가시킴
	stackPointer int
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		stackPointer: 0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.stackPointer == 0 {
		return nil
	}

	return vm.stack[vm.stackPointer-1]
}

/*
Run - 인출 - 복호화 - 실행 주기가 loop로 동작함
*/
func (vm *VM) Run() error {
	// {"1 + 2"}가 들어왔을 경우
	//fmt.Printf("gerere %s", vm.constants) -> [object.Integer(1), object.Integer(2)]
	//fmt.Printf("gerere %s", vm.instructions) -> 0000 OpConstant 0, 0003 OpConstant 1
	for instructionPointer := 0; instructionPointer < len(vm.instructions); instructionPointer++ {
		op := code.Opcode(vm.instructions[instructionPointer])

		switch op {
		case code.OpConstant:
			// opcode 다음부터 읽음
			constIndex := code.ReadUnit16(vm.instructions[instructionPointer+1:])
			// OpConstant의 operandWidth는 2임
			instructionPointer += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.stackPointer >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.stackPointer] = o
	vm.stackPointer++

	return nil
}
