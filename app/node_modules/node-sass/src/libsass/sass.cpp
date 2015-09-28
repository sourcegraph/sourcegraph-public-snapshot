#include <cstdlib>
#include <cstring>
#include <vector>
#include <sstream>

#include "sass.h"
#include "file.hpp"
#include "util.hpp"

extern "C" {
  using namespace std;
  using namespace Sass;
  using namespace File;

  // caller must free the returned memory
  char* ADDCALL sass_string_quote (const char *str, const char quote_mark)
  {
    string quoted = quote(str, quote_mark);
    return sass_strdup(quoted.c_str());
  }

  // caller must free the returned memory
  char* ADDCALL sass_string_unquote (const char *str)
  {
    string unquoted = unquote(str);
    return sass_strdup(unquoted.c_str());
  }

  // Make sure to free the returned value!
  // Incs array has to be null terminated!
  char* ADDCALL sass_resolve_file (const char* file, const char* paths[])
  {
    string resolved(find_file(file, paths));
    return sass_strdup(resolved.c_str());
  }

  // Get compiled libsass version
  const char* ADDCALL libsass_version(void)
  {
    return LIBSASS_VERSION;
  }

}
