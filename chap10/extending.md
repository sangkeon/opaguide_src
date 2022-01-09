# Chap 10. Extending OPA

## Contents
- OPA built-in function.
- Example of OPA built-in function
- OPA plug-in.
- Setting up OPA version information
 
OPA can be expanded through embedded functions or plug-in implementations. OPA does not have a plug-in system that dynamically loads modules like programs implemented in Java or other languages, and uses a method of creating a new binary by expanding existing OPA implementations. This seems inevitable because the dynamic module system of the Go language itself is still constrained. However, although there is an inconvenience in building a binary, it can be expanded while maintaining readability and independence of the code by writing a new main function using the existing OPA code itself as if it were a library. Chapter 10 first examines how to create a new binary by adding built-in functions and plug-ins to the existing OPA binary. In addition, since the newly created binary has an added function compared to the existing OPA, version information must be filled out to express this. Therefore, a method of setting version information of a binary to prevent user confusion and enable shape management will also be described.

First, access the OPA’s  Github repository (https://github.com/open-policy-agent/opa)) and open the main.go file of the top directory. If the license and annotation part for the go:generate is omitted, the main contents are as follows.

```
// Omit the license part.
 
package main
 
import (
    "fmt"
    "os"
 
    "github.com/open-policy-agent/opa/cmd"
)
 
func main() {
    if err := cmd.RootCommand.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
 
// Omit the go:generate part
```

All you do in the main function is run cmd.Root Command.Execute() with error handling. The only task required to expand OPA is to create and register built-in functions or plug-ins before executing cmd.Root Command.Execute().
 
## OPA built-in function.
### Implementation of OPA built-in function.
Let's look at the implementation of the OPA built-in function in an example. If you use the Go client library, you can run queries by adding built-in functions without building a new OPA binary. First, a project for example is created as follows.

```
$ mkdir hellofunc
$ cd hellofunc
$ go mod init hellofunc
```

When the project creation is completed, the main.go is created and stored with the following contents.

```
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
```

Brief description of the contents of the code is a program that performs and outputs a query called result=equalsIgnoreCase ("Hello", "HELLO"), which created an equivalentIgnoreCase function that compares strings regardless of case and included it at the time of initialization. String comparison is a function that uses the strings.EqualFold function provided by the basic library of the Go language to compare whether the strings are the same for the UTF-8 string without case discrimination.

The parts related to the built-in function are as follows. First, create an option to call the rego.function2 function and hand it over to rego.New as an argument. rego.New takes over the list of func (*Rego) type functions as arguments, and reflects options by sequentially applying the functions taken over as arguments to the Rego data structure. This is a common code pattern in the Go language. In addition to rego.Function2, rego.Function1, rego.Function3, and rego.Function4 also exist, and 1, 2, 3, and 4 each represent the number of function arguments. When more arguments are needed, rego.FunctionDyn, which takes over the arguments in a dynamic array, can be used.
 
If you run the function as follows, you can see the expected results, and if it is the first time to run, additional contents for installing related go modules will be output. rego. FunctionXXX receives two arguments, the first argument being a pointer to the structure containing the declaration of the function, and the second being a function that implements the built-in function and is written as an anonymous function in the code. The declaration part contains the name of the function and the Decl field contains the type information of the function. In the above example, types.S and types.B represent String and Boolean types, respectively. Therefore, the EqualsIncoreCase function receives two string arguments and returns the bullion value.

Looking at the anonymous function part, the anonymous function has one rego.BuiltinContext type argument and two *ast.Term arguments. _ means that the corresponding argument  exists to match only the type without reference within the function, *ast.Term type arguments  were named a and b, respectively. Also, the return value of the arguments are *ast.Term and error type, and all arguments of the built-in function are also *ast.Term type. The Term structure includes a Value containing values and a Location field containing location (file name, line, etc.) information in the Rego source code. if str1, ok := a.Value (ast.String);ok {} part is described as follows. After casting  a.Value as an ast.String type, if successful, the value is assigned to the str1 variable, ok = true, and the content of the block is executed. If casting is impossible, it becomes ok = false and the content of the block is not executed. In the nested if statement, Value field of a and b of *ast.Term types were cast as ast.String types, respectively, and converted into strings and handed over to the EqualFold function of the Go language. Ast.String is actually only an alias of a string type, but since it is recognized as a different type when compiling, string() conversion is required. Finally, the Boolean value returned as a result of the comparison was converted into ast.BooleanTerm and returned. As explained when explaining the use of OPA through the Go client library, the subsequent part is the part that executes queries, receives results, and outputs them.

Now that we've finished describing the code, let's run it. If you run it as follows, you can see the expected result value.

``` 
$ go run main.go
result=true
```

### Integration of OPA built-in functions
Now that we have learned how to implement the built-in function, let's implement the existing rule as a built-in function and run the unit test.

In the final API authorization scenario, including the API hierarchy described in Chapter 5, if API URLs are not matched, the function of matching higher URLs to verify permissions was implemented as a longestmatchapi rule. Due to the nature of the Rego language used by OPA, the longestmatchapi rule found all the matching top URLs without interruption on the way, and returned the first URLs sorted by url’s length in descending order. Since the built-in function is implemented in the Go language, you can return immediately without finding all higher URLs if you find a matching URL.

First, let's create a directory that will act as a project under the name extfunc. After creation, go to the extfunc directory and execute the following command.

``` 
$ go mod init extfunc
go: creating new go.mod: module extfunc
 
The Go module has been created and the created go.mod file is as follows. 

module extfunc
 
go 1.15
```

Currently, only the module name and Go version are specified, but if you add a library later, information related to dependence will also be added. If the Go version is 1.13 or higher, there will be no big problem.

Write the following and save it as main.go.

``` 
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
            Name: "longest_match_api",
            Decl: types.NewFunction(types.Args(apiType, types.S, types.S), types.S),
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
```

Let's look at the main function first. As explained earlier in Chapter 10, the main function registers the built-in function before executing cmd.Root Command.Execute(). To register a built-in function as a rego.RegisterBuiltin3 function, a rego.Function structure containing a function declaration and a function that implements the built-in function are used as arguments. Number 3 of rego.RegisterBuiltin3 represents the number of arguments similar to the previous example, and the rego.Function structure is generated inside rego.RegisterBuiltin3. Compared to the previous example, the argument type of the function is much more complex, and the first argument is the type for referring to the data.api object. The contents of the data.api object are as follows.

```
{
  "api": {
    "/users/{user_id}/profile" : {
      "GET": ["profile.read"],
      "PUT": ["profile.read", "profile.update"]
    },
    ...
  },
...
}
```

The api object is an object with a URL as a key and a value as an object containing authorization information. The key to the authorization object is an HTTP method, and the value is an array of strings containing permissions. The type of value of the api object was generated by types.NewObject (nil, types.NewDynamicProperty (types.S, types.NewArray (nil, types.S))). The first nil part of type.NewObject or type.NewArray is a list of static attributes. A static attribute refers to an essential attribute that an object must contain, for example, a field with a specific key, and in this example, there may be no key of the object and there is no static attribute because there is no essential attribute. The dynamic attribute is a key string type and the value is a string array. An object type having a dynamic attribute in which the string is the key type and the method PermType is the value type was defined as apiType.
 
The method of defining the type is not documented in detail in the official document, but readers who want to check the actual content can check the type/type.go file of the OPA source code. Type.go defines Null, String, Number, Boolean, Any, Function, Set, Object, Array, etc. When the actual code is checked, type S is also generated as var S = type.NewString(). Since it is a simple type, it refers to a value generated once without generating it every time. Types.N and types.B have similar patterns, such as generated by types.NewNumber() and types.NewBoolean(), respectively.

Looking at the function declaration part, the name of the function is longest_api_match, and the type of arguments are data.api to receive the object, string for URL, and string for HTTP method, and the function returns the matched URL string. Memoize is a field that caches the results for functions and if the arguments are the same, the result returned from the cache. Memoize is set to false, the results are not cached. API permission data can be updated by bundle polling, etc., and the results can vary depending on the time of execution due to external effects, so it was set to false. In the opposite case, if we run a function called strlen ("hello") that counts the length of a string, we will always get 5, and we will not have a problem if we cache 5 for the argument "hello". With this deterministic function, it is possible to improve performance through caching by setting Memoize to true.
 
Looking at the implementation of the function, as in the previous example, context information was not used, so it was referred as _, and three arguments were declared as *ast.Term type. Since api is a reference to an object, it was cast as ast.Object, and the Get function was called with the url as key, and if there was an item for that url, and the Get was called again with method as key. If no value is found, the upperURL function is called and repeated to change url to one level up.

Since we have completed implementing the built-in function, let's compile the main.go and create a new OPA binary. The following command allows you to compile main.go and create a binary called opaex, which in a window environment requires an extension for execution, so let's specify opaex.exe instead of opaex.

``` 
$ go build -o opaex main.go
```

If you run REPL for opa and newly built opaex as follows and invoke the longest_match_api function, you can see that the result is output as undefined, not the undefined function error in opaex.

```
$ ./opaex run
OPA 0.26.0 (commit , built at )
 
Run 'help' to see a list of commands and check for updates.
 
> longest_match_api({},"","")
undefined
 
$ ./opa run
OPA 0.26.0 (commit 62d3900, built at 2021-01-20T18:56:12Z)
 
Run 'help' to see a list of commands and check for updates.
> longest_match_api({},"","")
 
1 error occurred: 1:1: rego_type_error: undefined function longest_match_api
```

This time, let's modify the rego file containing the existing authorization scenario to use the newly added longest_match_api built-in function. Let's modify the content as follows and save it as a new file policy_extfunc.rego.

``` 
package example.urlhierarchy
 
default allowed = false
 
allowed {
  input.user.role == "ADMIN"
}
 
allowed {
  input.user.role == "USER"
 
  input.user.target_user_id != ""
  input.user.target_user_id == input.user.id
}
 
allowed {
  input.user.role == "OPERATOR"
 
  method := input.api.method
  url := longest_match_api(data.api, input.api.url, method)
 
  permission := data.operator_permission[input.user.id]
  required_permssion := data.api[url][method]
  satisfied_permission := {p | permissionmatch(permission[_], required_permssion[p], ".")}
 
  count(required_permssion) > 0
  count(satisfied_permission) == count(required_permssion)
}
 
allowed {
  method := input.api.method
  url := longest_match_api(data.api, input.api.url, method)
  url
 
  required_permssion := data.api[url][method]
 
  count(required_permssion) == 0
}
 
permissionmatch(permission, req_permission, delim) = true {
  permission == req_permission
} else = result {
  result := startswith(req_permission, concat("", [permission, delim]))
}
```

If you check the changes, you can see that the part where the rule longestmatchapi (url, method) was declared has been completely removed and the rule call has been changed to a built-in function call. Unlike the rules, the built-in function cannot directly access the data, so it can also be confirmed that the data.api object was passed as the first argument of the function to refer to API authority.

Let's run the unit test using the modified file. The test scenario used the same one written in Chapter 5. If you change only the rego file to the newly created policy_extfunc.rego and run it as it is with opa, an error occurs that the function cannot be found as follows.

```
$ opa test -v policy_extfunc.rego policy_test.rego data.json
2 errors occurred:
policy_extfunc.rego:21: rego_type_error: undefined function longest_match_api
policy_extfunc.rego:34: rego_type_error: undefined function longest_match_api

Let's run the same command by adding built-in functions and turning them into built opaex binaries. It can be confirmed that the unit test was successfully executed.
 
$ opaex test -v policy_extfunc.rego policy_test.rego data.json
data.example.urlhierarchy.test_admin_allowed: PASS (12.3149ms)
data.example.urlhierarchy.test_oper1_profile_read_allowed: PASS (993µs)
data.example.urlhierarchy.test_oper1_profile_update_allowed: PASS (1.0001ms)
data.example.urlhierarchy.test_oper2_profile_read_allowed: PASS (999.7µs)
data.example.urlhierarchy.test_oper2_profile_update_not_allowed: PASS (522.4µs)
data.example.urlhierarchy.test_oper3_profile_read_not_allowed: PASS (1.0086ms)
data.example.urlhierarchy.test_oper3_profile_update_not_allowed: PASS (2.9998ms)
data.example.urlhierarchy.test_user_allowed: PASS (0s)
data.example.urlhierarchy.test_user_not_allowed: PASS (1.0033ms)
data.example.urlhierarchy.test_user_public: PASS (997.8µs)
data.example.urlhierarchy.test_useroper_userlist_read_allowed: PASS (0s)
data.example.urlhierarchy.test_useroper_user_read_allowed: PASS (2.0002ms)
data.example.urlhierarchy.test_useroper_user_update_not_allowed: PASS (1.0009ms)
data.example.urlhierarchy.test_oper1_user_read_not_allowed: PASS (0s)
data.example.urlhierarchy.test_goods_not_allowed: PASS (998.9µs)
--------------------------------------------------------------------------------
PASS: 15/15
```

Compared to the implementation of longest_match_api as a built-in function, the implementation of OPA rules clearly has an inefficient aspect, but the above results alone do not show much difference.
 
Let's check the actual performance using the benchmark function. According to the OPA official document, it is guided that to compare benchmark results, you can save benchmark results in the gobench format and compare them in benchstat.
 
Let's install benchstat. If the Go compiler is installed, benchstat is installed by executing the following command

``` 
$ go get -u golang.org/x/perf/cmd/benchstat
```

If the benchstat command is not executed after installation, check GOPATH information by executing the goenv command. You can find it in the bin directory below GOPATH. The location can be added to the PATH
 
First, let's save the performance of the policy.rego written using only the rule as the next command. The benchstat utility does not compare the results of one run with each other, so it was compared five times with the —count option. The OPA official document also explains that 5 to 10 times is enough. The default timeout was 5 seconds, but in the author's environment, a timeout occurred during some rule performance measurements, increasing it to 10 seconds with the –t option. Finally, the binary used opaex, which can be used as a pure opa binary, but when compiling the binary, there can be differences in other factors such as optimization options, and policy.rego can be executed with the opaex binary, so it was executed as an opaex binary to remove external elements

``` 
$ opaex test -v -t 10s --count 5 --format gobench --bench policy.rego policy_test.rego data.json | tee policy.txt
```

The name of the policy file and the name of the file to save the result were changed and executed once more as follows.

``` 
$ opaex test -v -t 10s --count 5 --format gobench —bench policy_extfunc.rego policy_test.rego data.json | tee extfunc.txt
```

Let's compare the performance by transferring two files to the benchstat command. Comparing the performance, it can be seen that when the built-in function is used, the execution time is reduced to about 50%. In other words, it shows twice as much performance.

``` 
$ benchstat policy.txt extfunc.txt
name old time/op new time/op delta
DataExampleUrlhierarchyTestAdminAllowed 90.8µs ± 0% 46.9µs ± 3% -48.39% (p=0.016 n=5+4)
DataExampleUrlhierarchyTestOper1ProfileReadAllowed 179µs ± 1% 78µs ± 4% -56.66% (p=0.016 n=5+4)
DataExampleUrlhierarchyTestOper1ProfileUpdateAllowed 189µs ± 1% 95µs ±34% -49.71% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper2ProfileReadAllowed 174µs ± 1% 79µs ±33% -54.63% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper2ProfileUpdateNotAllowed 188µs ± 0% 88µs ±19% -53.19% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper3ProfileReadNotAllowed 182µs ± 1% 81µs ±13% -55.40% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper3ProfileUpdateNotAllowed 189µs ± 1% 88µs ±15% -53.55% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUserAllowed 95.7µs ± 1% 51.3µs ± 3% -46.41% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUserNotAllowed 98.1µs ± 1% 52.9µs ± 3% -46.13% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUserPublic 69.1µs ± 1% 45.4µs ± 3% -34.30% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserlistReadAllowed 149µs ± 0% 84µs ± 5% -43.92% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserReadAllowed 166µs ± 1% 83µs ± 4% -49.86% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserUpdateNotAllowed 186µs ± 1% 106µs ± 5% -43.20% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper1UserReadNotAllowed 162µs ± 1% 79µs ± 4% -51.02% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestGoodsNotAllowed 94.9µs ± 0% 46.9µs ± 3% -50.57% (p=0.008 n=5+5)
 
name old timer_rego_external_resolve_ns/op new timer_rego_external_resolve_ns/op delta
DataExampleUrlhierarchyTestAdminAllowed 743 ± 1% 259 ± 2% -65.16% (p=0.016 n=5+4)
DataExampleUrlhierarchyTestOper1ProfileReadAllowed 1.25k ± 2% 0.43k ± 1% -65.62% (p=0.016 n=5+4)
DataExampleUrlhierarchyTestOper1ProfileUpdateAllowed 1.25k ± 2% 0.47k ±28% -62.73% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper2ProfileReadAllowed 1.26k ± 4% 0.43k ± 1% -65.92% (p=0.016 n=5+4)
DataExampleUrlhierarchyTestOper2ProfileUpdateNotAllowed 1.26k ± 2% 0.45k ±14% -64.35% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper3ProfileReadNotAllowed 1.25k ± 4% 0.45k ± 5% -64.38% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper3ProfileUpdateNotAllowed 1.26k ± 1% 0.45k ±17% -64.03% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUserAllowed 733 ± 2% 265 ± 3% -63.81% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUserNotAllowed 737 ± 2% 269 ± 4% -63.46% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUserPublic 418 ± 2% 261 ± 8% -37.59% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserlistReadAllowed 770 ± 2% 436 ± 5% -43.34% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserReadAllowed 1.09k ± 2% 0.44k ± 5% -59.75% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserUpdateNotAllowed 1.10k ± 3% 0.44k ± 2% -60.45% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper1UserReadNotAllowed 1.09k ± 2% 0.44k ± 3% -60.19% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestGoodsNotAllowed 778 ± 2% 262 ± 2% -66.26% (p=0.008 n=5+5)
 
name old timer_rego_query_eval_ns/op new timer_rego_query_eval_ns/op delta
DataExampleUrlhierarchyTestAdminAllowed 84.0k ± 0% 40.8k ± 3% -51.50% (p=0.016 n=5+4)
DataExampleUrlhierarchyTestOper1ProfileReadAllowed 156k ± 1% 71k ± 4% -54.33% (p=0.016 n=5+4)
DataExampleUrlhierarchyTestOper1ProfileUpdateAllowed 166k ± 1% 87k ±30% -47.77% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper2ProfileReadAllowed 151k ± 1% 71k ±31% -52.70% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper2ProfileUpdateNotAllowed 165k ± 0% 81k ±17% -51.00% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper3ProfileReadNotAllowed 158k ± 1% 74k ±11% -53.14% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper3ProfileUpdateNotAllowed 165k ± 0% 80k ±13% -51.30% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUserAllowed 88.2k ± 0% 45.0k ± 3% -48.91% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUserNotAllowed 90.3k ± 1% 46.7k ± 3% -48.25% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUserPublic 63.0k ± 0% 39.1k ± 3% -38.03% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserlistReadAllowed 127k ± 1% 77k ± 4% -39.50% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserReadAllowed 143k ± 1% 77k ± 4% -46.38% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserUpdateNotAllowed 163k ± 1% 97k ± 4% -40.63% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper1UserReadNotAllowed 139k ± 1% 73k ± 4% -47.58% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestGoodsNotAllowed 87.6k ± 0% 40.8k ± 3% -53.36% (p=0.008 n=5+5)
 
name old alloc/op new alloc/op delta
DataExampleUrlhierarchyTestAdminAllowed 28.8kB ± 0% 15.9kB ± 0% -44.64% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper1ProfileReadAllowed 54.3kB ± 0% 27.2kB ± 0% -49.85% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper1ProfileUpdateAllowed 59.7kB ± 0% 32.5kB ± 0% -45.56% (p=0.016 n=5+4)
DataExampleUrlhierarchyTestOper2ProfileReadAllowed 51.8kB ± 0% 24.7kB ± 0% -52.27% (p=0.016 n=5+4)
DataExampleUrlhierarchyTestOper2ProfileUpdateNotAllowed 59.0kB ± 0% 31.8kB ± 0% -46.12% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper3ProfileReadNotAllowed 55.8kB ± 0% 28.7kB ± 0% -48.51% (p=0.000 n=5+4)
DataExampleUrlhierarchyTestOper3ProfileUpdateNotAllowed 59.0kB ± 0% 31.7kB ± 0% -46.15% (p=0.016 n=5+4)
DataExampleUrlhierarchyTestUserAllowed 29.9kB ± 0% 17.0kB ± 0% -43.33% (p=0.029 n=4+4)
DataExampleUrlhierarchyTestUserNotAllowed 31.8kB ± 0% 18.8kB ± 0% -40.81% (p=0.029 n=4+4)
DataExampleUrlhierarchyTestUserPublic 22.0kB ± 0% 15.1kB ± 0% -31.53% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserlistReadAllowed 46.0kB ± 0% 30.2kB ± 0% -34.41% (p=0.000 n=5+4)
DataExampleUrlhierarchyTestUseroperUserReadAllowed 51.1kB ± 0% 30.3kB ± 0% -40.78% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserUpdateNotAllowed 63.4kB ± 0% 42.5kB ± 0% -33.00% (p=0.000 n=5+4)
DataExampleUrlhierarchyTestOper1UserReadNotAllowed 49.7kB ± 0% 28.8kB ± 0% -42.00% (p=0.029 n=4+4)
DataExampleUrlhierarchyTestGoodsNotAllowed 31.6kB ± 0% 17.8kB ± 0% -43.71% (p=0.008 n=5+5)
 
name old allocs/op new allocs/op delta
DataExampleUrlhierarchyTestAdminAllowed 647 ± 0% 342 ± 0% -47.14% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper1ProfileReadAllowed 1.23k ± 0% 0.56k ± 0% -54.49% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper1ProfileUpdateAllowed 1.31k ± 0% 0.64k ± 0% -51.30% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper2ProfileReadAllowed 1.19k ± 0% 0.52k ± 0% -56.23% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper2ProfileUpdateNotAllowed 1.29k ± 0% 0.62k ± 0% -52.05% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper3ProfileReadNotAllowed 1.24k ± 0% 0.57k ± 0% -53.74% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper3ProfileUpdateNotAllowed 1.29k ± 0% 0.62k ± 0% -52.05% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUserAllowed 669 ± 0% 360 ± 0% -46.19% (p=0.029 n=4+4)
DataExampleUrlhierarchyTestUserNotAllowed 692 ± 0% 383 ± 0% -44.65% (p=0.029 n=4+4)
DataExampleUrlhierarchyTestUserPublic 471 ± 1% 327 ± 1% -30.43% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserlistReadAllowed 965 ± 0% 604 ± 0% -37.41% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserReadAllowed 1.11k ± 0% 0.61k ± 0% -45.08% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestUseroperUserUpdateNotAllowed 1.28k ± 0% 0.78k ± 0% -39.02% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestOper1UserReadNotAllowed 1.08k ± 0% 0.58k ± 0% -46.29% (p=0.008 n=5+5)
DataExampleUrlhierarchyTestGoodsNotAllowed 681 ± 0% 348 ± 0% -48.90% (p=0.008 n=5+5)
```

Double performance improvement is an impressive figure. However, the embedded function implementation can only be used through a separate binary build or through a method included in the use of the Go client library. Therefore, unless the performance difference is severe due to large data or complicated logic, implementing it with rego-written rules is much easier to manage and maintain, so we recommend implementation as a rule unless high performance is particularly important.

## OPA plug-in
This time, let's look at the expansion of functions through the OPA plug-in. The OPA built-in function provides functions that can be called in rules, while the plug-in implements extensions for REST servers or REPLs running through the opa run command.
 
### Structure of OPA plug-in
What OPA requires to implement a plug-in is the implementation of a plug-in interface and a plug-in factory interface. Two interfaces must be implemented to register as an OPA plug-in. In addition, it provides a plug-in manager to manage the life cycle and state of the plug-in and to inquire with the status API.
 
The functions that the OPA plug-in should implement are as follows. Start, Stop, and Reconfigure are called when the plug-in starts, when it is interrupted, and when it is reset, respectively. Looking at these functions, it can be seen that they are all about the life-cycle of the plug-in.

``` 
type Plugin interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context)
    Reconfigure(ctx context.Context, config interface{})
}
```

The interfaces that the plug-in factory should implement are as follows. Validate verifies that there is no problem with the configuration of the plug-in and performs tasks such as applying the default value if not set. New creates and returns a new plug-in instance with the configurations it has taken over.

``` 
type Factory interface {
    Validate(manager *Manager, config []byte) (interface{}, error)
    New(manager *Manager, config Plugin
}
```
 
The plug-in may be registered by calling runtime.RegisterPlugin() before cmd.RootCommand.Execute() is called. The first argument in RegisterPlugin is the name of the plugin, and the second argument is the factory that creates the plugin.

To use the plug-in, you must set the plugins section of the OPA configuration file. The structure of the plugins section is as follows.

``` 
plugins:
  <Plug-in 1 Name>:
    <Plug-in1 Property1 Name>: <Plug-in1 Property1 Value>
    <Plug-in 1 Property 2 Name>: <Plug-in 1 Property 2 Value>
...
  <Plug-in 2 Name>:
    <Plug-in 2 Property 1 Name>: <Plug-in 2 Property 1 Value>
    <Plug-in 2 Property 1 Name>: <Plug-in 2 Property 1 Value>
...
```

When the OPA configuration file is loaded, the configuration information in the plugins section is transferred to the plugin having the same name as JSON with the corresponding <Property Name>: <Property Value>.
 
### Plug-in example
Let's take an example and see how the actual plug-in works. The plug-in to create is a ping server plug-in that responds with a simple message when sending ping requests that are frequently used in various network program examples.

The first thing to do is to create a project. After creating a directory called pingplugin as follows, initialize the go module.

```
$ mkidr pingplugin
$ cd pingplugin
$ go mod init
```

Before writing the code, let's define the plug-in configurations. The values to be used as configurations in the ping plug-in are port and response messages for the server, which are defined as follows, and stored under the name oppa_plugin.conf. This file is an OPA configuration file containing only plug-in settings and is in YAML format. Note that we need a space behind the colon.

```
plugins:
  pingpong_plugin:
    port: 9992
    msg: Pong!!!
```

In the plug-in example, the file was divided into three for code readability: main.go, factory.go, and plugin.go. Since it belongs to the same main package, splitting files does not affect each other's reference.

First of all, the contents of main.go are as follows.

``` 
package main
 
import (
    "fmt"
    "os"
 
    "github.com/open-policy-agent/opa/cmd"
    "github.com/open-policy-agent/opa/runtime"
)
 
func main() {
    runtime.RegisterPlugin(PluginName, Factory{})
 
    if err := cmd.RootCommand.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

As described above, register the plug-in with runtime.RegisterPlugin() before executing cmd.RootCommand.Execute(). PluginName and Factory referred to what was declared in another file in the main package. The contents of plugin.go containing the implementation of plugin are as follows.
 
The contents of plugin.go containing the implementation of plugin are as follows.

```
package main
 
import (
    "context"
    "fmt"
    "log"
    "net/http"
    "sync"
 
    "github.com/open-policy-agent/opa/plugins"
)
 
const PluginName = "pingpong_plugin"
 
const DefaultPort = 9999
const DefaultMessage = "Pong!"
 
type Config struct {
    Port int32 `json:"port"`
    Msg string `json:"msg"`
}
 
type PingPongServer struct {
    manager *plugins.Manager
    mtx sync.Mutex
    config Config
}
 
func (p *PingPongServer) Start(ctx context.Context) error {
    log.Printf("Start PingPong Server, config=%+v\n", p.config)
 
    start(&p.config)
 
    p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})
 
    return nil
}
 
