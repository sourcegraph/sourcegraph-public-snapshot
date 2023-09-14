;;; Annotations

(annotation
	"@" @identifier.attribute (use_site_target)? @identifier.attribute)
(annotation
	(user_type
		(type_identifier) @identifier.attribute))
(annotation
	(constructor_invocation
		(user_type
			(type_identifier) @identifier.attribute)))

(file_annotation
	"@" @identifier.attribute "file" @identifier.attribute ":" @identifier.attribute)
(file_annotation
	(user_type
		(type_identifier) @identifier.attribute))
(file_annotation
	(constructor_invocation
		(user_type
			(type_identifier) @identifier.attribute)))

;;; Comments

[
  (line_comment)
  (multiline_comment)
] @comment

(shebang_line) @comment

;;; Function definitions

(function_declaration
  (simple_identifier) @identifier.function)

(getter
	("get") @function.builtin)
(setter
	("set") @function.builtin)

(anonymous_initializer
	("init") @function.builtin)

(secondary_constructor
	("constructor") @function.builtin)

(parameter
	(simple_identifier) @identifier.parameter)

(parameter_with_optional_type
	(simple_identifier) @identifier.parameter)

(lambda_literal
	(lambda_parameters
		(variable_declaration
			(simple_identifier) @identifier.parameter)))

;;; Function calls

(call_expression
	. (simple_identifier) @function.builtin
    (#match? @function.builtin "^(arrayOf|arrayOfNulls|byteArrayOf|shortArrayOf|intArrayOf|longArrayOf|ubyteArrayOf|ushortArrayOf|uintArrayOf|ulongArrayOf|floatArrayOf|doubleArrayOf|booleanArrayOf|charArrayOf|emptyArray|mapOf|setOf|listOf|emptyMap|emptySet|emptyList|mutableMapOf|mutableSetOf|mutableListOf|print|println|error|TODO|run|runCatching|repeat|lazy|lazyOf|enumValues|enumValueOf|assert|check|checkNotNull|require|requireNotNull|with|suspend|synchronized)$"
))

; function()
(call_expression
	. (simple_identifier) @identifier.function)

; ::function
(callable_reference
	. (simple_identifier) @identifier.function)

; object.function() or object.property.function()
(call_expression
	(navigation_expression
		(navigation_suffix
			(simple_identifier) @identifier.function) . ))

;;; Identifiers

; Basics
((simple_identifier) @constant
  (#match? @constant "^[A-Z][A-Z0-9_]*$"))

(simple_identifier) @identifier
(interpolated_identifier) @identifier

; `this` keyword inside classes
(this_expression) @identifier.builtin

; `super` keyword inside classes
(super_expression) @identifier.builtin

(enum_entry
	(simple_identifier) @constant)

; Types
((type_identifier) @type.builtin
	(#match? @type.builtin "^(Byte|Short|Int|Long|UByte|UShort|UInt|ULong|Float|Double|Boolean|Char|String|Array|ByteArray|ShortArray|IntArray|LongArray|UByteArray|UShortArray|UIntArray|ULongArray|FloatArray|DoubleArray|BooleanArray|CharArray|Map|Set|List|EmptyMap|EmptySet|EmptyList|MutableMap|MutableSet|MutableList)$"
))

(type_identifier) @identifier.type

;;; Keywords

[
 "import"
 "package"]
@include

"fun" @keyword.function
(jump_expression ("return") @keyword.return)

[
	"if"
	"else"
	"when"
] @conditional

(label) @keyword

[
    "break"
	"catch"
	"class"
	"companion"
	"continue"
	"do"
	"enum"
    "finally"
	"for"
	"interface"
	"object"
	"suspend"
	"throw"
	"try"
	"typealias"
	"while"
    "val"
    "var"
] @keyword

[
	(class_modifier)
	(member_modifier)
	(function_modifier)
	(property_modifier)
	(platform_modifier)
	(variance_modifier)
	(parameter_modifier)
	(visibility_modifier)
	(reification_modifier)
	(inheritance_modifier)
] @keyword

;;; Literals

(real_literal) @float
[
	(integer_literal)
	(long_literal)
	(hex_literal)
	(bin_literal)
	(unsigned_literal)
] @number

(boolean_literal) @boolean
"null" @constant.null

(string_literal) @string
(string_literal
	["$" "${" "}"] @string.escape
)

(character_literal) @character
(character_literal (character_escape_seq) @string.escape)

;;; Operators and Punctuation

[
	"!"
	"!="
	"!=="
	"="
	"=="
	"==="
	">"
	">="
	"<"
	"<="
	"||"
	"&&"
	"+"
	"++"
	"+="
	"-"
	"--"
	"-="
	"*"
	"*="
	"/"
	"/="
	"%"
	"%="
	"?."
	"?:"
	"!!"
	"is"
	"!is"
	"in"
	"!in"
	"as"
	"as?"
	".."
	"->"
] @operator

; '?' operator, replacement for Java @Nullable
(nullable_type) @punctuation

[
	"(" ")"
	"[" "]"
	"{" "}"
] @punctuation.bracket

[
	"."
	","
	";"
	":"
	"::"
] @punctuation.delimiter
