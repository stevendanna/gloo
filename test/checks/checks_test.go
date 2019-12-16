package checks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/utils/gomodutils"
	"golang.org/x/mod/modfile"
)

var _ = Describe("Checks", func() {

	It("should used forked klog instead of klog", func() {
		// regular klog writes to disk, so make sure we used a forked version that doesn't write to
		// disk, which is a problem with hardened containers with root only file systems.

		modParser := &gomodutils.ModParser{}
		modFile, err := modParser.Parse()
		Expect(err).NotTo(HaveOccurred())

		var klogReplace *modfile.Replace
		for _, replace := range modFile.Replace {
			if replace.Old.Path == "k8s.io/klog" {
				klogReplace = replace
				break
			}
		}
		Expect(klogReplace).NotTo(BeNil())
		Expect(klogReplace.New.Path).To(Equal("github.com/stefanprodan/klog"))
		Expect(klogReplace.New.Version).To(Equal("v0.0.0-20190418165334-9cbb78b20423"))
	})

})
