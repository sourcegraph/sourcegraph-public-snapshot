// Package custom is similar to the WebAssembly Custom Sections. These are
// needed because `GOOS=js GOARCH=wasm` functions aren't defined naturally
// in WebAssembly. For example, every function has a single parameter "sp",
// which implicitly maps to stack parameters in this package.
package custom

const (
	// NamePadding is a marker for a parameter which has no purpose, except
	// padding. It should not be logged.
	NamePadding = "padding"
)

type Names struct {
	// Name is the WebAssembly function name.
	Name string

	// ParamNames are the parameters read in 8-byte strides from the stack
	// pointer (SP). This may be nil or include NamePadding.
	ParamNames []string

	// ResultNames are the results written in 8-byte strides from the stack
	// pointer (SP), after ParamNames.
	ResultNames []string
}

const (
	NameCallback = "callback"
	NameDebug    = "debug"
)

const (
	NameRuntimeWasmExit             = "runtime.wasmExit"
	NameRuntimeWasmWrite            = "runtime.wasmWrite"
	NameRuntimeResetMemoryDataView  = "runtime.resetMemoryDataView"
	NameRuntimeNanotime1            = "runtime.nanotime1"
	NameRuntimeWalltime             = "runtime.walltime"
	NameRuntimeScheduleTimeoutEvent = "runtime.scheduleTimeoutEvent"
	NameRuntimeClearTimeoutEvent    = "runtime.clearTimeoutEvent"
	NameRuntimeGetRandomData        = "runtime.getRandomData"
)

const (
	NameSyscallFinalizeRef        = "syscall/js.finalizeRef"
	NameSyscallStringVal          = "syscall/js.stringVal"
	NameSyscallValueGet           = "syscall/js.valueGet"
	NameSyscallValueSet           = "syscall/js.valueSet"
	NameSyscallValueDelete        = "syscall/js.valueDelete" // stubbed
	NameSyscallValueIndex         = "syscall/js.valueIndex"
	NameSyscallValueSetIndex      = "syscall/js.valueSetIndex" // stubbed
	NameSyscallValueCall          = "syscall/js.valueCall"
	NameSyscallValueInvoke        = "syscall/js.valueInvoke" // stubbed
	NameSyscallValueNew           = "syscall/js.valueNew"
	NameSyscallValueLength        = "syscall/js.valueLength"
	NameSyscallValuePrepareString = "syscall/js.valuePrepareString"
	NameSyscallValueLoadString    = "syscall/js.valueLoadString"
	NameSyscallValueInstanceOf    = "syscall/js.valueInstanceOf" // stubbed
	NameSyscallCopyBytesToGo      = "syscall/js.copyBytesToGo"
	NameSyscallCopyBytesToJS      = "syscall/js.copyBytesToJS"
)

