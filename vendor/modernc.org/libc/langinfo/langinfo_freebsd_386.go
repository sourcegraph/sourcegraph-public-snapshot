// Code generated by 'ccgo langinfo/gen.c -crt-import-path "" -export-defines "" -export-enums "" -export-externs X -export-fields F -export-structs "" -export-typedefs "" -header -hide _OSSwapInt16,_OSSwapInt32,_OSSwapInt64 -ignore-unsupported-alignment -o langinfo/langinfo_freebsd_386.go -pkgname langinfo', DO NOT EDIT.

package langinfo

import (
	"math"
	"reflect"
	"sync/atomic"
	"unsafe"
)

var _ = math.Pi
var _ reflect.Kind
var _ atomic.Value
var _ unsafe.Pointer

const (
	ABDAY_1              = 14 // langinfo.h:60:1:
	ABDAY_2              = 15 // langinfo.h:61:1:
	ABDAY_3              = 16 // langinfo.h:62:1:
	ABDAY_4              = 17 // langinfo.h:63:1:
	ABDAY_5              = 18 // langinfo.h:64:1:
	ABDAY_6              = 19 // langinfo.h:65:1:
	ABDAY_7              = 20 // langinfo.h:66:1:
	ABMON_1              = 33 // langinfo.h:83:1:
	ABMON_10             = 42 // langinfo.h:92:1:
	ABMON_11             = 43 // langinfo.h:93:1:
	ABMON_12             = 44 // langinfo.h:94:1:
	ABMON_2              = 34 // langinfo.h:84:1:
	ABMON_3              = 35 // langinfo.h:85:1:
	ABMON_4              = 36 // langinfo.h:86:1:
	ABMON_5              = 37 // langinfo.h:87:1:
	ABMON_6              = 38 // langinfo.h:88:1:
	ABMON_7              = 39 // langinfo.h:89:1:
	ABMON_8              = 40 // langinfo.h:90:1:
	ABMON_9              = 41 // langinfo.h:91:1:
	ALTMON_1             = 58 // langinfo.h:120:1:
	ALTMON_10            = 67 // langinfo.h:129:1:
	ALTMON_11            = 68 // langinfo.h:130:1:
	ALTMON_12            = 69 // langinfo.h:131:1:
	ALTMON_2             = 59 // langinfo.h:121:1:
	ALTMON_3             = 60 // langinfo.h:122:1:
	ALTMON_4             = 61 // langinfo.h:123:1:
	ALTMON_5             = 62 // langinfo.h:124:1:
	ALTMON_6             = 63 // langinfo.h:125:1:
	ALTMON_7             = 64 // langinfo.h:126:1:
	ALTMON_8             = 65 // langinfo.h:127:1:
	ALTMON_9             = 66 // langinfo.h:128:1:
	ALT_DIGITS           = 49 // langinfo.h:100:1:
	AM_STR               = 5  // langinfo.h:47:1:
	CODESET              = 0  // langinfo.h:42:1:
	CRNCYSTR             = 56 // langinfo.h:113:1:
	DAY_1                = 7  // langinfo.h:51:1:
	DAY_2                = 8  // langinfo.h:52:1:
	DAY_3                = 9  // langinfo.h:53:1:
	DAY_4                = 10 // langinfo.h:54:1:
	DAY_5                = 11 // langinfo.h:55:1:
	DAY_6                = 12 // langinfo.h:56:1:
	DAY_7                = 13 // langinfo.h:57:1:
	D_FMT                = 2  // langinfo.h:44:1:
	D_MD_ORDER           = 57 // langinfo.h:116:1:
	D_T_FMT              = 1  // langinfo.h:43:1:
	ERA                  = 45 // langinfo.h:96:1:
	ERA_D_FMT            = 46 // langinfo.h:97:1:
	ERA_D_T_FMT          = 47 // langinfo.h:98:1:
	ERA_T_FMT            = 48 // langinfo.h:99:1:
	MON_1                = 21 // langinfo.h:69:1:
	MON_10               = 30 // langinfo.h:78:1:
	MON_11               = 31 // langinfo.h:79:1:
	MON_12               = 32 // langinfo.h:80:1:
	MON_2                = 22 // langinfo.h:70:1:
	MON_3                = 23 // langinfo.h:71:1:
	MON_4                = 24 // langinfo.h:72:1:
	MON_5                = 25 // langinfo.h:73:1:
	MON_6                = 26 // langinfo.h:74:1:
	MON_7                = 27 // langinfo.h:75:1:
	MON_8                = 28 // langinfo.h:76:1:
	MON_9                = 29 // langinfo.h:77:1:
	NOEXPR               = 53 // langinfo.h:106:1:
	NOSTR                = 55 // langinfo.h:110:1:
	PM_STR               = 6  // langinfo.h:48:1:
	RADIXCHAR            = 50 // langinfo.h:102:1:
	THOUSEP              = 51 // langinfo.h:103:1:
	T_FMT                = 3  // langinfo.h:45:1:
	T_FMT_AMPM           = 4  // langinfo.h:46:1:
	YESEXPR              = 52 // langinfo.h:105:1:
	YESSTR               = 54 // langinfo.h:109:1:
	X_FILE_OFFSET_BITS   = 64 // <builtin>:25:1:
	X_ILP32              = 1  // <predefined>:1:1:
	X_LANGINFO_H_        = 0  // langinfo.h:32:1:
	X_LOCALE_T_DEFINED   = 0  // _langinfo.h:37:1:
	X_MACHINE__LIMITS_H_ = 0  // _limits.h:36:1:
	X_MACHINE__TYPES_H_  = 0  // _types.h:42:1:
	X_NL_ITEM_DECLARED   = 0  // langinfo.h:39:1:
	X_Nonnull            = 0  // cdefs.h:790:1:
	X_Null_unspecified   = 0  // cdefs.h:792:1:
	X_Nullable           = 0  // cdefs.h:791:1:
	X_SYS_CDEFS_H_       = 0  // cdefs.h:39:1:
	X_SYS__TYPES_H_      = 0  // _types.h:32:1:
	X_XLOCALE_LANGINFO_H = 0  // _langinfo.h:34:1:
	I386                 = 1  // <predefined>:335:1:
	Unix                 = 1  // <predefined>:336:1:
)

