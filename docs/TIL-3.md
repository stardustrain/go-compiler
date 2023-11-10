# 3일차 TIL
## 지금까지 한 일

- vm의 stack을 비울 수 있게 `OpPop`을 추가함
- vm의 stack에 무한히 push하지 않게 `StackTop()` 대신 `LastPoppedStackElement()`를 추가함

```go
package vm

func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.stackPointer]
}
```

- `OpAdd`, `OpSub`, `OpMul`, `OpDiv`를 추가함
- `OpTrue`, `OpFalse`, `OpEqual`, `OpNotEqual`, `OpGreaterThan`을 추가함
  - 일단은 Integer, boolean object에 대해서만 지원하도록 되어있는데 추후 다른 타입을 지원하는 것도 추가해야할 듯
- 그렇게 어렵지는 않았고, 각 opcode에 대해서 실질적인 코드 동작을 반복 추가하는 작업이었음

## 3일차까지 이해한 것을 정리해 보자면

- `OpPop`
  - 그동안 vm의 stack을 opcode 실행 후 비워주지 않았었음
  - stackPointer만 바꿔서 연산을 하기 때문에 최상단의 element 실행시 해당 요소를 제거하고 stackPointer를 줄여줄 로직을 구현함
  - OpPop은 _표현식_ 의 평가가 끝나면 compiler에서 emit되어 stack에 쌓임
    - `표현식`은 변수의 선언, 1 + 2 같은 연산을 의미함
    - 따라서 예시를 들어보자면 다음과 같음
      - `1 + 2` -> `OpConstant(0)`, `OpConstant(1)`, `OpAdd`, `OpPop`
      - `1;2;` -> `OpConstant(0)`, `OpPop`, `OpConstant(1)`, `OpPop`
- `OpGreaterThan`
  - `>` 연산을 의미함
    - left element를 기준으로, stack에 쌓인 두 요소를 비교함 -> left > right
  - 만약 `<`연산을 해야한다면 새로운 opcode를 추가하는 것이 아니라 비교 순서를 반대로 함
    - left < right를 right > left로 변환함
```go
package compiler

func (c *Compiler) Compile(node ast.Node) error {
  switch node := node.(type) {
  case *ast.InfixExpression:
    if node.Operator == "<" {
      // 왼쪽 오른쪽 피연산자 순서가 바뀌어야 하기 때문에 컴파일 순서 자체를 바꾼다.
      err := c.compileInfixExpressions(node.Right, node.Left)
      if err != nil {
        return err
      }

      c.emit(code.OpGreaterThan)
      return nil
    }

    // 양쪽 left, right를 컴파일
    err := c.compileInfixExpressions(node.Left, node.Right)
    if err != nil {
      return err
    }

    switch node.Operator {
    case ">":
      c.emit(code.OpGreaterThan)
    default:
      return fmt.Errorf("unknown operator: %s", node.Operator)
    }
  }

  return nil
}
```
- `OpTrue`, `OpFalse`
  - ture, false를 표현하는 monkey object는 하나면 충분함
  - 바뀔 일이 없는 값이고 프로그램 전체를 통틀이 공통적으로 사용되는 상수임
  - 그러므로 아래와 같이 monkey object의 포인터 변수를 만들고 사용함
```go
package vm

import ("monkey/object")

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}

func (vm *VM) Run() error {
  for instructionPointer := 0; instructionPointer < len(vm.instructions); instructionPointer++ {
    op := code.Opcode(vm.instructions[instructionPointer])

    switch op {
    case code.OpTrue, code.OpFalse:
      var monkeyObject = True
      if op == code.OpFalse {
        monkeyObject = False
      }

      err := vm.push(monkeyObject)

      if err != nil {
        return err
      }
    }
  }

  return nil
}
```

- `StackTop()`, `LastPoppedStackElement()`
  - 엄청 헷갈리긴 했는데, 간신히 정리함

```go
package vm

func (vm *VM) StackTop() object.Object {
	if vm.stackPointer == 0 {
		return nil
	}

	return vm.stack[vm.stackPointer-1]
}
```
  - 먼저 기존의 StackTop 메서드를 보면 위와 같음
    - `OpPop`이 추가되기 전이었고, stackPointer는 항상 스택의 다름 슬롯을 가리키고 있었기 때문에 단순히 stackPointer에서 1을 감소시킨 인덱스(스택의 최상단)를 가리키게 함
    - 하지만 `OpPop`이 추가되어 표현식이 끝나면 vm.pop()이 실행되어 stackPointer가 자동으로 1 감소하게 됨
  - 그렇기 때문에 stackPointer는 스택의 다음 슬롯이 아닌 최상단을 가리키게 됨
```go
package vm

func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.stackPointer]
}
```