var NameSection = map[string]*Names{
	NameDebug: {
		Name:        NameDebug,
		ParamNames:  []string{},
		ResultNames: []string{},
	},

	NameRuntimeWasmExit: {
		Name:        NameRuntimeWasmExit,
		ParamNames:  []string{"code"},
		ResultNames: []string{},
	},
	NameRuntimeWasmWrite: {
		Name:        NameRuntimeWasmWrite,
		ParamNames:  []string{"fd", "p", "p_len"},
		ResultNames: []string{},
	},
	NameRuntimeResetMemoryDataView: {
		Name:        NameRuntimeResetMemoryDataView,
		ParamNames:  []string{},
		ResultNames: []string{},
	},
	NameRuntimeNanotime1: {
		Name:        NameRuntimeNanotime1,
		ParamNames:  []string{},
		ResultNames: []string{"nsec"},
	},
	NameRuntimeWalltime: {
		Name:        NameRuntimeWalltime,
		ParamNames:  []string{},
		ResultNames: []string{"sec", "nsec"},
	},
	NameRuntimeScheduleTimeoutEvent: {
		Name:        NameRuntimeScheduleTimeoutEvent,
		ParamNames:  []string{"ms"},
		ResultNames: []string{"id"},
	},
	NameRuntimeClearTimeoutEvent: {
		Name:        NameRuntimeClearTimeoutEvent,
		ParamNames:  []string{"id"},
		ResultNames: []string{},
	},
	NameRuntimeGetRandomData: {
		Name:        NameRuntimeGetRandomData,
		ParamNames:  []string{"r", "r_len"},
		ResultNames: []string{},
	},

	NameSyscallFinalizeRef: {
		Name:        NameSyscallFinalizeRef,
		ParamNames:  []string{"r"},
		ResultNames: []string{},
	},
	NameSyscallStringVal: {
		Name:        NameSyscallStringVal,
		ParamNames:  []string{"x", "x_len"},
		ResultNames: []string{"r"},
	},
	NameSyscallValueGet: {
		Name:        NameSyscallValueGet,
		ParamNames:  []string{"v", "p", "p_len"},
		ResultNames: []string{"r"},
	},
	NameSyscallValueSet: {
		Name:        NameSyscallValueSet,
		ParamNames:  []string{"v", "p", "p_len", "x"},
		ResultNames: []string{},
	},
	NameSyscallValueDelete: {
		Name:        NameSyscallValueDelete,
		ParamNames:  []string{"v", "p", "p_len"},
		ResultNames: []string{},
	},
	NameSyscallValueIndex: {
		Name:        NameSyscallValueIndex,
		ParamNames:  []string{"v", "i"},
		ResultNames: []string{"r"},
	},
	NameSyscallValueSetIndex: {
		Name:        NameSyscallValueSetIndex,
		ParamNames:  []string{"v", "i", "x"},
		ResultNames: []string{},
	},
	NameSyscallValueCall: {
		Name:        NameSyscallValueCall,
		ParamNames:  []string{"v", "m", "m_len", "args", "args_len", NamePadding},
		ResultNames: []string{"res", "ok"},
	},
	NameSyscallValueInvoke: {
		Name:        NameSyscallValueInvoke,
		ParamNames:  []string{"v", "args", "args_len", NamePadding},
		ResultNames: []string{"res", "ok"},
	},
	NameSyscallValueNew: {
		Name:        NameSyscallValueNew,
		ParamNames:  []string{"v", "args", "args_len", NamePadding},
		ResultNames: []string{"res", "ok"},
	},
	NameSyscallValueLength: {
		Name:        NameSyscallValueLength,
		ParamNames:  []string{"v"},
		ResultNames: []string{"len"},
	},
	NameSyscallValuePrepareString: {
		Name:        NameSyscallValuePrepareString,
		ParamNames:  []string{"v"},
		ResultNames: []string{"str", "length"},
	},
	NameSyscallValueLoadString: {
		Name:        NameSyscallValueLoadString,
		ParamNames:  []string{"v", "b", "b_len"},
		ResultNames: []string{},
	},
	NameSyscallValueInstanceOf: {
		Name:        NameSyscallValueInstanceOf,
		ParamNames:  []string{"v", "t"},
		ResultNames: []string{"ok"},
	},
	NameSyscallCopyBytesToGo: {
		Name:        NameSyscallCopyBytesToGo,
		ParamNames:  []string{"dst", "dst_len", NamePadding, "src"},
		ResultNames: []string{"n", "ok"},
	},
	NameSyscallCopyBytesToJS: {
		Name:        NameSyscallCopyBytesToJS,
		ParamNames:  []string{"dst", "src", "src_len", NamePadding},
		ResultNames: []string{"n", "ok"},
	},
}

var NameSectionSyscallValueCall = map[string]map[string]*Names{
	NameCrypto:  CryptoNameSection,
	NameDate:    DateNameSection,
	NameFs:      FsNameSection,
	NameProcess: ProcessNameSection,
}
