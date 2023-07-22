package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/henderiw/value-propagation/pkg/data"
	"github.com/pkg/errors"
)

const (
	celExprPrefix = "celexpr: "
)

const (
	index              = "index"
	key                = "key"
	value              = "value"
	var1               = "var1"
	replicaSetFileName = "../data/node-replicaset-cel2.yaml"
	endpointFileName   = "../data/endpoint_leaf1-e1-1.yaml"
)

func main() {

	// prepare data
	cr := data.GetReplicaSet(replicaSetFileName)
	fmt.Println("template data:\n", string(cr.Spec.Template.Raw))

	x := map[string]any{}
	if err := json.Unmarshal(cr.Spec.Template.Raw, &x); err != nil {
		panic(err)
	}

	// get ep data
	ep := data.GetEndpoint(endpointFileName)

	if cr.Spec.Replicas != nil {
		for i := 0; i < int(*cr.Spec.Replicas); i++ {
			r := newRenderer(map[string]any{
				index: strconv.Itoa(i),
				var1:  ep,
			})

			newx, err := DeepCopy(x)
			if err != nil {
				panic(err)
			}
			fmt.Printf("orig data:\n%v\n", newx)

			rx, err := r.Render(newx)
			if err != nil {
				panic(err)
			}

			b, err := json.MarshalIndent(rx, "", "  ")
			if err != nil {
				panic(err)
			}

			fmt.Printf("result: %d\n%v\n", i, string(b))

		}
	}

	/*
		for _, str := range strs {
			if strings.HasPrefix(str, celExprPrefix) {
				//fmt.Println(strings.TrimPrefix(str, celExprPrefix))

				ast, iss := env.Compile(strings.TrimPrefix(str, celExprPrefix))

				if iss.Err() != nil {
					panic(iss.Err())
				}
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

				i := 22
				result, _, err := program.Eval(map[string]any{
					index: strconv.Itoa(i),
				})
				if err != nil {
					panic(err)
				}

				fmt.Printf("result %d:\n%v\n", i, result)
			}
		}
	*/
}

func getCelEnv() (*cel.Env, error) {
	var opts []cel.EnvOption
	opts = append(opts, cel.HomogeneousAggregateLiterals())
	//opts = append(opts, cel.EagerlyValidateDeclarations(true), cel.DefaultUTCTimeZone(true))
	//opts = append(opts, library.ExtensionLibs...)
	opts = append(opts, cel.Variable(index, cel.StringType))
	opts = append(opts, cel.Variable(key, cel.StringType))
	opts = append(opts, cel.Variable(value, cel.DynType))
	opts = append(opts, cel.Variable(var1, cel.DynType))

	return cel.NewEnv(opts...)
}

// Make a deep copy from in into out object.
func DeepCopy(in interface{}) (interface{}, error) {
	if in == nil {
		return nil, errors.New("in cannot be nil")
	}
	//fmt.Printf("json copy input %v\n", in)
	bytes, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal input data")
	}
	var out interface{}
	err = json.Unmarshal(bytes, &out)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal to output data")
	}
	//fmt.Printf("json copy output %v\n", out)
	return out, nil
}

type Renderer interface {
	Render(x any) (any, error)
}

type vars map[string]any

func newRenderer(v vars) Renderer {
	return &renderer{
		vars: v,
	}
}

type renderer struct {
	vars map[string]any
}

func (r *renderer) Render(x any) (any, error) {
	var err error
	switch x := x.(type) {
	case map[string]any:
		for k, v := range x {
			x[k], err = r.Render(v)
			if err != nil {
				return nil, err
			}
		}
	case []any:
		for i, t := range x {
			x[i], err = r.Render(t)
			if err != nil {
				return nil, err
			}
		}
	case string:
		if strings.HasPrefix(x, celExprPrefix) {
			fmt.Printf("expression: %s\n", strings.TrimPrefix(x, celExprPrefix))
			//return strings.TrimPrefix(x, celExprPrefix), nil
			env, err := getCelEnv()
			if err != nil {
				return nil, err
			}
			ast, iss := env.Compile(strings.TrimPrefix(x, celExprPrefix))
			if iss.Err() != nil {
				//panic(iss.Err())
				return nil, err
			}
			_, err = cel.AstToCheckedExpr(ast)
			if err != nil {
				//panic(err)
				return nil, err
			}
			prog, err := env.Program(ast,
				cel.EvalOptions(cel.OptOptimize),
				// TODO: uncomment after updating to latest k8s
				//cel.OptimizeRegex(library.ExtensionLibRegexOptimizations...),
			)
			if err != nil {
				//panic(err)
				return nil, err
			}

			fmt.Println("vars", r.vars)

			val, _, err := prog.Eval(r.vars)
			if err != nil {
				return nil, err
			}

			result, err := val.ConvertToNative(reflect.TypeOf(""))
			if err != nil {
				return nil, err
			}

			s, ok := result.(string)
			if !ok {
				return nil, fmt.Errorf("expression returned non-string value: %v", result)
			}
			return s, nil

		}
	}
	return x, err
}
