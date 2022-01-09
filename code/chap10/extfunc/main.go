package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/cmd"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

func main() {
	methodPermType := types.NewObject(nil, types.NewDynamicProperty(types.S, types.NewArray(nil, types.S)))

	apiType := types.NewObject(nil, types.NewDynamicProperty(types.S, methodPermType))

	rego.RegisterBuiltin3(
		&rego.Function{
			Name:    "longest_match_api",
			Decl:    types.NewFunction(types.Args(apiType, types.S, types.S), types.S),
			Memoize: false,
		},
		longestMatchAPI,
	)

	if err := cmd.RootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func longestMatchAPI(_ rego.BuiltinContext, api, url, method *ast.Term) (*ast.Term, error) {
	if obj, ok := api.Value.(ast.Object); ok {
		targetURL := string(url.Value.(ast.String))

		for targetURL != "" {
			targetAPI := obj.Get(ast.StringTerm(targetURL))

			if targetAPI != nil && targetAPI.Get(method) != nil {
				return ast.StringTerm(targetURL), nil
			}

			targetURL = upperURL(targetURL)
		}
	}

	return nil, nil
}

func upperURL(url string) string {
	last := strings.LastIndex(url, "/")

	if last < 0 {
		return ""
	}

	return url[0:last]
}