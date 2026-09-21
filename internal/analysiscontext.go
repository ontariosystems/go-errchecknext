package internal

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

// AnalysisContext information about the files being scanned, the types from the ast, and reporting
type AnalysisContext struct {
	Files     []*ast.File
	Fset      *token.FileSet
	TypesInfo *types.Info
	Collector *Collector
	Report    func(analysis.Diagnostic)
}

// NewAnalysisContextFromPkg creates an AnalysisContext from the information in a packages.Package
func NewAnalysisContextFromPkg(pkg *packages.Package) *AnalysisContext {
	return &AnalysisContext{
		Files:     pkg.Syntax,
		Fset:      pkg.Fset,
		TypesInfo: pkg.TypesInfo,
	}
}

// NewAnalysisContextFromPass creates an AnalysisContext from the information in an analysis.Pass
func NewAnalysisContextFromPass(pass *analysis.Pass) *AnalysisContext {
	return &AnalysisContext{
		Files:     pass.Files,
		Fset:      pass.Fset,
		TypesInfo: pass.TypesInfo,
		Report:    pass.Report,
	}
}

// WithCollector adds a Collector to this AnalysisContext
func (aCtx *AnalysisContext) WithCollector(collector *Collector) *AnalysisContext {
	aCtx.Collector = collector
	return aCtx
}

// Analyze inspect code in packages specified by AnalysisContext. Called either via the main function
// or go's analyzer framework
func (aCtx *AnalysisContext) Analyze() (err error) {
	for _, file := range aCtx.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.FuncDecl:
				NewFunctionContext(aCtx, file, n).AnalyzeFunction()
				return false
			}
			return true
		})
	}
	return
}
