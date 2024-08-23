// Code generated by 'ccgo netdb/gen.c -crt-import-path "" -export-defines "" -export-enums "" -export-externs X -export-fields F -export-structs "" -export-typedefs "" -header -hide _OSSwapInt16,_OSSwapInt32,_OSSwapInt64 -ignore-unsupported-alignment -o netdb/netdb_freebsd_amd64.go -pkgname netdb', DO NOT EDIT.

package netdb

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
	AI_ADDRCONFIG        = 0x00000400            // netdb.h:200:1:
	AI_ALL               = 0x00000100            // netdb.h:198:1:
	AI_CANONNAME         = 0x00000002            // netdb.h:190:1:
	AI_DEFAULT           = 1536                  // netdb.h:203:1:
	AI_MASK              = 3343                  // netdb.h:194:1:
	AI_NUMERICHOST       = 0x00000004            // netdb.h:191:1:
	AI_NUMERICSERV       = 0x00000008            // netdb.h:192:1:
	AI_PASSIVE           = 0x00000001            // netdb.h:189:1:
	AI_V4MAPPED          = 0x00000800            // netdb.h:201:1:
	AI_V4MAPPED_CFG      = 0x00000200            // netdb.h:199:1:
	EAI_AGAIN            = 2                     // netdb.h:168:1:
	EAI_BADFLAGS         = 3                     // netdb.h:169:1:
	EAI_BADHINTS         = 12                    // netdb.h:181:1:
	EAI_FAIL             = 4                     // netdb.h:170:1:
	EAI_FAMILY           = 5                     // netdb.h:171:1:
	EAI_MAX              = 15                    // netdb.h:184:1:
	EAI_MEMORY           = 6                     // netdb.h:172:1:
	EAI_NONAME           = 8                     // netdb.h:177:1:
	EAI_OVERFLOW         = 14                    // netdb.h:183:1:
	EAI_PROTOCOL         = 13                    // netdb.h:182:1:
	EAI_SERVICE          = 9                     // netdb.h:178:1:
	EAI_SOCKTYPE         = 10                    // netdb.h:179:1:
	EAI_SYSTEM           = 11                    // netdb.h:180:1:
	HOST_NOT_FOUND       = 1                     // netdb.h:155:1:
	IPPORT_RESERVED      = 1024                  // netdb.h:146:1:
	NETDB_INTERNAL       = -1                    // netdb.h:153:1:
	NETDB_SUCCESS        = 0                     // netdb.h:154:1:
	NI_DGRAM             = 0x00000010            // netdb.h:218:1:
	NI_MAXHOST           = 1025                  // netdb.h:208:1:
	NI_MAXSERV           = 32                    // netdb.h:209:1:
	NI_NAMEREQD          = 0x00000004            // netdb.h:216:1:
	NI_NOFQDN            = 0x00000001            // netdb.h:214:1:
	NI_NUMERICHOST       = 0x00000002            // netdb.h:215:1:
	NI_NUMERICSCOPE      = 0x00000020            // netdb.h:219:1:
	NI_NUMERICSERV       = 0x00000008            // netdb.h:217:1:
	NO_ADDRESS           = 4                     // netdb.h:159:1:
	NO_DATA              = 4                     // netdb.h:158:1:
	NO_RECOVERY          = 3                     // netdb.h:157:1:
	SCOPE_DELIMITER      = 37                    // netdb.h:224:1:
	TRY_AGAIN            = 2                     // netdb.h:156:1:
	X_FILE_OFFSET_BITS   = 64                    // <builtin>:25:1:
	X_IN_ADDR_T_DECLARED = 0                     // netdb.h:67:1:
	X_IN_PORT_T_DECLARED = 0                     // netdb.h:72:1:
	X_LP64               = 1                     // <predefined>:1:1:
	X_MACHINE__LIMITS_H_ = 0                     // _limits.h:36:1:
	X_MACHINE__TYPES_H_  = 0                     // _types.h:42:1:
	X_NETDB_H_           = 0                     // netdb.h:60:1:
	X_Nonnull            = 0                     // cdefs.h:790:1:
	X_Null_unspecified   = 0                     // cdefs.h:792:1:
	X_Nullable           = 0                     // cdefs.h:791:1:
	X_PATH_HEQUIV        = "/etc/hosts.equiv"    // netdb.h:91:1:
	X_PATH_HOSTS         = "/etc/hosts"          // netdb.h:93:1:
	X_PATH_NETWORKS      = "/etc/networks"       // netdb.h:94:1:
	X_PATH_PROTOCOLS     = "/etc/protocols"      // netdb.h:95:1:
	X_PATH_SERVICES      = "/etc/services"       // netdb.h:96:1:
	X_PATH_SERVICES_DB   = "/var/db/services.db" // netdb.h:97:1:
	X_SIZE_T_DECLARED    = 0                     // netdb.h:77:1:
	X_SOCKLEN_T_DECLARED = 0                     // netdb.h:82:1:
	X_SYS_CDEFS_H_       = 0                     // cdefs.h:39:1:
	X_SYS__TYPES_H_      = 0                     // _types.h:32:1:
	X_UINT32_T_DECLARED  = 0                     // netdb.h:87:1:
	Unix                 = 1                     // <predefined>:340:1:
)