type Ptrdiff_t = int32 /* <builtin>:3:26 */

type Size_t = uint32 /* <builtin>:9:23 */

type Wchar_t = int32 /* <builtin>:15:24 */

type X__builtin_va_list = uintptr /* <builtin>:46:14 */
type X__float128 = float64        /* <builtin>:47:21 */

// -
// SPDX-License-Identifier: BSD-2-Clause-FreeBSD
//
// Copyright (c) 2001 Alexey Zelkin <phantom@FreeBSD.org>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHOR AND CONTRIBUTORS ``AS IS'' AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED.  IN NO EVENT SHALL THE AUTHOR OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
// OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
// LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
// OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// SUCH DAMAGE.
//
// $FreeBSD$

// -
// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 1991, 1993
//	The Regents of the University of California.  All rights reserved.
//
// This code is derived from software contributed to Berkeley by
// Berkeley Software Design, Inc.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the University nor the names of its contributors
//    may be used to endorse or promote products derived from this software
//    without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS ``AS IS'' AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED.  IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
// OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
// LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
// OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// SUCH DAMAGE.
//
//	@(#)cdefs.h	8.8 (Berkeley) 1/9/95
// $FreeBSD$

// Testing against Clang-specific extensions.

// This code has been put in place to help reduce the addition of
// compiler specific defines in FreeBSD code.  It helps to aid in
// having a compiler-agnostic source tree.

// Compiler memory barriers, specific to gcc and clang.

// XXX: if __GNUC__ >= 2: not tested everywhere originally, where replaced

// Macro to test if we're using a specific version of gcc or later.

// The __CONCAT macro is used to concatenate parts of symbol names, e.g.
// with "#define OLD(foo) __CONCAT(old,foo)", OLD(foo) produces oldfoo.
// The __CONCAT macro is a bit tricky to use if it must work in non-ANSI
// mode -- there must be no spaces between its arguments, and for nested
// __CONCAT's, all the __CONCAT's must be at the left.  __CONCAT can also
// concatenate double-quoted strings produced by the __STRING macro, but
// this only works with ANSI C.
//
// __XSTRING is like __STRING, but it expands any macros in its argument
// first.  It is only available with ANSI C.

// Compiler-dependent macros to help declare dead (non-returning) and
// pure (no side effects) functions, and unused variables.  They are
// null except for versions of gcc that are known to support the features
// properly (old versions of gcc-2 supported the dead and pure features
// in a different (wrong) way).  If we do not provide an implementation
// for a given compiler, let the compile fail if it is told to use
// a feature that we cannot live without.

// Keywords added in C11.

