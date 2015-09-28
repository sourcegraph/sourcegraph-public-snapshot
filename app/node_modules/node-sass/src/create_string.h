#ifndef CREATE_STRING_H
#define CREATE_STRING_H

#include <nan.h>

using namespace v8;

char* create_string(Local<Value>);

#endif