func (p *PingPongServer) Stop(ctx context.Context) {
    log.Println("Stop PingPong Server")
 
    p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateNotReady})
}
 
func (p *PingPongServer) Reconfigure(ctx context.Context, config interface{}) {
    p.mtx.Lock()
    defer p.mtx.Unlock()
 
    p.config = config.(Config)
}
 
func start(config *Config) {
    listen := fmt.Sprintf(":%d", config.Port)
 
    log.Printf("listen addr=%s", listen)
 
    http.HandleFunc("/ping", func(res http.ResponseWriter, req *http.Request) {
        fmt.Fprintf(res, "%s\n", config.Msg)
    })

    log.Fatal(http.ListenAndServe(listen, nil))
}
```

The PluginName constant referenced in main.go was defined in this file. The plug-in name is pingpong_plugin, and the plug-in configuration has a port field containing a port to operate the service and an msg field containing a response message. The plug-in's Start() function starts the HTTP server and changes the plug-in state to OK through the plug-in manager. The Stop() function of the plug-in changes the state of the plug-in to NotReady. The Configure() function is a function called when changing configuration, and it can be called multiple times concurrently, so the configuration will be guarded by a lock.

```
package main
 
import (
    "log"
 
    "github.com/open-policy-agent/opa/plugins"
    "github.com/open-policy-agent/opa/util"
)
 
