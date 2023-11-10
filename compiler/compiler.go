package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
)

type Compiler struct {
	// 생성한 바이트코드 담기
	instructions code.Instructions
	// constants pool 역할
	constants []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

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
		c.emit(code.OpPop)
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
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case ">":
			c.emit(code.OpGreaterThan)
		default:
			return fmt.Errorf("unknown operator: %s", node.Operator)
		}
	case *ast.IntegerLiteral:
		// 리터럴은 상수 표현식이므로, 값이 변하지 않아 *object.Integer를 생성
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.Boolean:
		if node.Value == true {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	}

	return nil
}

/*
addConstant - OpConstant 명령어가 사용할 피연산자.
가상 머신에게 이 상수를 상수 풀에서 가져와 콜 스택에 집어 넣게 만드는 역할
*/
func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

/*
emit - 명령어를 만들고(instruction) 명령어 배열에 해당 명령어를 추가한 다음, 해당 명령어의 위치를 반환
*/
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	instructions := code.Make(op, operands...)
	position := c.addInstruction(instructions)
	// 지금 만들어낸 명령어의 시작 위치를 반환
	return position
}

func (c *Compiler) addInstruction(instructions []byte) int {
	newInstructionPosition := len(c.instructions)
	c.instructions = append(c.instructions, instructions...)
	return newInstructionPosition
}

/*
Bytecode - 컴파일러가 만들어낸 Instructions와 컴파일러가 평가한 Constants를 담는다.
나중에 가상머신에 전달할 대상이 된다.
*/
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

func (c *Compiler) compileInfixExpressions(nodes ...ast.Node) error {
	for _, node := range nodes {
		err := c.Compile(node)
		if err != nil {
			return err
		}
	}
	return nil
}