// Emulation of C11 _Generic().  Unlike the previously defined C11
// keywords, it is not possible to implement this using exactly the same
// syntax.  Therefore implement something similar under the name
// __generic().  Unlike _Generic(), this macro can only distinguish
// between a single type, so it requires nested invocations to
// distinguish multiple cases.

// C99 Static array indices in function parameter declarations.  Syntax such as:
// void bar(int myArray[static 10]);
// is allowed in C99 but not in C++.  Define __min_size appropriately so
// headers using it can be compiled in either language.  Use like this:
// void bar(int myArray[__min_size(10)]);

// XXX: should use `#if __STDC_VERSION__ < 199901'.

// C++11 exposes a load of C99 stuff

// GCC 2.95 provides `__restrict' as an extension to C90 to support the
// C99-specific `restrict' type qualifier.  We happen to use `__restrict' as
// a way to define the `restrict' type qualifier without disturbing older
// software that is unaware of C99 keywords.

// GNU C version 2.96 adds explicit branch prediction so that
// the CPU back-end can hint the processor and also so that
// code blocks can be reordered such that the predicted path
// sees a more linear flow, thus improving cache behavior, etc.
//
// The following two macros provide us with a way to utilize this
// compiler feature.  Use __predict_true() if you expect the expression
// to evaluate to true, and __predict_false() if you expect the
// expression to evaluate to false.
//
// A few notes about usage:
//
//	* Generally, __predict_false() error condition checks (unless
//	  you have some _strong_ reason to do otherwise, in which case
//	  document it), and/or __predict_true() `no-error' condition
//	  checks, assuming you want to optimize for the no-error case.
//
//	* Other than that, if you don't know the likelihood of a test
//	  succeeding from empirical or other `hard' evidence, don't
//	  make predictions.
//
//	* These are meant to be used in places that are run `a lot'.
//	  It is wasteful to make predictions in code that is run
//	  seldomly (e.g. at subsystem initialization time) as the
//	  basic block reordering that this affects can often generate
//	  larger code.

// We define this here since <stddef.h>, <sys/queue.h>, and <sys/types.h>
// require it.

// Given the pointer x to the member m of the struct s, return
// a pointer to the containing structure.  When using GCC, we first
// assign pointer x to a local variable, to check that its type is
// compatible with member m.

// Compiler-dependent macros to declare that functions take printf-like
// or scanf-like arguments.  They are null except for versions of gcc
// that are known to support the features properly (old versions of gcc-2
// didn't permit keeping the keywords out of the application namespace).

// Compiler-dependent macros that rely on FreeBSD-specific extensions.

// Embed the rcs id of a source file in the resulting library.  Note that in
// more recent ELF binutils, we use .ident allowing the ID to be stripped.
// Usage:
//	__FBSDID("$FreeBSD$");

// -
// The following definitions are an extension of the behavior originally
// implemented in <sys/_posix.h>, but with a different level of granularity.
// POSIX.1 requires that the macros we test be defined before any standard
// header file is included.
//
// Here's a quick run-down of the versions:
//  defined(_POSIX_SOURCE)		1003.1-1988
//  _POSIX_C_SOURCE == 1		1003.1-1990
//  _POSIX_C_SOURCE == 2		1003.2-1992 C Language Binding Option
//  _POSIX_C_SOURCE == 199309		1003.1b-1993
//  _POSIX_C_SOURCE == 199506		1003.1c-1995, 1003.1i-1995,
//					and the omnibus ISO/IEC 9945-1: 1996
//  _POSIX_C_SOURCE == 200112		1003.1-2001
//  _POSIX_C_SOURCE == 200809		1003.1-2008
//
// In addition, the X/Open Portability Guide, which is now the Single UNIX
// Specification, defines a feature-test macro which indicates the version of
// that specification, and which subsumes _POSIX_C_SOURCE.
//
// Our macros begin with two underscores to avoid namespace screwage.

// Deal with IEEE Std. 1003.1-1990, in which _POSIX_C_SOURCE == 1.

// Deal with IEEE Std. 1003.2-1992, in which _POSIX_C_SOURCE == 2.

// Deal with various X/Open Portability Guides and Single UNIX Spec.

// Deal with all versions of POSIX.  The ordering relative to the tests above is
// important.
// -
// Deal with _ANSI_SOURCE:
// If it is defined, and no other compilation environment is explicitly
// requested, then define our internal feature-test macros to zero.  This
// makes no difference to the preprocessor (undefined symbols in preprocessing
// expressions are defined to have value zero), but makes it more convenient for
// a test program to print out the values.
//
// If a program mistakenly defines _ANSI_SOURCE and some other macro such as
// _POSIX_C_SOURCE, we will assume that it wants the broader compilation
// environment (and in fact we will never get here).

