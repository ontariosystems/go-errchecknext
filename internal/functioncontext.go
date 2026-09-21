package internal

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"
)

type FunctionContext struct {
	Info            *AnalysisContext
	File            *ast.File
	Type            *ast.FuncType
	Body            *ast.BlockStmt
	NamedErrReturns map[types.Object]struct{}
}

var zeroStruct = struct{}{}

func NewFunctionContext(info *AnalysisContext, file *ast.File, n ast.Node) *FunctionContext {
	var (
		ft *ast.FuncType
		b  *ast.BlockStmt
	)

	switch n := n.(type) {
	case *ast.FuncDecl:
		ft, b = n.Type, n.Body
	case *ast.FuncLit:
		ft, b = n.Type, n.Body
	default:
		return nil
	}

	return &FunctionContext{
		Info:            info,
		File:            file,
		Type:            ft,
		Body:            b,
		NamedErrReturns: namedErrorReturns(info.TypesInfo, ft),
	}
}

// namedErrorReturns returns map with keys of obj references for named returns that are error type
func namedErrorReturns(typesInfo *types.Info, funcType *ast.FuncType) (result map[types.Object]struct{}) {
	result = make(map[types.Object]struct{})
	if funcType.Results == nil {
		return
	}

	for _, field := range funcType.Results.List {
		if !isErrorType(typesInfo.TypeOf(field.Type)) {
			continue
		}
		for _, name := range field.Names {
			obj := typesInfo.ObjectOf(name)
			result[obj] = zeroStruct
		}
	}
	return
}

// AnalyzeFunction analyzes function literals (creating a new nested FunctionContext) or block statements (including nested block statements)
func (fnCtx *FunctionContext) AnalyzeFunction() {
	ast.Inspect(fnCtx.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			NewFunctionContext(fnCtx.Info, fnCtx.File, n).AnalyzeFunction()
			return false // do not continue to run inspect, analyzeFunction will start it over
		case *ast.BlockStmt:
			fnCtx.AnalyzeBlock(n)
			return true // continue to run analyzeFunction, for nested blocks and function literals
		}
		return true
	})
}

// AnalyzeBlock loops through all the statements in a block and reports findings if:
//  1. the statement assigns an error variable; and either
//  2. there is no following statement; or
//  3. the following statement is not an if statement checking the error
func (fnCtx *FunctionContext) AnalyzeBlock(block *ast.BlockStmt) {
	stmts := block.List

	for i, stmt := range stmts {
		assignedErrObj := assignedErrorObject(fnCtx.Info.TypesInfo, stmt)
		if assignedErrObj == nil {
			continue
		}

		// no following statement
		if i+1 >= len(stmts) {
			fnCtx.report(stmt.Pos(), noFollowingStatement)
			continue
		}

		if !fnCtx.isAllowedNextStatement(stmts[i+1], assignedErrObj) {
			fnCtx.report(stmts[i+1].Pos(), noNextStatmentNotCheck)
		}
	}
}

// isErrReturn returns true if the statement returns the error
func (fnCtx *FunctionContext) isErrReturn(stmt ast.Stmt, assignedErrObj types.Object) bool {
	ret, ok := stmt.(*ast.ReturnStmt)
	if !ok {
		return false
	}

	if fnCtx.isBareReturnOfAssignedError(stmt, assignedErrObj) {
		return true
	}

	for _, result := range ret.Results {
		if t := fnCtx.Info.TypesInfo.TypeOf(result); t != nil && isErrorType(t) {
			return true
		}
	}

	return false
}

// isBareReturnOfAssignedError returns true if we have a bare return, named returns, and the error is one of the values
func (fnCtx *FunctionContext) isBareReturnOfAssignedError(stmt ast.Stmt, assignedErrObj types.Object) bool {
	ret, ok := stmt.(*ast.ReturnStmt)
	if !ok {
		return false
	}

	// no returned results, therefore not an error assigned to named return
	if len(ret.Results) != 0 {
		return false
	}

	// see if this object is one of those assigned to the returns
	_, ok = fnCtx.NamedErrReturns[assignedErrObj]
	return ok
}

// isAllowedNextStatement returns true if the statement is either an error check or returning the error
func (fnCtx *FunctionContext) isAllowedNextStatement(stmt ast.Stmt, assignedErrObj types.Object) bool {
	return isErrCheck(fnCtx.Info.TypesInfo, stmt) || fnCtx.isErrReturn(stmt, assignedErrObj)
}

// report sends the code position to the analysis pass.Report or the Collector if not suppressed
func (fnCtx *FunctionContext) report(pos token.Pos, message string) {
	if fnCtx.isSuppressed(pos) {
		return
	}

	if fnCtx.Info.Report != nil {
		fnCtx.Info.Report(analysis.Diagnostic{Pos: pos, Message: message})
	}

	if fnCtx.Info.Collector != nil {
		fnCtx.Info.Collector.ReportFinding(fnCtx.Info.Fset, pos, message)
	}
}

// isSuppressed returns true when the line being reported has a comment with nolint
func (fnCtx *FunctionContext) isSuppressed(pos token.Pos) bool {
	reportLine := fnCtx.Info.Fset.Position(pos).Line
	return slices.ContainsFunc(fnCtx.File.Comments, func(cg *ast.CommentGroup) bool {
		start := fnCtx.Info.Fset.Position(cg.Pos()).Line
		end := fnCtx.Info.Fset.Position(cg.End()).Line

		sameLine := start == end && start == reportLine
		return sameLine && slices.ContainsFunc(cg.List, hasNoLint)
	})
}
