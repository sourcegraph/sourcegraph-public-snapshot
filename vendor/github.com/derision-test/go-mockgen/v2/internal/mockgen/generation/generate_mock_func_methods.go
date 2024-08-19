package generation

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

func generateMockFuncSetHookMethod(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	commentText := fmt.Sprintf(
		`SetDefaultHook sets function that is called when the %s method of the parent %s instance is invoked and the hook queue is empty.`,
		method.Name,
		iface.mockStructName,
	)

	assignStatement := jen.Id("f").Dot("defaultHook").Op("=").Id("hook")

	params := []jen.Code{compose(jen.Id("hook"), method.signature)}
	return generateMockFuncMethod(iface, outputImportPath, method, "SetDefaultHook", commentText, params, nil,
		assignStatement, // f.defaultHook = hook
	)
}

func generateMockFuncPushHookMethod(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	commentText := strings.Join([]string{
		`PushHook adds a function to the end of hook queue.`,
		fmt.Sprintf(`Each invocation of the %s method of the parent %s instance invokes the hook at the front of the queue and discards it.`, method.Name, iface.mockStructName),
		`After the queue is empty, the default hook function is invoked for any future action.`,
	}, " ")

	lockStatement := jen.Id("f").Dot("mutex").Dot("Lock").Call()
	unlockStatement := jen.Id("f").Dot("mutex").Dot("Unlock").Call()
	appendStatement := selfAppend(jen.Id("f").Dot("hooks"), jen.Id("hook"))

	params := []jen.Code{compose(jen.Id("hook"), method.signature)}
	return generateMockFuncMethod(iface, outputImportPath, method, "PushHook", commentText, params, nil,
		lockStatement,   // f.mutex.Lock()
		appendStatement, // f.mutex.Unlock()
		unlockStatement, // f.hooks = append(f.hooks, hook)
	)
}

func generateMockFuncSetReturnMethod(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	return generateMockReturnMethod(iface, method, "SetDefault", outputImportPath)
}

func generateMockFuncPushReturnMethod(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	return generateMockReturnMethod(iface, method, "Push", outputImportPath)
}

func generateMockReturnMethod(iface *wrappedInterface, method *wrappedMethod, methodPrefix, outputImportPath string) jen.Code {
	commentText := fmt.Sprintf(
		`%sReturn calls %sHook with a function that returns the given values.`,
		methodPrefix,
		methodPrefix,
	)

	names := make([]jen.Code, 0, len(method.resultTypes))
	params := make([]jen.Code, 0, len(method.resultTypes))
	for i, typ := range method.resultTypes {
		name := jen.Id(fmt.Sprintf("r%d", i))
		names = append(names, name)
		params = append(params, compose(name, typ))
	}

	returnStatement := jen.Return().List(names...)
	functionExpression := jen.Func().Params(method.paramTypes...).Params(method.resultTypes...).Block(returnStatement)
	callStatement := jen.Id("f").Dot(fmt.Sprintf("%sHook", methodPrefix)).Call(functionExpression)

	return generateMockFuncMethod(iface, outputImportPath, method, fmt.Sprintf("%sReturn", methodPrefix), commentText, params, nil,
		callStatement, // f.<SetDefault|Push>Hook(func( T<n>, ... ) { return r<n>, ... })
	)
}

