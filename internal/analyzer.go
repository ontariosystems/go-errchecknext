package internal

import (
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

// Analyze inspect code in packages specified by AnalysisInfo. Called either via the main function
// or go's analyzer framework
func Analyze(info *AnalysisInfo) (err error) {
	for _, file := range info.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			block, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}

			if err = checkBlock(info, file, block); err != nil {
				return false
			}
			return true
		})
	}
	return
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

// checkBlock loops through all the statements in a block and reports findings if:
//  1. the statement assigns an error variable; and either
//  2. there is no following statement; or
//  3. the following statement is not an if statement checking the error
func checkBlock(info *AnalysisInfo, file *ast.File, block *ast.BlockStmt) error {
	stmts := block.List

	for i, stmt := range stmts {
		if !assignsErr(info, stmt) {
			continue
		}

		// no following statement
		if i+1 >= len(stmts) {
			report(info, file, stmt.Pos(), noFollowingStatement)
			continue
		}

		if !isErrCheck(info, stmts[i+1]) {
			report(info, file, stmts[i+1].Pos(), noNextStatmentNotCheck)
		}
	}
	return nil
}

// assignsErr returns true if a statement has an error variable assigned on the left-hand side
func assignsErr(info *AnalysisInfo, stmt ast.Stmt) bool {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok {
		return false
	}

	for _, lhs := range assign.Lhs {
		id, ok := lhs.(*ast.Ident)
		if !ok || id.Name == "_" {
			continue
		}

		if isErrorType(info.TypesInfo.TypeOf(id)) {
			return true
		}
	}

	return false
}

// isErrCheck returns true if a statement is an if that checks an error
func isErrCheck(info *AnalysisInfo, stmt ast.Stmt) bool {
	ifs, ok := stmt.(*ast.IfStmt)
	if !ok {
		return false
	}

	return conditionReferencesErr(info, ifs.Cond)
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
