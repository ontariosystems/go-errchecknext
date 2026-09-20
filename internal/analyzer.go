package internal

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var (
	Analyzer = &analysis.Analyzer{
		Name: name,
		Doc:  "require an immediate error check after assigning to err",
		Run:  run,
	}

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

const (
	name                   = "errchecknext"
	noFollowingStatement   = "assignment to err is not immediately followed by an error check"
	noNextStatmentNotCheck = "statement between assignment to err and error check"
)

// AnalysisInfo information about the files being scanned, the types from the ast, and reporting
type AnalysisInfo struct {
	Files     []*ast.File
	Fset      *token.FileSet
	TypesInfo *types.Info
	Collector *Collector
	Report    func(analysis.Diagnostic)
}

// run the main run method of the go analyzer framework, calls our Analyze method so we can use the
// same logic from our main function
func run(pass *analysis.Pass) (any, error) {
	info := &AnalysisInfo{
		Files:     pass.Files,
		Fset:      pass.Fset,
		TypesInfo: pass.TypesInfo,
		Report:    pass.Report,
	}
	if err := Analyze(info); err != nil {
		return nil, err
	}

	return nil, nil
}

type funcInfo struct {
	Type *ast.FuncType
	Body *ast.BlockStmt
}

// Analyze inspect code in packages specified by AnalysisInfo. Called either via the main function
// or go's analyzer framework
func Analyze(info *AnalysisInfo) (err error) {
	for _, file := range info.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			fnInfo := &funcInfo{fn.Type, fn.Body}
			if err = checkBlock(info, file, fnInfo, fn.Body); err != nil {
				return false
			}
			return true
		})
	}
	return
}

// checkBlock loops through all the statements in a block and reports findings if:
//  1. the statement assigns an error variable; and either
//  2. there is no following statement; or
//  3. the following statement is not an if statement checking the error
func checkBlock(info *AnalysisInfo, file *ast.File, fnInfo *funcInfo, block *ast.BlockStmt) (err error) {
	stmts := block.List

	for i, stmt := range stmts {
		ast.Inspect(stmt, func(n ast.Node) bool {
			var nestedFnInfo *funcInfo
			switch n := n.(type) {
			case *ast.FuncLit:
				nestedFnInfo = &funcInfo{n.Type, n.Body}
			case *ast.BlockStmt:
				nestedFnInfo = &funcInfo{fnInfo.Type, n}
			default:
				return true
			}

			if nestedErr := checkBlock(info, file, nestedFnInfo, nestedFnInfo.Body); nestedErr != nil {
				err = errors.Join(err, nestedErr)
				return false
			}
			return true
		})

		assignedErrObj := assignedErrorObject(info, stmt)
		if assignedErrObj == nil {
			continue
		}

		// no following statement
		if i+1 >= len(stmts) {
			report(info, file, stmt.Pos(), noFollowingStatement)
			continue
		}

		if !isAllowedNextStatement(info, stmts[i+1], fnInfo, assignedErrObj) {
			report(info, file, stmts[i+1].Pos(), noNextStatmentNotCheck)
		}
	}
	return nil
}

// assignedErrorObject returns list of error objects assigned on the left-hand side
func assignedErrorObject(info *AnalysisInfo, stmt ast.Stmt) types.Object {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok {
		return nil
	}

	if isErrorConstructor(info, assign) {
		return nil
	}

	for _, lhs := range assign.Lhs {
		id, ok := lhs.(*ast.Ident)
		if !ok || id.Name == "_" {
			continue
		}

		if isErrorType(info.TypesInfo.TypeOf(id)) {
			return info.TypesInfo.ObjectOf(id)
		}
	}

	return nil
}

// isErrorConstructor returns true if the RHS of the assignment is in the ignoredCalls map of
// allowed error constructing functions.
func isErrorConstructor(info *AnalysisInfo, assign *ast.AssignStmt) bool {
	if len(assign.Rhs) != 1 {
		return false
	}

	call, ok := assign.Rhs[0].(*ast.CallExpr)
	if !ok {
		return false
	}

	fnObj := functionObject(info, call.Fun)
	_, ok = ignoredCalls[functionName(fnObj)]
	return ok
}

