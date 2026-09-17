package internal

import (
	"bytes"
	"go/token"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collector", func() {
	var collector *Collector

	BeforeEach(func() {
		collector = NewCollector()
	})

	It("starts empty", func() {
		Expect(collector).NotTo(BeNil())
		Expect(collector.HasErrors()).To(BeFalse())
	})

	It("records findings", func() {
		fset := token.NewFileSet()

		collector.ReportFinding(
			fset,
			posAt(fset, 1, 3),
			"test finding",
		)
		collector.ReportFinding(
			fset,
			posAt(fset, 3, 0),
			"another finding",
		)

		Expect(collector.HasErrors()).To(BeTrue())
		Expect(collector.findings).To(HaveLen(2))

		f := collector.findings[0]
		Expect(f.filename).To(Equal(testFilename))
		Expect(f.message).To(Equal("test finding"))

		f = collector.findings[1]
		Expect(f.filename).To(Equal(testFilename))
		Expect(f.message).To(Equal("another finding"))
	})

	It("prints findings", func() {
		fset := token.NewFileSet()

		collector.ReportFinding(
			fset,
			posAt(fset, 1, 3),
			"something bad",
		)

		var buf bytes.Buffer
		collector.Print(&buf)

		Expect(buf.String()).To(ContainSubstring(testFilename))
		Expect(buf.String()).To(ContainSubstring("something bad"))
		Expect(buf.String()).To(ContainSubstring("1:3"))
	})

	It("prints nothing when empty", func() {
		var buf bytes.Buffer
		collector.Print(&buf)

		Expect(buf.String()).To(BeEmpty())
	})

	It("writes a checkstyle report", func() {
		fset := token.NewFileSet()

		collector.ReportFinding(
			fset,
			posAt(fset, 1, 3),
			"problem",
		)

		output := filepath.Join(
			GinkgoT().TempDir(),
			"report.xml",
		)

		Expect(collector.WriteCheckstyle(output)).To(Succeed())

		content, err := os.ReadFile(output)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(content)).To(ContainSubstring("checkstyle"))
		Expect(string(content)).To(ContainSubstring("problem"))
		Expect(string(content)).To(ContainSubstring(`line="1"`))
		Expect(string(content)).To(ContainSubstring(`column="3"`))
	})
})

const (
	testFilename = "test.go"
	testLines    = `package test

func main() {
}
`
)

// posAt returns a position for a line and column in the test file. This can exceed the length of the file, which is
// weird but really doesn't matter for the tests.
func posAt(fset *token.FileSet, line, col int) token.Pos {
	file := fset.AddFile(testFilename, -1, len(testLines))
	file.SetLinesForContent([]byte(testLines))
	start := file.LineStart(line)
	return start + token.Pos(col-1)
}
