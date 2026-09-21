package internal

import (
	"go/ast"
	"go/types"
)

var (
	errorObj       = types.Universe.Lookup("error")
	errorInterface = errorObj.Type().Underlying().(*types.Interface)

	ignoredCalls = map[string]struct{}{
		"fmt.Errorf":  {},
		"errors.New":  {},
		"errors.Join": {},
		"github.com/hashicorp/go-multierror.Append": {},
		"github.com/hashicorp/errwrap.Wrap":         {},
		"github.com/hashicorp/errwrap.Wrapf":        {},
		"github.com/pkg/errors.Errorf":              {},
		"github.com/pkg/errors.New":                 {},
		"github.com/pkg/errors.WithMessage":         {},
		"github.com/pkg/errors.WithMessagef":        {},
		"github.com/pkg/errors.WithStack":           {},
		"github.com/pkg/errors.Wrap":                {},
		"github.com/pkg/errors.Wrapf":               {},
	}
)

// assignedErrorObject returns list of error objects assigned on the left-hand side
func assignedErrorObject(typesInfo *types.Info, stmt ast.Stmt) types.Object {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok {
		return nil
	}

	if isErrorConstructor(typesInfo, assign) {
		return nil
	}

	for _, lhs := range assign.Lhs {
		id, ok := lhs.(*ast.Ident)
		if !ok || id.Name == "_" {
			continue
		}

		if isErrorType(typesInfo.TypeOf(id)) {
			return typesInfo.ObjectOf(id)
		}
	}

	return nil
}

// isErrorConstructor returns true if the RHS of the assignment is in the ignoredCalls map of
// allowed error constructing functions.
func isErrorConstructor(typesInfo *types.Info, assign *ast.AssignStmt) bool {
	if len(assign.Rhs) != 1 {
		return false
	}

	call, ok := assign.Rhs[0].(*ast.CallExpr)
	if !ok {
		return false
	}

	fnObj := functionObject(typesInfo, call.Fun)
	_, ok = ignoredCalls[functionName(fnObj)]
	return ok
}

// functionObject returns the type object for either an identifier or selector
func functionObject(typesInfo *types.Info, expr ast.Expr) types.Object {
	switch fn := expr.(type) {
	case *ast.Ident:
		return typesInfo.ObjectOf(fn)
	case *ast.SelectorExpr:
		return typesInfo.ObjectOf(fn.Sel)
	default:
		return nil
	}
}

// functionName returns the package and name of an object given its type
func functionName(obj types.Object) string {
	if obj == nil || obj.Pkg() == nil {
		return ""
	}

	return obj.Pkg().Path() + "." + obj.Name()
}

// isErrCheck returns true if a statement is an if that checks an error
func isErrCheck(typesInfo *types.Info, stmt ast.Stmt) bool {
	ifs, ok := stmt.(*ast.IfStmt)
	if !ok {
		return false
	}

	return conditionReferencesErr(typesInfo, ifs.Cond)
}

// conditionReferencesErr returns true if a conditional expression references an error type
func conditionReferencesErr(typesInfo *types.Info, expr ast.Expr) bool {
	found := false

	ast.Inspect(expr, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if obj := typesInfo.ObjectOf(id); obj != nil && isErrorType(obj.Type()) {
				found = true
				return false
			}
		}
		return true
	})

	return found
}

// isErrorType returns true if the type is an error
func isErrorType(t types.Type) bool {
	if t == nil {
		return false
	}

	if types.Implements(t, errorInterface) {
		return true
	}

	if p, ok := t.(*types.Pointer); ok {
		return types.Implements(p, errorInterface)
	}

	return false
}
