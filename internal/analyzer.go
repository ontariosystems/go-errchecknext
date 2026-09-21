package internal

import (
	"golang.org/x/tools/go/analysis"
)

const name = "errchecknext"

var (
	Analyzer = &analysis.Analyzer{
		Name: name,
		Doc:  "require an immediate error check after assigning to err",
		Run:  run,
	}
)

// run the main run method of the go analyzer framework, calls our Analyze method so we can use the
// same logic from our main function
func run(pass *analysis.Pass) (any, error) {
	if err := NewAnalysisContextFromPass(pass).Analyze(); err != nil {
		return nil, err
	}
	return nil, nil
}