type Ptrdiff_t = int64 /* <builtin>:3:26 */

type Size_t = uint64 /* <builtin>:9:23 */

type Wchar_t = int32 /* <builtin>:15:24 */

type X__int128_t = struct {
	Flo int64
	Fhi int64
} /* <builtin>:21:43 */ // must match modernc.org/mathutil.Int128
type X__uint128_t = struct {
	Flo uint64
	Fhi uint64
} /* <builtin>:22:44 */ // must match modernc.org/mathutil.Int128

type X__builtin_va_list = uintptr /* <builtin>:46:14 */
type X__float128 = float64        /* <builtin>:47:21 */

// -
// SPDX-License-Identifier: (BSD-3-Clause AND ISC)
//
// Copyright (c) 1980, 1983, 1988, 1993
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
// -
// Portions Copyright (c) 1993 by Digital Equipment Corporation.
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies, and that
// the name of Digital Equipment Corporation not be used in advertising or
// publicity pertaining to distribution of the document or software without
// specific, written prior permission.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND DIGITAL EQUIPMENT CORP. DISCLAIMS ALL
// WARRANTIES WITH REGARD TO THIS SOFTWARE, INCLUDING ALL IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS.   IN NO EVENT SHALL DIGITAL EQUIPMENT
// CORPORATION BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL
// DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR
// PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS
// ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS
// SOFTWARE.
// -
// --Copyright--

//      @(#)netdb.h	8.1 (Berkeley) 6/2/93
//      From: Id: netdb.h,v 8.9 1996/11/19 08:39:29 vixie Exp $
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

// Quads and longs are the same on the amd64.  Ensure they stay in sync.

// Minimum signal stack size.

// Basic types upon which most other types are built.
type X__int8_t = int8     /* _types.h:55:22 */
type X__uint8_t = uint8   /* _types.h:56:24 */
type X__int16_t = int16   /* _types.h:57:17 */
type X__uint16_t = uint16 /* _types.h:58:25 */
type X__int32_t = int32   /* _types.h:59:15 */
type X__uint32_t = uint32 /* _types.h:60:23 */
type X__int64_t = int64   /* _types.h:62:16 */
type X__uint64_t = uint64 /* _types.h:63:24 */

