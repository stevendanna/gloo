package translator

import (
	"github.com/solo-io/go-utils/errors"
)

var (
	VirtualServiceInWrongNamespace = func(vsName, vsNamespace string, selectFromNamespaces []string) error {
		return errors.Errorf("virtual service %s in namespace %s not included in selectFromNamespaces: %v", vsName, vsNamespace, selectFromNamespaces)
	}
)
