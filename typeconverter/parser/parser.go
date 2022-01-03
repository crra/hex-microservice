package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type Parser struct {
	file *ast.File
	fSet *token.FileSet
}

type MyStructType struct {
	Name   string
	Fields []string
}

type ParseResult struct {
	Package string
	Imports []string
	Structs []MyStructType
}

// New creates a new parser for the given source file.
func New(content string) (*Parser, error) {
	fSet := token.NewFileSet()

	f, err := parser.ParseFile(fSet, "", content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return &Parser{
		file: f,
		fSet: fSet,
	}, nil
}

// Parse parses the given source file.
func (p *Parser) Parse() (ParseResult, error) {
	var res ParseResult

	ast.Inspect(p.file, func(n ast.Node) bool {
		// Visit type declarations / specification
		switch t := n.(type) {
		case *ast.TypeSpec:
			if st, ok := t.Type.(*ast.StructType); ok {
				s := MyStructType{
					Name: t.Name.Name,
				}

				for _, f := range st.Fields.List {
					if len(f.Names) != 1 {
						panic("Only single fields are supported")
					}
					if f.Names[0].IsExported() {
						s.Fields = append(s.Fields, f.Names[0].Name)
					}

					/*
						tags, err := structtag.Parse(string(tag))
						if err != nil {
							panic(err)
						}
					*/
				}

				res.Structs = append(res.Structs, s)
			}
		case *ast.File:
			res.Package = t.Name.Name
		case *ast.ImportSpec:
			res.Imports = append(res.Imports, strings.Trim(t.Path.Value, "\""))
		}

		return true
	})

	return res, nil
}
