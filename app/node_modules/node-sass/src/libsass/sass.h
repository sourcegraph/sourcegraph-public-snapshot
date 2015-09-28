#ifndef SASS_H
#define SASS_H

#include <stddef.h>
#include <stdbool.h>

#ifdef __GNUC__
  #define DEPRECATED(func) func __attribute__ ((deprecated))
#elif defined(_MSC_VER)
  #define DEPRECATED(func) __declspec(deprecated) func
#else
  #pragma message("WARNING: You need to implement DEPRECATED for this compiler")
  #define DEPRECATED(func) func
#endif

#ifdef _WIN32

  /* You should define ADD_EXPORTS *only* when building the DLL. */
  #ifdef ADD_EXPORTS
    #define ADDAPI __declspec(dllexport)
	#define ADDCALL __cdecl
  #else
    #define ADDAPI
	#define ADDCALL
  #endif

#else /* _WIN32 not defined. */

  /* Define with no value on non-Windows OSes. */
  #define ADDAPI
  #define ADDCALL

#endif

// include API headers
#include "sass_version.h"
#include "sass_values.h"
#include "sass_functions.h"

/* Make sure functions are exported with C linkage under C++ compilers. */
#ifdef __cplusplus
extern "C" {
#endif


// Different render styles
enum Sass_Output_Style {
  SASS_STYLE_NESTED,
  SASS_STYLE_EXPANDED,
  SASS_STYLE_COMPACT,
  SASS_STYLE_COMPRESSED
};

// Some convenient string helper function
ADDAPI char* ADDCALL sass_string_quote (const char* str, const char quote_mark);
ADDAPI char* ADDCALL sass_string_unquote (const char* str);

// Resolve a file via the given include paths in the include char* array
ADDAPI char* ADDCALL sass_resolve_file (const char* path, const char* incs[]);

// Get compiled libsass version
ADDAPI const char* ADDCALL libsass_version(void);

#ifdef __cplusplus
} // __cplusplus defined.
#endif

#endif
