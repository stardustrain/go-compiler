# 2일차 TIL
## 지금까지 한 일
- OpConstant라는 이름의 명령코드 하나를 정의
    - 정수 리터럴을 평가하고, 해당 정수 리터럴을 상수 풀에 추가할 수 있는 기능을 추가
- AST를 순회하는 방법을 알고 있는 작은 컴파일러를 구현
    - 컴파일러에 만들어진 명령코드와 피연산자를 저장함
- 만들어진 명령코드와 피연산자는 vm 인스턴스로 전달되고, vm 인스턴스는 이를 실행함
    - vm 인스턴스는 스택(2048kb)을 가지고 있음
    - vm 인스턴스는 스택에 명령코드와 값을 넣고 빼면서 연산을 수행함

```go
package compiler

func (c *Compiler) Compile(node ast.Node) error {
  switch node := node.(type) {
  case *ast.Program:
    // 모든 node.Statements를 순회하며 c.Compile을 재귀 호출
    for _, statement := range node.Statements {
      err := c.Compile(statement)
      if err != nil {
        return err
      }
    }
  case *ast.ExpressionStatement:
    // 1 + 2를 표현하는 노드
    err := c.Compile(node.Expression)
    if err != nil {
      return err
    }
  case *ast.InfixExpression:
    // 양쪽 left, right를 컴파일
    err := c.Compile(node.Left)
    if err != nil {
      return err
    }

    err = c.Compile(node.Right)
    if err != nil {
      return err
    }

    switch node.Operator {
    case "+":
      c.emit(code.OpAdd)
    }
  case *ast.IntegerLiteral:
    // 리터럴은 상수 표현식이므로, 값이 변하지 않아 *object.Integer를 생성
    integer := &object.Integer{Value: node.Value}
    c.emit(code.OpConstant, c.addConstant(integer))
  }

  return nil
}
```

- Compiler 구조체의 메서드로 Compile을 추가함
    - AST node의 타입에 따라 재귀적으로 동작하며 코드를 방출(emit)함

```go
package compiler

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
  instructions := code.Make(op, operands...)
  position := c.addInstruction(instructions)
  // 지금 만들어낸 명령어의 시작 위치를 반환
  return position
}
```
- emit 메서드를 정의함
    - code.Make 함수를 통해 명령어를 만들고, addInstruction 메서드를 통해 명령어를 저장함
    - 지금 만들어낸 명령어의 시작 위치를 반환함
    - 그런데 어디다 쓰는거지...? 일단 사용처는 아직까지는 없음. 아마 리터럴 타입 변수만 다뤄서 그런것 같음.

```go
package compiler

func (c *Compiler) addInstruction(instructions []byte) int {
  newInstructionPosition := len(c.instructions)
  c.instructions = append(c.instructions, instructions...)
  return newInstructionPosition
}
```
- addInstruction 메서드를 정의함
    - instructions 배열에 새로운 명령어를 추가함
    - 파라미터로 받은 instructions 배열의 길이를 다음 명령어의 시작 위치로 간주하고 해당 값을 반환함

```go
package compiler

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

```
- Bytecode 구조체를 정의함
    - 컴파일러가 코드를 전부 컴파일한 뒤 가지고 있는 명렁어와 상수풀을 구조체의 변수로 선

```go
package vm

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
```
- vm 패키지를 추가함

- StackSize는 2048mb로 제한
- VM 구조체를 정의함
    - stackPointer는 항상 스택의 다음 슬롯을 가리킴. 따라서 stack의 최상단은 stackPointer - 1.
        - StackTop() 메서드를 통해 확인
    - 스택에 들어있는 요소가 하나이고 0번째 인덱스에 위치한다면, 이 값은 1이고 존재하는 요소에 접근하려면 `stack[stackPointer - 1]`로 접근

```go
package vm
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
		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()
			leftValue := left.(*object.Integer).Value
			rightValue := right.(*object.Integer).Value

			result := leftValue + rightValue
			vm.push(&object.Integer{Value: result})
		}
	}

	return nil
}
```

- Run 매서드를 정의함
    - Stack에 쌓인 명령코드를 실질적으로 동작시킴
    - 명령코드의 길이만큼 순회하며(`len(vm.instructions)`) opcode를 확인하고, 해당 opcode에 맞는 동작을 분기처리해 수행
    - OpConstant는 상수를 스택에 넣는 동작 수행
        - instructions 배열에서 opcode 다음부터 값을 읽어들임
        - 그 후 등록된 상수 풀에서 값을 꺼내어 stack에 집어 넣음
    - OpAdd는 스택에 쌓인 두 값을 꺼내어 더한 뒤 다시 스택에 넣음

```go
package vm

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
```

- push, pop을 정의함
    - push는 최상단 스택의 다음 슬롯에 값을 넣고 stackPointer를 1 증가 시킴
    - pop은 최상단 스택의 값을 꺼내고 stackPointer를 1 감소 시킴

## 2일차까지 이해한 것을 정리해 보자면

- 기본적인 흐름은 다음과 같다.
  - vm에서 실행할 bytecode를 만든다. (code 패키지)
    - 앞의 1byte는 op(연산자) 뒤는 operand(피연산자)로 구성된다.
    - 피연산자의 길이는 opcode와 같이 정해져 있으며, 피연산자를 받지 않을수도 있다.
  - 컴파일러는 AST를 순회하며 bytecode를 만든다. (compiler 패키지)
  - vm은 compiler 패키지에서 만들어진 상수풀과 명령코드를 전달받아 stack을 이용해 동작한다. (vm 패키지)

- 많이 어렵긴한데, 개념이 어려운거지 아직 golang에서 어려운 부분은 크게 없어서 다행이다.