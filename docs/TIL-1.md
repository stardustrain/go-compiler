# 1일차 TIL
## 1. 프로그램의 동작?
- 컴퓨터가 이해할 수 있는 명령어: instructions
- 인터프리터와 컴파일러 모두 frontend를 갖는다.
    - frontend에서는 소스 언어로 작성된 코드를 읽어들여 특정 데이터 구조로 변환한다.
    - lexer(소스 코드 토큰화), parser(토큰 파싱)는 함께 AST를 만든다.
    - AST를 순회하면서
        - 인터프리터는 AST에 부호화된 명령을 실행한다.
        - 컴파일러는 목적 언어(target language)로 된 소스코드를 만든다.
- Frontend 다음에는 옵티마이저가 AST를 또 다른 내부 표현(internal representation; IR)으로 변환한다.
- 컴퓨터 메모리는 `워드`라는 단위로 구분함.
- CPU에게 특정 위치의 워드에 있는 데이터를 인출하기를 요청한다면, CPU는 그 워드의 값을 인출한다.
    - 결국 pointer를 만드는 것
- 워드 어드레싱이 아닌 byte-addressing 방식을 사용하기도 한다.

## 2. 가상 머신이란?
- 가상 머신은 소프트웨어로 구현된 컴퓨터
- 가상 머신은 bytecode를 실행한다.
    - 바이트 코드라 부르는 이유는, push, pop, add, call 같은 명령코드(opcode)의 크기가 1바이트이기 때문
- 명령 코드(opcode)
    - 연산자의 역할을 함
    - 1 바이트로 구현됨
    - 사람이 쉽게 알아볼 수 있어야 하기 때문에 보통 PUSH, POP 같은 상수를 명령코드와 mapping함.
- 피연산자(operands)
    - 명령 코드와 바이트코드에 나란히 포함됨.
    - 피연산자는 반드시 1바이트일 필요는 없음.
- Endian
    - 피연산자와 연산자는 부호화됨.
    - big endian 방식은 데이터의 최상위 비트가 첫 번째에 오도록 함.
    - little endian 방식은 데이터의 최하위 비트가 첫 번째에 오도록 함.

## 3. 코드에서 배운 것
### code.go

```go
package code

type Instructions []byte
type Opcode byte
```

- 컴퓨터가 이해할 수 있는 명령어의 집합인 []byte 타입의 Instructions 타입 정의.
- 연산자 타입을 byte 타입의 Opcode로 정의

```go
package code

const (
	OpConstant Opcode = iota
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definition = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definition[Opcode(op)]

	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}
```

- Opcode 타입의 OpConstant 상수를 정의하고 iota를 사용해 0부터 순차적으로 증가하도록 함.
    - `iota`는 const 키워드로 선언된 블록 내부에서 0부터 1씩 값이 증가하도록 만들어 줌
- Definition 타입을 정의함
    - Name: 연산자의 이름
    - OperandWidths: 피연산자의 크기
- definition이라는 map을 정의함
    - Opcode를 key로하고, Definition 타입의 구조체를 value로 함
- Lookup 함수를 정의함
    - byte 타입의 op를 인자로 받아 이를 Opcode 타입으로 캐스팅 함
    - definition map에서 Opcode 타입의 op를 key로 하여 value를 가져옴

```go
package code

func Make(op Opcode, operands ...int) []byte {
	def, ok := definition[op]

	if !ok {
		return []byte{}
	}

	instructionLen := 1

	for _, width := range def.OperandWidths {
		instructionLen += width
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, operand := range operands {
		width := def.OperandWidths[i]

		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(operand))
		}

		offset += width
	}

	return instruction
}
```

- Make 함수를 정의함
    - Opcode - operands 쌍을 만드는 함수

- 연산자와 피연산자를 인자로 받음
- 연산자가 있다면 연산자의 이름과 피연산자의 크기를 가져옴
    - 이때 피연산자의 배열 길이의 기본값을 1로 설정한 상태에서 각 Definition 타입의 OperandWidths를 순회하며 최종적인 피연산자의 크기를 더해 instructionLen을 구함
- 피연산자 배열의 크기로 컴퓨터가 이해할 수 있는 명령어의 집함 instruction 배열을 미리 선언
    - instruction 배열의 첫 번째 원소에 연산자를 넣음
    - 피연산자 배열을 순회하며 바이트코드로 만들어 instruction 배열에 넣음
    - 그 후 차지한 길이만큼 offset을 설정해 다음 공간에 다음 피연산자의 바이트 코드를 넣음

```go
package code

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUnit16(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUnit16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

```

- ReadOperands 함수를 정의함
    - 명령어의 집합 instructions 배열을 받아서 피연산자의 byte 크기 만큼의 바이트 코드를 읽음
    - 이때, 바이트 코드를 다시 적당한 타입의 값으로 변환해서 (e.g. BigEndian -> Unit16) 배열에 넣어줌