#include <nan.h>
#include <stdlib.h>
#include <string.h>
#include "create_string.h"

char* create_string(Local<Value> value) {
  if (value->IsNull() || !value->IsString()) {
    return 0;
  }

  String::Utf8Value string(value);
  char *str = (char *)malloc(string.length() + 1);
  strcpy(str, *string);
  return str;
}
