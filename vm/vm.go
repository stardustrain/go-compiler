package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2048

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object
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

/*
Run - 인출 - 복호화 - 실행 주기가 loop로 동작함
*/
func (vm *VM) Run() error {
	// {"1 + 2"}가 들어왔을 경우
	// fmt.Printf("gerere %s", vm.constants)    // -> [object.Integer(1), object.Integer(2)]
	// fmt.Printf("gerere %s", vm.instructions) // -> 0000 OpConstant 0, 0003 OpConstant 1
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
		case code.OpPop:
			vm.pop()
		case code.OpAdd, code.OpSub, code.OpDiv, code.OpMul:
			err := vm.executeBinaryOperation(op)

			if err != nil {
				return err
			}
		case code.OpTrue, code.OpFalse:
			var monkeyObject = True
			if op == code.OpFalse {
				monkeyObject = False
			}

			err := vm.push(monkeyObject)

			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		fmt.Errorf("unknown Integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeBinaryIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		fmt.Errorf("unkown operator: %d", op)
	}

	return nil
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input == true {
		return True
	}

	return False
}

func (vm *VM) push(o object.Object) error {
	if vm.stackPointer >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.stackPointer] = o
	vm.stackPointer++

	return nil
}

func (vm *VM) pop() object.Object {
	topOfStack := vm.stack[vm.stackPointer-1]
	vm.stackPointer--
	return topOfStack
}

//func (vm *VM) StackTop() object.Object {
//	if vm.stackPointer == 0 {
//		return nil
//	}
//
//	return vm.stack[vm.stackPointer-1]
//}

/*
LastPoppedStackElement
StackTop()과는 다르게 stackPointer를 감소시키지 않고, stackPointer가 가리키는 요소를 반환.
기존의 StackTop()은 단순히 stackPointer에서 1이 감소된 위치(스택의 최상단)의 요소를 반환하기 때문에, stack에 계속 값이 쌓일수 밖에 없음.
OpPop이라는 opcode가 추가되어 -> 표현식이 끝나면 (e.g. 1 + 2) vm.pop() 으로 vm의 스택을 명시적으로 감소시키면서 스택의 최상단 요소가 더 이상 stackPointer - 1이 아니게 됨
즉, 1 -> 2 -> OpAdd -> OpPop이 되어서, OpAdd까지의 연산결과는 3 -> OpPop이 되고, 이때 OpPop(vm.pop())이 실행되면서 stackPointer가 1 감소하게 됨
그러므로, stackPointer는 정확하게 스택의 최상단을 가리키게 됨
*/
func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.stackPointer]
}