// functionObject returns the type object for either an identifier or selector
func functionObject(info *AnalysisInfo, expr ast.Expr) types.Object {
	switch fn := expr.(type) {
	case *ast.Ident:
		return info.TypesInfo.ObjectOf(fn)
	case *ast.SelectorExpr:
		return info.TypesInfo.ObjectOf(fn.Sel)
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

// isAllowedNextStatement returns true if the statement is either an error check or returning the error
func isAllowedNextStatement(info *AnalysisInfo, stmt ast.Stmt, fnInfo *funcInfo, assignedErrObj types.Object) bool {
	return isErrCheck(info, stmt) || isErrReturn(info, stmt, fnInfo, assignedErrObj)
}

// isErrCheck returns true if a statement is an if that checks an error
func isErrCheck(info *AnalysisInfo, stmt ast.Stmt) bool {
	ifs, ok := stmt.(*ast.IfStmt)
	if !ok {
		return false
	}

	return conditionReferencesErr(info, ifs.Cond)
}

// isErrReturn returns true if the statement returns the error
func isErrReturn(info *AnalysisInfo, stmt ast.Stmt, fnInfo *funcInfo, assignedErrObj types.Object) bool {
	ret, ok := stmt.(*ast.ReturnStmt)
	if !ok {
		return false
	}

	if isBareReturnOfAssignedError(info, stmt, fnInfo, assignedErrObj) {
		return true
	}

	for _, result := range ret.Results {
		if t := info.TypesInfo.TypeOf(result); t != nil && isErrorType(t) {
			return true
		}
	}

	return false
}

// namedErrorReturns returns list of obj references for named returns that are error type
func namedErrorReturns(info *AnalysisInfo, fnInfo *funcInfo) (result []types.Object) {
	if fnInfo.Type.Results == nil {
		return
	}

	for _, field := range fnInfo.Type.Results.List {
		if !isErrorType(info.TypesInfo.TypeOf(field.Type)) {
			continue
		}

		for _, name := range field.Names {
			result = append(result, info.TypesInfo.ObjectOf(name))
		}
	}

	return
}

// isBareReturnOfAssignedError returns true if we have a bare return, named returns, and the error is one of the values
func isBareReturnOfAssignedError(info *AnalysisInfo, stmt ast.Stmt, fnInfo *funcInfo, assignedErrObj types.Object) bool {
	ret, ok := stmt.(*ast.ReturnStmt)
	if !ok {
		return false
	}

	if len(ret.Results) != 0 {
		return false
	}

	return slices.Contains(namedErrorReturns(info, fnInfo), assignedErrObj)
}

// conditionReferencesErr returns true if a conditional expression references an error type
func conditionReferencesErr(info *AnalysisInfo, expr ast.Expr) bool {
	found := false

	ast.Inspect(expr, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if obj := info.TypesInfo.ObjectOf(id); obj != nil && isErrorType(obj.Type()) {
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

// report sends the code position to the analysis pass.Report or the Collector if not suppressed
func report(info *AnalysisInfo, file *ast.File, pos token.Pos, message string) {
	if isSuppressed(file, info.Fset, pos) {
		return
	}

	if info.Report != nil {
		info.Report(analysis.Diagnostic{Pos: pos, Message: message})
	}

	if info.Collector != nil {
		info.Collector.ReportFinding(info.Fset, pos, message)
	}
}

// isSuppressed returns true when the line being reported has a comment with nolint
func isSuppressed(file *ast.File, fileSet *token.FileSet, pos token.Pos) bool {
	reportLine := fileSet.Position(pos).Line
	return slices.ContainsFunc(file.Comments, func(cg *ast.CommentGroup) bool {
		start := fileSet.Position(cg.Pos()).Line
		end := fileSet.Position(cg.End()).Line

		sameLine := start == end && start == reportLine
		return sameLine && slices.ContainsFunc(cg.List, hasNoLint)
	})
}

// hasNoLint returns true if we have //nolint or //nolint:errchecknext,
// or it's in the comma-separated list of ignored linters.
func hasNoLint(comment *ast.Comment) bool {
	text := strings.TrimPrefix(comment.Text, "//")

	if !strings.HasPrefix(text, "nolint") {
		return false
	}

	parts := strings.SplitN(text, ":", 2)
	if len(parts) == 1 {
		// bare nolint comment
		return true
	}

	// contains our linter name
	return slices.Contains(strings.Split(parts[1], ","), name)
}