// Standard type definitions.
type X__clock_t = X__int32_t         /* _types.h:75:19 */ // clock()...
type X__critical_t = X__int64_t      /* _types.h:76:19 */
type X__double_t = float64           /* _types.h:78:17 */
type X__float_t = float32            /* _types.h:79:16 */
type X__intfptr_t = X__int64_t       /* _types.h:81:19 */
type X__intptr_t = X__int64_t        /* _types.h:82:19 */
type X__intmax_t = X__int64_t        /* _types.h:93:19 */
type X__int_fast8_t = X__int32_t     /* _types.h:94:19 */
type X__int_fast16_t = X__int32_t    /* _types.h:95:19 */
type X__int_fast32_t = X__int32_t    /* _types.h:96:19 */
type X__int_fast64_t = X__int64_t    /* _types.h:97:19 */
type X__int_least8_t = X__int8_t     /* _types.h:98:18 */
type X__int_least16_t = X__int16_t   /* _types.h:99:19 */
type X__int_least32_t = X__int32_t   /* _types.h:100:19 */
type X__int_least64_t = X__int64_t   /* _types.h:101:19 */
type X__ptrdiff_t = X__int64_t       /* _types.h:103:19 */ // ptr1 - ptr2
type X__register_t = X__int64_t      /* _types.h:104:19 */
type X__segsz_t = X__int64_t         /* _types.h:105:19 */ // segment size (in pages)
type X__size_t = X__uint64_t         /* _types.h:106:20 */ // sizeof()
type X__ssize_t = X__int64_t         /* _types.h:107:19 */ // byte count or error
type X__time_t = X__int64_t          /* _types.h:108:19 */ // time()...
type X__uintfptr_t = X__uint64_t     /* _types.h:109:20 */
type X__uintptr_t = X__uint64_t      /* _types.h:110:20 */
type X__uintmax_t = X__uint64_t      /* _types.h:121:20 */
type X__uint_fast8_t = X__uint32_t   /* _types.h:122:20 */
type X__uint_fast16_t = X__uint32_t  /* _types.h:123:20 */
type X__uint_fast32_t = X__uint32_t  /* _types.h:124:20 */
type X__uint_fast64_t = X__uint64_t  /* _types.h:125:20 */
type X__uint_least8_t = X__uint8_t   /* _types.h:126:19 */
type X__uint_least16_t = X__uint16_t /* _types.h:127:20 */
type X__uint_least32_t = X__uint32_t /* _types.h:128:20 */
type X__uint_least64_t = X__uint64_t /* _types.h:129:20 */
type X__u_register_t = X__uint64_t   /* _types.h:131:20 */
type X__vm_offset_t = X__uint64_t    /* _types.h:132:20 */
type X__vm_paddr_t = X__uint64_t     /* _types.h:133:20 */
type X__vm_size_t = X__uint64_t      /* _types.h:134:20 */
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
type X__key_t = int64            /* _types.h:49:15 */ // IPC key (for Sys V IPC)
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
type X__suseconds_t = int64      /* _types.h:63:15 */ // microseconds (signed)
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
	F__ccgo_pad1 [0]uint64
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

type In_addr_t = X__uint32_t /* netdb.h:66:20 */

type In_port_t = X__uint16_t /* netdb.h:71:20 */

type Socklen_t = X__socklen_t /* netdb.h:81:21 */

type Uint32_t = X__uint32_t /* netdb.h:86:20 */

// Structures returned by network data base library.  All addresses are
// supplied in host order, and returned in network order (suitable for
// use in system calls).
type Hostent = struct {
	Fh_name      uintptr
	Fh_aliases   uintptr
	Fh_addrtype  int32
	Fh_length    int32
	Fh_addr_list uintptr
} /* netdb.h:106:1 */

type Netent = struct {
	Fn_name     uintptr
	Fn_aliases  uintptr
	Fn_addrtype int32
	Fn_net      Uint32_t
} /* netdb.h:115:1 */

type Servent = struct {
	Fs_name      uintptr
	Fs_aliases   uintptr
	Fs_port      int32
	F__ccgo_pad1 [4]byte
	Fs_proto     uintptr
} /* netdb.h:122:1 */

type Protoent = struct {
	Fp_name      uintptr
	Fp_aliases   uintptr
	Fp_proto     int32
	F__ccgo_pad1 [4]byte
} /* netdb.h:129:1 */

type Addrinfo = struct {
	Fai_flags     int32
	Fai_family    int32
	Fai_socktype  int32
	Fai_protocol  int32
	Fai_addrlen   Socklen_t
	F__ccgo_pad1  [4]byte
	Fai_canonname uintptr
	Fai_addr      uintptr
	Fai_next      uintptr
} /* netdb.h:135:1 */

var _ int8 /* gen.c:2:13: */
