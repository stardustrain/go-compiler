package compiler

import (
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
	return nil
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