type Factory struct{}
 
func (Factory) New(m *plugins.Manager, config interface{}) plugins.Plugin {
 
    log.Printf("Config=%+v\n", config)
 
    m.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateNotReady})
 
    return &PingPongServer{
        manager: m,
        config: config.(Config),
    }
}
 
func (Factory) Validate(_ *plugins.Manager, config []byte) (interface{}, error) {
    parsedConfig := Config{}
    err := util.Unmarshal(config, &parsedConfig)
 
    if err != nil {
        log.Println("Error occured while validate config:%v\n", err)
    } else {
        if parsedConfig.Msg == "" {
            parsedConfig.Msg = DefaultMessage
        }
 
        if parsedConfig.Port == 0 {
            parsedConfig.Port = DefaultPort
        }
    }
 
    return parsedConfig, err
}
```

New and Validate functions for plug-in factory and the Factory structure implemented. The New function initializes the plug-in state to the StatusNotReady state and returns the PingPongServer struct with the pointer of the plug-in manager and the pointer of the configuration. The Validate function parses the configurations and places them in the Config struct declared in plugin.go. If the Msg or Port field is not set, the default constants declared in plugin.go are set.
  
To build the code, the following command can be executed, and the binary to be generated is named opapp.

``` 
$ go build –o opapp
```

Once the build is complete, let's operate the server with a configuration file with oppa_plugin.conf, as specified when operating the OPA server.

``` 
$ ./opapp run -s –c opa_plugin.conf
2021/02/24 15:09:32 Config={Port:9992 Msg:Ping!!!}
{"addrs":[":8181"],"diagnostic-addrs":[],"level":"info","msg":"Initializing server.","time":"2021-02-24T15:09:32+09:00"}
2021/02/24 15:09:32 Start PingPong Server, config={Port:9992 Msg:Pong!!!}
2021/02/24 15:09:32 listen addr=:9992
```

When the service operates, sending an http request to the /ping URL for the port responds with a configured message.

``` 
$ curl localhost:9992/ping
Pong!!!
```

### Setting Plug-in Version 
Let's check the version of the opapp binary with the ping plug-in added with the next command. When checking the contents, the OPA version, Go version, and web assembly support were displayed, but information related to the build was not displayed.

``` 
$ ./opapp version
Version: 0.26.0
Build Commit:
Build Timestamp:
Build Hostname:
Go Version: go1.15.2
WebAssembly: unavailable
```

If there is a way to express that the version information is a different OPA binary from the official version including the ping plug-in, modifying the version information may prevent confusion.

The following is the content of the version/version.go file containing the output version information among the OPA source codes. When importing the OPA library into the module, a stable version of 0.26, is installed, and since it is the master branch of the source code, the version variable is set to 0.27.0-dev, which is the current development version.

``` 
package version
 
