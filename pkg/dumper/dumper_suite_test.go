package dumper_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDumper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dumper Suite")
}