// User override __EXT1_VISIBLE

// Old versions of GCC use non-standard ARM arch symbols; acle-compat.h
// translates them to __ARM_ARCH and the modern feature symbols defined by ARM.

// Nullability qualifiers: currently only supported by Clang.

// Type Safety Checking
//
// Clang provides additional attributes to enable checking type safety
// properties that cannot be enforced by the C type system.

// Lock annotations.
//
// Clang provides support for doing basic thread-safety tests at
// compile-time, by marking which locks will/should be held when
// entering/leaving a functions.
//
// Furthermore, it is also possible to annotate variables and structure
// members to enforce that they are only accessed when certain locks are
// held.

// Structure implements a lock.

// Function acquires an exclusive or shared lock.

// Function attempts to acquire an exclusive or shared lock.

// Function releases a lock.

// Function asserts that an exclusive or shared lock is held.

// Function requires that an exclusive or shared lock is or is not held.

// Function should not be analyzed.

// Function or variable should not be sanitized, e.g., by AddressSanitizer.
// GCC has the nosanitize attribute, but as a function attribute only, and
// warns on use as a variable attribute.

// Guard variables and structure members by lock.

// Alignment builtins for better type checking and improved code generation.
// Provide fallback versions for other compilers (GCC/Clang < 10):

// -
// SPDX-License-Identifier: BSD-2-Clause-FreeBSD
//
// Copyright (c) 2002 Mike Barcroft <mike@FreeBSD.org>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHOR AND CONTRIBUTORS ``AS IS'' AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED.  IN NO EVENT SHALL THE AUTHOR OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
// OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
// LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
// OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// SUCH DAMAGE.
//
// $FreeBSD$

// -
// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 1991, 1993
//	The Regents of the University of California.  All rights reserved.
//
// This code is derived from software contributed to Berkeley by
// Berkeley Software Design, Inc.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the University nor the names of its contributors
//    may be used to endorse or promote products derived from this software
//    without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS ``AS IS'' AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED.  IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
// OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
// LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
// OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// SUCH DAMAGE.
//
//	@(#)cdefs.h	8.8 (Berkeley) 1/9/95
// $FreeBSD$

// -
// This file is in the public domain.
// $FreeBSD$

// -
// SPDX-License-Identifier: BSD-4-Clause
//
// Copyright (c) 2002 Mike Barcroft <mike@FreeBSD.org>
// Copyright (c) 1990, 1993
//	The Regents of the University of California.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. All advertising materials mentioning features or use of this software
//    must display the following acknowledgement:
//	This product includes software developed by the University of
//	California, Berkeley and its contributors.
// 4. Neither the name of the University nor the names of its contributors
//    may be used to endorse or promote products derived from this software
//    without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS ``AS IS'' AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED.  IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
// OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
// LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
// OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// SUCH DAMAGE.
//
//	From: @(#)ansi.h	8.2 (Berkeley) 1/4/94
//	From: @(#)types.h	8.3 (Berkeley) 1/5/94
// $FreeBSD$

// -
// This file is in the public domain.
// $FreeBSD$

// -
// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 1988, 1993
//	The Regents of the University of California.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the University nor the names of its contributors
//    may be used to endorse or promote products derived from this software
//    without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS ``AS IS'' AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED.  IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
// OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
// LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
// OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// SUCH DAMAGE.
//
//	@(#)limits.h	8.3 (Berkeley) 1/4/94
// $FreeBSD$

// According to ANSI (section 2.2.4.2), the values below must be usable by
// #if preprocessing directives.  Additionally, the expression must have the
// same type as would an expression that is an object of the corresponding
// type converted according to the integral promotions.  The subtraction for
// INT_MIN, etc., is so the value is not unsigned; e.g., 0x80000000 is an
// unsigned int for 32-bit two's complement ANSI compilers (section 3.1.3.2).

// max value for an unsigned long long

// Minimum signal stack size.

// Basic types upon which most other types are built.
type X__int8_t = int8     /* _types.h:55:22 */
type X__uint8_t = uint8   /* _types.h:56:24 */
type X__int16_t = int16   /* _types.h:57:17 */
type X__uint16_t = uint16 /* _types.h:58:25 */
type X__int32_t = int32   /* _types.h:59:15 */
type X__uint32_t = uint32 /* _types.h:60:23 */

