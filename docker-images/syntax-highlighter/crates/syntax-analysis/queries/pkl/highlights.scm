 ; Copyright © 2024 Apple Inc. and the Pkl project authors. All rights reserved.
 ;
 ; Licensed under the Apache License, Version 2.0 (the "License");
 ; you may not use this file except in compliance with the License.
 ; You may obtain a copy of the License at
 ;
 ;     https://www.apache.org/licenses/LICENSE-2.0
 ;
 ; Unless required by applicable law or agreed to in writing, software
 ; distributed under the License is distributed on an "AS IS" BASIS,
 ; WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 ; See the License for the specific language governing permissions and
 ; limitations under the License.
 ;
 ; Based on https://github.com/apple/tree-sitter-pkl/blob/main/queries/highlights.scm


; Types

(clazz (identifier) @type)
(typeAlias (identifier) @type)


(annotation ("@" @identifier.attribute) (qualifiedIdentifier (identifier) @identifier.attribute))

((identifier) @constant.builtin
 (#eq? @constant.builtin "Infinity"))

((identifier) @constant.builtin
 (#eq? @constant.builtin "NaN"))


((identifier) @type
 (#match? @type "^[A-Z]"))


(typeArgumentList
  "<" @punctuation.bracket
  ">" @punctuation.bracket)

; Method calls

(methodCallExpr
  (identifier) @function.method)

; Method definitions

(classMethod (methodHeader (identifier)) @function.method)
(objectMethod (methodHeader (identifier)) @function.method)

; Identifiers
(objectSpread ("..." @identifier.operator) (variableExpr))

(classProperty (identifier) @property)
(objectProperty (identifier) @property)

(parameterList (typedIdentifier (identifier) @variable.parameter))
(objectBodyParameters (typedIdentifier (identifier) @variable.parameter))

(identifier) @variable

; Literals

(stringConstant) @string
(slStringLiteral) @string
(mlStringLiteral) @string

(escapeSequence) @escape

(intLiteral) @number
(floatLiteral) @number

(interpolationExpr
  "\\(" @punctuation.special
  ")" @punctuation.special) @embedded

(interpolationExpr
 "\\#(" @punctuation.special
 ")" @punctuation.special) @embedded

(interpolationExpr
  "\\##(" @punctuation.special
  ")" @punctuation.special) @embedded

(lineComment) @comment
(blockComment) @comment
(docComment) @comment

; Operators

"??" @operator
"@"  @operator
"="  @operator
"<"  @operator
">"  @operator
"!"  @operator
"==" @operator
"!=" @operator
"<=" @operator
">=" @operator
"&&" @operator
"||" @operator
"+"  @operator
"-"  @operator
"**" @operator
"*"  @operator
"/"  @operator
"~/" @operator
"%"  @operator
"|>" @operator

"?"  @operator.type
"|"  @operator.type
"->" @operator.type

"," @punctuation.delimiter
":" @punctuation.delimiter
"." @punctuation.delimiter
"?." @punctuation.delimiter

"(" @punctuation.bracket
")" @punctuation.bracket
"[" @punctuation.bracket
"]" @punctuation.bracket
"{" @punctuation.bracket
"}" @punctuation.bracket

; Keywords

"abstract" @keyword
"amends" @include
"as" @keyword
"class" @keyword
"const" @keyword
"else" @keyword
"extends" @keyword
"external" @keyword
(falseLiteral) @boolean
"fixed" @keyword
"for" @keyword
"function" @keyword.function
"hidden" @keyword
"if" @keyword
(importExpr "import" @include)
(importGlobExpr "import*" @include)
"import" @include
"import*" @include
"in" @keyword
"is" @keyword
"let" @keyword
"local" @keyword
(moduleExpr "module" @include)
"module" @keyword
"new" @keyword
"nothing" @type.builtin
(nullLiteral) @constant.null
"open" @keyword
"out" @keyword
(outerExpr) @variable.builtin
"read" @function.method.builtin
"read?" @function.method.builtin
"read*" @function.method.builtin
"super" @variable.builtin
(thisExpr) @variable.builtin
"throw" @function.method.builtin
"trace" @function.method.builtin
(trueLiteral) @boolean
"typealias" @keyword
"unknown" @type.builtin
"when" @keyword
