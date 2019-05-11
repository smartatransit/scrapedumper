package martaapi_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMartaapi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Martaapi Suite")
}