type X__int64_t = int64 /* _types.h:66:20 */

type X__uint64_t = uint64 /* _types.h:68:28 */

// Standard type definitions.
type X__clock_t = uint32             /* _types.h:84:23 */
type X__critical_t = X__int32_t      /* _types.h:85:19 */
type X__double_t = float64           /* _types.h:87:21 */
type X__float_t = float64            /* _types.h:88:21 */
type X__intfptr_t = X__int32_t       /* _types.h:90:19 */
type X__intptr_t = X__int32_t        /* _types.h:91:19 */
type X__intmax_t = X__int64_t        /* _types.h:93:19 */
type X__int_fast8_t = X__int32_t     /* _types.h:94:19 */
type X__int_fast16_t = X__int32_t    /* _types.h:95:19 */
type X__int_fast32_t = X__int32_t    /* _types.h:96:19 */
type X__int_fast64_t = X__int64_t    /* _types.h:97:19 */
type X__int_least8_t = X__int8_t     /* _types.h:98:18 */
type X__int_least16_t = X__int16_t   /* _types.h:99:19 */
type X__int_least32_t = X__int32_t   /* _types.h:100:19 */
type X__int_least64_t = X__int64_t   /* _types.h:101:19 */
type X__ptrdiff_t = X__int32_t       /* _types.h:112:19 */
type X__register_t = X__int32_t      /* _types.h:113:19 */
type X__segsz_t = X__int32_t         /* _types.h:114:19 */
type X__size_t = X__uint32_t         /* _types.h:115:20 */
type X__ssize_t = X__int32_t         /* _types.h:116:19 */
type X__time_t = X__int32_t          /* _types.h:117:19 */
type X__uintfptr_t = X__uint32_t     /* _types.h:118:20 */
type X__uintptr_t = X__uint32_t      /* _types.h:119:20 */
type X__uintmax_t = X__uint64_t      /* _types.h:121:20 */
type X__uint_fast8_t = X__uint32_t   /* _types.h:122:20 */
type X__uint_fast16_t = X__uint32_t  /* _types.h:123:20 */
type X__uint_fast32_t = X__uint32_t  /* _types.h:124:20 */
type X__uint_fast64_t = X__uint64_t  /* _types.h:125:20 */
type X__uint_least8_t = X__uint8_t   /* _types.h:126:19 */
type X__uint_least16_t = X__uint16_t /* _types.h:127:20 */
type X__uint_least32_t = X__uint32_t /* _types.h:128:20 */
type X__uint_least64_t = X__uint64_t /* _types.h:129:20 */
type X__u_register_t = X__uint32_t   /* _types.h:136:20 */
type X__vm_offset_t = X__uint32_t    /* _types.h:137:20 */
type X__vm_paddr_t = X__uint64_t     /* _types.h:138:20 */
type X__vm_size_t = X__uint32_t      /* _types.h:139:20 */
type X___wchar_t = int32             /* _types.h:141:14 */

// Standard type definitions.
type X__blksize_t = X__int32_t   /* _types.h:40:19 */ // file block size
type X__blkcnt_t = X__int64_t    /* _types.h:41:19 */ // file block count
type X__clockid_t = X__int32_t   /* _types.h:42:19 */ // clock_gettime()...
type X__fflags_t = X__uint32_t   /* _types.h:43:20 */ // file flags
type X__fsblkcnt_t = X__uint64_t /* _types.h:44:20 */
type X__fsfilcnt_t = X__uint64_t /* _types.h:45:20 */
type X__gid_t = X__uint32_t      /* _types.h:46:20 */
type X__id_t = X__int64_t        /* _types.h:47:19 */ // can hold a gid_t, pid_t, or uid_t
type X__ino_t = X__uint64_t      /* _types.h:48:20 */ // inode number
type X__key_t = int32            /* _types.h:49:15 */ // IPC key (for Sys V IPC)
type X__lwpid_t = X__int32_t     /* _types.h:50:19 */ // Thread ID (a.k.a. LWP)
type X__mode_t = X__uint16_t     /* _types.h:51:20 */ // permissions
type X__accmode_t = int32        /* _types.h:52:14 */ // access permissions
type X__nl_item = int32          /* _types.h:53:14 */
type X__nlink_t = X__uint64_t    /* _types.h:54:20 */ // link count
type X__off_t = X__int64_t       /* _types.h:55:19 */ // file offset
type X__off64_t = X__int64_t     /* _types.h:56:19 */ // file offset (alias)
type X__pid_t = X__int32_t       /* _types.h:57:19 */ // process [group]
type X__rlim_t = X__int64_t      /* _types.h:58:19 */ // resource limit - intentionally
// signed, because of legacy code
// that uses -1 for RLIM_INFINITY
type X__sa_family_t = X__uint8_t /* _types.h:61:19 */
type X__socklen_t = X__uint32_t  /* _types.h:62:20 */
type X__suseconds_t = int32      /* _types.h:63:15 */ // microseconds (signed)
type X__timer_t = uintptr        /* _types.h:64:24 */ // timer_gettime()...
type X__mqd_t = uintptr          /* _types.h:65:21 */ // mq_open()...
type X__uid_t = X__uint32_t      /* _types.h:66:20 */
type X__useconds_t = uint32      /* _types.h:67:22 */ // microseconds (unsigned)
type X__cpuwhich_t = int32       /* _types.h:68:14 */ // which parameter for cpuset.
type X__cpulevel_t = int32       /* _types.h:69:14 */ // level parameter for cpuset.
type X__cpusetid_t = int32       /* _types.h:70:14 */ // cpuset identifier.
type X__daddr_t = X__int64_t     /* _types.h:71:19 */ // bwrite(3), FIOBMAP2, etc