import (
    "runtime"
)
 
var Version = "0.27.0-dev"
 
var GoVersion = runtime.Version()
 
var (
    Vcs = ""
    Timestamp = ""
    Hostname = ""
)
```

In order to change the above values so that any version is recorded and output in the binary, you can override a specific variable value by setting a compiler option at the time of build. It is similar to setting the property with the -D option when executing Java, but the difference is that it is applied when compiling and included in the binary.
 
Let's rebuild using the next command. The version.version variable of opa was set to 0.26+pingplugin to show that the OPA 0.26 version contains a ping plugin, and the host name where the source code was built was set to devpc

``` 
$ go build –o opapp -ldflags "-X github.com/open-policy-agent/opa/version.Version=0.26+pingplugin –X github.com/open-policy-agent/opa/version.Hostname=devpc“
```

If you check the version command again for the modified binary, it changed as follows.

``` 
$ ./opapp version
Version: 0.26+pingplugin
Build Commit:
Build Timestamp:
Build Hostname: devpc
Go Version: go1.15.2
WebAssembly: unavailable
```

Commit information and timestamp related to the build are different only from the field name, but can be changed in the same way. Changing to build relevant information has to work and CI/CD will be a great help to manage the version.

## Special OPA plugins
The OPA plug-in does not require anything special other than implementing the plug-in factory interface and the plug-in interface, but some of the OPA's features define a more specific interface so that it can easily replace certain functions of the OPA with other implementations.

The current OPA version (0.26 when writing) defines two more specific plug-in interfaces: an HTTP authentication plug-in that allows users to set authentication methods, such as tokens, and a policy logger plug-in that leaves policy logs.

In addition to the Plugin interface, the HTTP authentication plugin must implement the following HTTPAuthPlugin interface. The NewClient function creates an HTTP client from the setting and returns pointers and errors. The Prepare function takes over the pointer of the http.Request structure and implements the task of setting the token in the HTTP header before HTTP request.

``` 
type HTTPAuthPlugin interface {
    NewClient(c Config) (*http.Client, error)
    Prepare(req *http.Request) error
}
```

When using an HTTP authentication plugin, you can declare plugin and plugin properties in the plugins section, and then specify the plugin name in the plugin properties of the services.<service name>.credentials section. For example, if you have developed an HTTPAuthPlugin plugin named my_auth which accepts token abc1234, you can configure it as follows.

```
services:
  my_service:
     url: https://localhost/sevice/v1/
     credentials:
        plugin: my_auth
  plugins:
    my_auth:
      token: abc1234
