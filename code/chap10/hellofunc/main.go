package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

func main() {
	ctx := context.Background()

	r := rego.New(
		rego.Query(`result = equalsIgnoreCase("Hello", "HELLO")`),
		rego.Function2(
			&rego.Function{
				Name: "equalsIgnoreCase",
				Decl: types.NewFunction(types.Args(types.S, types.S), types.B),
			},
			func(_ rego.BuiltinContext, a, b *ast.Term) (*ast.Term, error) {
				if str1, ok := a.Value.(ast.String); ok {
					if str2, ok := b.Value.(ast.String); ok {

						equals := strings.EqualFold(string(str1), string(str2))
						return ast.BooleanTerm(equals), nil
					}

					return nil, nil
				}

				return nil, nil
			},
		),
	)

	query, err := r.PrepareForEval(ctx)
	if err != nil {
		log.Fatal(err)
	}

	rs, err := query.Eval(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("result=%v\n", rs[0].Bindings["result"])
}