#include <nan.h>
#include "color.h"

using namespace v8;

namespace SassTypes
{
  Color::Color(Sass_Value* v) : SassValueWrapper(v) {}

  Sass_Value* Color::construct(const std::vector<Local<v8::Value>> raw_val) {
    double a = 1.0, r = 0, g = 0, b = 0;
    unsigned argb;

    switch (raw_val.size()) {
    case 1:
      if (!raw_val[0]->IsNumber()) {
        throw std::invalid_argument("Only argument should be an integer.");
      }

      argb = raw_val[0]->ToInt32()->Value();
      a = (double)((argb >> 030) & 0xff) / 0xff;
      r = (double)((argb >> 020) & 0xff);
      g = (double)((argb >> 010) & 0xff);
      b = (double)(argb & 0xff);
      break;

    case 4:
      if (!raw_val[3]->IsNumber()) {
        throw std::invalid_argument("Constructor arguments should be numbers exclusively.");
      }

      a = raw_val[3]->ToNumber()->Value();
      // fall through vvv

    case 3:
      if (!raw_val[0]->IsNumber() || !raw_val[1]->IsNumber() || !raw_val[2]->IsNumber()) {
        throw std::invalid_argument("Constructor arguments should be numbers exclusively.");
      }

      r = raw_val[0]->ToNumber()->Value();
      g = raw_val[1]->ToNumber()->Value();
      b = raw_val[2]->ToNumber()->Value();
      break;

    case 0:
      break;

    default:
      throw std::invalid_argument("Constructor should be invoked with either 0, 1, 3 or 4 arguments.");
    }

    return sass_make_color(r, g, b, a);
  }

  void Color::initPrototype(Handle<ObjectTemplate> proto) {
    proto->Set(NanNew("getR"), NanNew<FunctionTemplate>(GetR)->GetFunction());
    proto->Set(NanNew("getG"), NanNew<FunctionTemplate>(GetG)->GetFunction());
    proto->Set(NanNew("getB"), NanNew<FunctionTemplate>(GetB)->GetFunction());
    proto->Set(NanNew("getA"), NanNew<FunctionTemplate>(GetA)->GetFunction());
    proto->Set(NanNew("setR"), NanNew<FunctionTemplate>(SetR)->GetFunction());
    proto->Set(NanNew("setG"), NanNew<FunctionTemplate>(SetG)->GetFunction());
    proto->Set(NanNew("setB"), NanNew<FunctionTemplate>(SetB)->GetFunction());
    proto->Set(NanNew("setA"), NanNew<FunctionTemplate>(SetA)->GetFunction());
  }

  NAN_METHOD(Color::GetR) {
    NanScope();
    NanReturnValue(NanNew(sass_color_get_r(unwrap(args.This())->value)));
  }

  NAN_METHOD(Color::GetG) {
    NanScope();
    NanReturnValue(NanNew(sass_color_get_g(unwrap(args.This())->value)));
  }

  NAN_METHOD(Color::GetB) {
    NanScope();
    NanReturnValue(NanNew(sass_color_get_b(unwrap(args.This())->value)));
  }

  NAN_METHOD(Color::GetA) {
    NanScope();
    NanReturnValue(NanNew(sass_color_get_a(unwrap(args.This())->value)));
  }

  NAN_METHOD(Color::SetR) {
    if (args.Length() != 1) {
      return NanThrowError(NanNew("Expected just one argument"));
    }

    if (!args[0]->IsNumber()) {
      return NanThrowError(NanNew("Supplied value should be a number"));
    }

    sass_color_set_r(unwrap(args.This())->value, args[0]->ToNumber()->Value());
    NanReturnUndefined();
  }

  NAN_METHOD(Color::SetG) {
    if (args.Length() != 1) {
      return NanThrowError(NanNew("Expected just one argument"));
    }

    if (!args[0]->IsNumber()) {
      return NanThrowError(NanNew("Supplied value should be a number"));
    }

    sass_color_set_g(unwrap(args.This())->value, args[0]->ToNumber()->Value());
    NanReturnUndefined();
  }

  NAN_METHOD(Color::SetB) {
    if (args.Length() != 1) {
      return NanThrowError(NanNew("Expected just one argument"));
    }

    if (!args[0]->IsNumber()) {
      return NanThrowError(NanNew("Supplied value should be a number"));
    }

    sass_color_set_b(unwrap(args.This())->value, args[0]->ToNumber()->Value());
    NanReturnUndefined();
  }

  NAN_METHOD(Color::SetA) {
    if (args.Length() != 1) {
      return NanThrowError(NanNew("Expected just one argument"));
    }

    if (!args[0]->IsNumber()) {
      return NanThrowError(NanNew("Supplied value should be a number"));
    }

    sass_color_set_a(unwrap(args.This())->value, args[0]->ToNumber()->Value());
    NanReturnUndefined();
  }
}
