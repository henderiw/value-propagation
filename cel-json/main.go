package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
	"github.com/henderiw/value-propagation/pkg/data"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	
)

const (
	index              = "index"
	key                = "key"
	value              = "value"
	var1               = "var1"
	replicaSetFileName = "../data/node-replicaset-cel.yaml"
	endpointFileName   = "../data/endpoint_leaf1-e1-1.yaml"
)

func main() {
	var opts []cel.EnvOption
	//opts = append(opts, cel.HomogeneousAggregateLiterals())
	//opts = append(opts, cel.EagerlyValidateDeclarations(true), cel.DefaultUTCTimeZone(true))
	//opts = append(opts, library.ExtensionLibs...)
	opts = append(opts, cel.Variable(index, cel.StringType))
	opts = append(opts, cel.Variable(key, cel.StringType))
	opts = append(opts, cel.Variable(value, cel.DynType))
	opts = append(opts, cel.Variable(var1, cel.DynType))

	env, err := cel.NewEnv(opts...)
	if err != nil {
		panic(err)
	}

	cr := data.GetReplicaSet(replicaSetFileName)
	fmt.Println("template data:\n", string(cr.Spec.Template.Raw))

	x := map[string]any{}
	if err := json.Unmarshal(cr.Spec.Template.Raw, &x); err != nil {
		panic(err)
	}
	b, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	fmt.Println("template data:\n", string(b))

	/*
		ast, issues := env.Compile(string(cr.Spec.Template.Raw))
		if issues != nil {
			panic(issues.Err())
		}
	*/
	//ast := compile(env, string(b), cel.MapType(cel.StringType, cel.DynType))
	ast := compile(env, `
		{'apiVersion': 'inv.nephio.org/v1alpha1',
		'kind': 'node',
		'metadata': {
			'name': 'server-' + index,
			'namespace': 'default',
			'labels': {
				'xxxx': var1.spec.interfaceName
			}
		},
		'spec': {
			'provider': 'server.nephio.com'
		}}`,
		cel.MapType(cel.StringType, cel.DynType))

	fmt.Println(ast.OutputType())
	_, err = cel.AstToCheckedExpr(ast)
	if err != nil {
		panic(err)
	}

	program, err := env.Program(ast,
		cel.EvalOptions(cel.OptOptimize),
		// TODO: uncomment after updating to latest k8s
		//cel.OptimizeRegex(library.ExtensionLibRegexOptimizations...),
	)
	if err != nil {
		panic(err)
	}

	ep := data.GetEndpoint(endpointFileName)

	if cr.Spec.Replicas != nil {
		for i := 0; i < int(*cr.Spec.Replicas); i++ {
			result, _, err := program.Eval(map[string]any{
				index: strconv.Itoa(i),
				var1: ep,
			})
			if err != nil {
				panic(err)
			}

			fmt.Printf("result %d:\n%s\n", i, valueToJSON(result))
		}
	}
}

// compile will parse and check an expression `expr` against a given
// environment `env` and determine whether the resulting type of the expression
// matches the `exprType` provided as input.
func compile(env *cel.Env, expr string, celType *cel.Type) *cel.Ast {
	ast, iss := env.Compile(expr)
	if iss.Err() != nil {
		panic(iss.Err())
	}
	if !reflect.DeepEqual(ast.OutputType(), celType) {
		panic(fmt.Errorf(
			"got %v, wanted %v result type", ast.OutputType(), celType))
	}
	//fmt.Printf("%s\n\n", strings.ReplaceAll(expr, "\t", " "))
	return ast
}

// valueToJSON converts the CEL type to a protobuf JSON representation and
// marshals the result to a string.
func valueToJSON(val ref.Val) string {
	v, err := val.ConvertToNative(reflect.TypeOf(&structpb.Value{}))
	if err != nil {
		panic(err)
	}
	marshaller := protojson.MarshalOptions{Indent: "    "}
	bytes, err := marshaller.Marshal(v.(proto.Message))
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