// Unusual type definitions.
// rune_t is declared to be an “int” instead of the more natural
// “unsigned long” or “long”.  Two things are happening here.  It is not
// unsigned so that EOF (-1) can be naturally assigned to it and used.  Also,
// it looks like 10646 will be a 31 bit standard.  This means that if your
// ints cannot hold 32 bits, you will be in trouble.  The reason an int was
// chosen over a long is that the is*() and to*() routines take ints (says
// ANSI C), but they use __ct_rune_t instead of int.
//
// NOTE: rune_t is not covered by ANSI nor other standards, and should not
// be instantiated outside of lib/libc/locale.  Use wchar_t.  wint_t and
// rune_t must be the same type.  Also, wint_t should be able to hold all
// members of the largest character set plus one extra value (WEOF), and
// must be at least 16 bits.
type X__ct_rune_t = int32     /* _types.h:91:14 */ // arg type for ctype funcs
type X__rune_t = X__ct_rune_t /* _types.h:92:21 */ // rune_t (see above)
type X__wint_t = X__ct_rune_t /* _types.h:93:21 */ // wint_t (see above)

// Clang already provides these types as built-ins, but only in C++ mode.
type X__char16_t = X__uint_least16_t /* _types.h:97:26 */
type X__char32_t = X__uint_least32_t /* _types.h:98:26 */
// In C++11, char16_t and char32_t are built-in types.

type X__max_align_t = struct {
	F__max_align1 int64
	F__max_align2 float64
} /* _types.h:111:3 */

type X__dev_t = X__uint64_t /* _types.h:113:20 */ // device number

type X__fixpt_t = X__uint32_t /* _types.h:115:20 */ // fixed point number

// mbstate_t is an opaque object to keep conversion state during multibyte
// stream conversions.
type X__mbstate_t = struct {
	F__ccgo_pad1 [0]uint32
	F__mbstate8  [128]int8
} /* _types.h:124:3 */

type X__rman_res_t = X__uintmax_t /* _types.h:126:25 */

// Types for varargs. These are all provided by builtin types these
// days, so centralize their definition.
type X__va_list = X__builtin_va_list /* _types.h:133:27 */ // internally known to gcc
type X__gnuc_va_list = X__va_list    /* _types.h:140:20 */ // compatibility w/GNU headers

// When the following macro is defined, the system uses 64-bit inode numbers.
// Programs can use this to avoid including <sys/param.h>, with its associated
// namespace pollution.

type Nl_item = X__nl_item /* langinfo.h:38:19 */

// -
// SPDX-License-Identifier: BSD-2-Clause-FreeBSD
//
// Copyright (c) 2011, 2012 The FreeBSD Foundation
//
// This software was developed by David Chisnall under sponsorship from
// the FreeBSD Foundation.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHOR AND CONTRIBUTORS ``AS IS'' AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED.  IN NO EVENT SHALL THE AUTHOR OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
// OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
// LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
// OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// SUCH DAMAGE.
//
// $FreeBSD$

type Locale_t = uintptr /* _langinfo.h:38:25 */

var _ int8 /* gen.c:2:13: */
