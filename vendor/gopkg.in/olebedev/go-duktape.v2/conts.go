package duktape

const (
	CompileEval uint = 1 << iota
	CompileFunction
	CompileStrict
	CompileSafe
	CompileNoResult
	CompileNoSource
	CompileStrlen
)

const (
	TypeNone Type = iota
	TypeUndefined
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeObject
	TypeBuffer
	TypePointer
	TypeLightFunc
)

const (
	TypeMaskNone uint = 1 << iota
	TypeMaskUndefined
	TypeMaskNull
	TypeMaskBoolean
	TypeMaskNumber
	TypeMaskString
	TypeMaskObject
	TypeMaskBuffer
	TypeMaskPointer
	TypeMaskLightFunc
)

const (
	EnumIncludeNonenumerable uint = 1 << iota
	EnumIncludeInternal
	EnumOwnPropertiesOnly
	EnumArrayIndicesOnly
	EnumSortArrayIndices
	NoProxyBehavior
)

const (
	ErrNone int = 0

	// Internal to Duktape
	ErrUnimplemented int = 50 + iota
	ErrUnsupported
	ErrInternal
	ErrAlloc
	ErrAssertion
	ErrAPI
	ErrUncaughtError
)

const (
	// Common prototypes
	ErrError int = 100 + iota
	ErrEval
	ErrRange
	ErrReference
	ErrSyntax
	ErrType
	ErrURI
)

const (
	// Returned error values
	ErrRetUnimplemented int = -(ErrUnimplemented + iota)
	ErrRetUnsupported
	ErrRetInternal
	ErrRetAlloc
	ErrRetAssertion
	ErrRetAPI
	ErrRetUncaughtError
)

const (
	ErrRetError int = -(ErrError + iota)
	ErrRetEval
	ErrRetRange
	ErrRetReference
	ErrRetSyntax
	ErrRetType
	ErrRetURI
)

const (
	ExecSuccess = iota
	ExecError
)

const (
	LogTrace int = iota
	LogDebug
	LogInfo
	LogWarn
	LogError
	LogFatal
)

const (
	// Keep it sync with duktape.h:555
	BufobjCreateArrbuf      = 1 << 4 // internal flag: create backing arraybuffer; keep in one byte
	BufobjDuktapeAuffer     = 0
	BufobjNodejsAuffer      = 1
	BufobjArraybuffer       = 2
	BufobjDataview          = 3 | BufobjCreateArrbuf
	BufobjInt8array         = 4 | BufobjCreateArrbuf
	BufobjUint8array        = 5 | BufobjCreateArrbuf
	BufobjUint8clampedarray = 6 | BufobjCreateArrbuf
	BufobjInt16array        = 7 | BufobjCreateArrbuf
	BufobjUint16array       = 8 | BufobjCreateArrbuf
	BufobjInt32array        = 9 | BufobjCreateArrbuf
	BufobjUint32array       = 10 | BufobjCreateArrbuf
	BufobjFloat32array      = 11 | BufobjCreateArrbuf
	BufobjFloat64array      = 12 | BufobjCreateArrbuf
)
