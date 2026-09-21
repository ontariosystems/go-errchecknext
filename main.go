package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"golang.org/x/tools/go/packages"

	"github.com/ontariosystems/go-errchecknext/internal"
)

func main() {
	var outputFile string

	flag.StringVar(&outputFile, "o", "", "checkstyle output file")
	flag.Parse()

	patterns := flag.Args()
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "package load failed: %v\n", err)
		os.Exit(1)
	}

	collector := internal.NewCollector()
	wg := &sync.WaitGroup{}
	for _, pkg := range pkgs {
		wg.Go(func() {
			if err := internal.NewAnalysisContextFromPkg(pkg).WithCollector(collector).Analyze(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", pkg.PkgPath, err)
			}
		})
	}
	wg.Wait()

	collector.Print(os.Stderr)

	if outputFile != "" {
		if err := collector.WriteCheckstyle(outputFile); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "write checkstyle: %v\n", err)
			os.Exit(99)
		}
	}

	if !collector.HasErrors() {
		_, _ = fmt.Fprintf(os.Stderr, "0 issues.\n")
		os.Exit(0)
	}

	// exit 1 if errors
	os.Exit(1)
}
