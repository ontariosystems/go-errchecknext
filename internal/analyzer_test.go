package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/ontariosystems/go-errchecknext/internal"
	"golang.org/x/tools/go/analysis/analysistest"
)

var _ = Describe("errchecknext", func() {
	It("reports statements between assignment and error check", func() {
		analysistest.Run(
			GinkgoTB(),
			analysistest.TestData(),
			internal.Analyzer,
		)
	})
})
