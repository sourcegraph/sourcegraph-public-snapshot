package generation

import "github.com/derision-test/go-mockgen/v2/internal/mockgen/types"

type wrappedInterface struct {
	*types.Interface
	prefix         string
	titleName      string
	mockStructName string
	wrappedMethods []*wrappedMethod
}

func wrapInterface(iface *types.Interface, prefix, titleName, mockStructName, outputImportPath string) *wrappedInterface {
	wrapped := &wrappedInterface{
		Interface:      iface,
		prefix:         prefix,
		titleName:      titleName,
		mockStructName: mockStructName,
	}

	for _, method := range iface.Methods {
		wrapped.wrappedMethods = append(wrapped.wrappedMethods, wrapMethod(iface, method, outputImportPath))
	}

	return wrapped
}