```

Examples of implementations of HTTP authentication plug-ins are well explained on the official website, so they are not explained again in this book. Related information can be found at https://www.openpolicyagent.org/docs/latest/configuration/#example-4.

The policy logger plug-in requires additional implementation of the Log function on the existing plug-in interface as follows. Declaring other interfaces inside the interface, such as plugin.Plugin, means that in the Go language, the inner interface must be also implemented. Since plugin.Plugin must be implemented when implementing the HTTP authentication plugin above, there is no difference even if HTTPAuthPlugin is forced to implement plugin.Plugin.

```
type Logger interface {
    plugins.Plugin
 
    Log(context.Context, EventV1) error
}
```

The decision logger plug-in can be set in a similar way. If you have developed a plug-in called new_decision_logger and have the settings set to leave a log only if an error occurs, the settings will be as follows.

``` 
decision_logs:
  plugin: new_decision_logger
plugins:
  new_decision_logger:
    onlyerror: false
```

Examples of decision logger plugins are well explained on the official website, so they are not explained again in this book. Related information can be found at https://www.openpolicyagent.org/docs/latest/extensions/#putting-it-together.
 
As OPA development progresses further, it is expected that there will be more and more parts that can replace the basic functions of OPA with plug-ins. In addition, the functions of plug-in and OPA exchanging information and controlling plug-in will gradually expand. At the time readers read books, more may be implemented as plug-ins, so it is recommended to check the OPA official document once again.

## Summary
Chapter 10 describes how to expand OPA through built-in functions and plug-ins. In Chapter 5, the part that could not be efficiently implemented when implemented as a rule was re-implemented and operated as a built-in function. It also looked at how to build an OPA binary including extensions and how to modify version information of a newly created binary. Finally, HTTP authentication and decision logger plugins that can be replaced through a more specific plug-in interface among the basic operations of OPA have also been briefly described.
 
The contents described in Chapter 10 are those that must be developed in the Go language. Chapter 11 describes a web assembly that shows the direction that can be used without developing it in a Go language or operating OPA as a separate REST server.