func generateMockFuncNextHookMethod(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	lockStatement := jen.Id("f").Dot("mutex").Dot("Lock").Call()
	deferUnlockStatement := jen.Defer().Id("f").Dot("mutex").Dot("Unlock").Call()
	lenHooksExpression := jen.Len(jen.Id("f").Dot("hooks"))
	earlyReturnStatement := jen.Return(jen.Id("f").Dot("defaultHook"))
	returnDefaultIfEmptyCondition := jen.If(lenHooksExpression.Op("==").Lit(0)).Block(earlyReturnStatement)
	firstHookStatement := jen.Id("hook").Op(":=").Id("f").Dot("hooks").Index(jen.Lit(0))
	popHookStatement := jen.Id("f").Dot("hooks").Op("=").Id("f").Dot("hooks").Index(jen.Lit(1).Op(":"))
	returnStatement := jen.Return(jen.Id("hook"))

	results := []jen.Code{method.signature}
	return generateMockFuncMethod(iface, outputImportPath, method, "nextHook", "", nil, results,
		lockStatement,                    // f.mutex.Lock()
		deferUnlockStatement, jen.Line(), // defer f.mutex.Unlock()
		returnDefaultIfEmptyCondition, jen.Line(), // if len(f.hooks) == 0 { return f.defaultHook }
		firstHookStatement, // hook := f.hooks[0]
		popHookStatement,   // f.hooks = f.hooks[1:]
		returnStatement,    // return hook
	)
}

func generateMockFuncAppendCallMethod(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	mockFuncCallStructName := fmt.Sprintf("%s%s%sFuncCall", iface.prefix, iface.titleName, method.Name)

	lockStatement := jen.Id("f").Dot("mutex").Dot("Lock").Call()
	unlockStatement := jen.Id("f").Dot("mutex").Dot("Unlock").Call()
	appendStatement := selfAppend(jen.Id("f").Dot("history"), jen.Id("r0"))

	params := []jen.Code{compose(jen.Id("r0"), addTypes(jen.Id(mockFuncCallStructName), iface.TypeParams, outputImportPath, false))}
	return generateMockFuncMethod(iface, outputImportPath, method, "appendCall", "", params, nil,
		lockStatement,   // f.mutex.Lock()
		appendStatement, // f.history = append(f.history, r0)
		unlockStatement, // f.mutex.Unlock()
	)
}

func generateMockFuncHistoryMethod(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	mockFuncCallStructName := fmt.Sprintf("%s%s%sFuncCall", iface.prefix, iface.titleName, method.Name)
	commentText := fmt.Sprintf(
		`History returns a sequence of %s objects describing the invocations of this function.`,
		mockFuncCallStructName,
	)

	lockStatement := jen.Id("f").Dot("mutex").Dot("Lock").Call()
	unlockStatement := jen.Id("f").Dot("mutex").Dot("Unlock").Call()
	callStructSliceType := compose(jen.Index(), addTypes(jen.Id(mockFuncCallStructName), iface.TypeParams, outputImportPath, false))
	lenHistoryExpression := jen.Len(jen.Id("f").Dot("history"))
	makeSliceStatement := jen.Id("history").Op(":=").Make(callStructSliceType, lenHistoryExpression)
	copyStatement := jen.Copy(jen.Id("history"), jen.Id("f").Dot("history"))
	returnStatement := jen.Return().Id("history")

	results := []jen.Code{compose(jen.Index(), addTypes(jen.Id(mockFuncCallStructName), iface.TypeParams, outputImportPath, false))}
	return generateMockFuncMethod(iface, outputImportPath, method, "History", commentText, nil, results,
		lockStatement,               // f.mutex.Lock()
		makeSliceStatement,          // history := make([]<callStructName>, len(f.history))
		copyStatement,               // copy(history, f.history)
		unlockStatement, jen.Line(), // f.mutex.Unlock()
		returnStatement, // return history
	)
}

func generateMockFuncMethod(
	iface *wrappedInterface,
	outputImportPath string,
	method *wrappedMethod,
	methodName string,
	commentText string,
	params, results []jen.Code,
	body ...jen.Code,
) jen.Code {
	mockFuncStructName := fmt.Sprintf("%s%s%sFunc", iface.prefix, iface.titleName, method.Name)
	receiver := compose(jen.Id("f").Op("*"), addTypes(jen.Id(mockFuncStructName), iface.TypeParams, outputImportPath, false))
	methodDeclaration := jen.Func().Params(receiver).Id(methodName).Params(params...).Params(results...).Block(body...)
	return addComment(methodDeclaration, 1, commentText)
}
