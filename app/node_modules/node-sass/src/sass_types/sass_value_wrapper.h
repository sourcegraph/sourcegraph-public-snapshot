#ifndef SASS_TYPES_SASS_VALUE_WRAPPER_H
#define SASS_TYPES_SASS_VALUE_WRAPPER_H

#include <stdexcept>
#include <vector>
#include <nan.h>
#include "value.h"
#include "factory.h"

namespace SassTypes
{
  using namespace v8;

  // Include this in any SassTypes::Value subclasses to handle all the heavy lifting of constructing JS
  // objects and wrapping sass values inside them
  template <class T>
  class SassValueWrapper : public Value {
    public:
      static char const* get_constructor_name() { return "SassValue"; }

      SassValueWrapper(Sass_Value*);
      virtual ~SassValueWrapper();

      Sass_Value* get_sass_value();
      Local<Object> get_js_object();

      static Handle<Function> get_constructor();
      static Local<FunctionTemplate> get_constructor_template();
      static NAN_METHOD(New);

    protected:
      Sass_Value* value;
      static T* unwrap(Local<Object>);

    private:
      static Persistent<Function> constructor;
      Persistent<Object> js_object;
  };

  template <class T>
  Persistent<Function> SassValueWrapper<T>::constructor;

  template <class T>
  SassValueWrapper<T>::SassValueWrapper(Sass_Value* v) {
    this->value = sass_clone_value(v);
  }

  template <class T>
  SassValueWrapper<T>::~SassValueWrapper() {
    NanDisposePersistent(this->js_object);
    sass_delete_value(this->value);
  }

  template <class T>
  Sass_Value* SassValueWrapper<T>::get_sass_value() {
    return sass_clone_value(this->value);
  }

  template <class T>
  Local<Object> SassValueWrapper<T>::get_js_object() {
    if (this->js_object.IsEmpty()) {
      Local<Object> wrapper = NanNew(T::get_constructor())->NewInstance();
      delete static_cast<T*>(NanGetInternalFieldPointer(wrapper, 0));
      NanSetInternalFieldPointer(wrapper, 0, this);
      NanAssignPersistent(this->js_object, wrapper);
    }

    return NanNew(this->js_object);
  }

  template <class T>
  Local<FunctionTemplate> SassValueWrapper<T>::get_constructor_template() {
    Local<FunctionTemplate> tpl = NanNew<FunctionTemplate>(New);
    tpl->SetClassName(NanNew(NanNew(T::get_constructor_name())));
    tpl->InstanceTemplate()->SetInternalFieldCount(1);
    T::initPrototype(tpl->PrototypeTemplate());

    return tpl;
  }

  template <class T>
  Handle<Function> SassValueWrapper<T>::get_constructor() {
    if (constructor.IsEmpty()) {
      NanAssignPersistent(constructor, T::get_constructor_template()->GetFunction());
    }

    return NanNew(constructor);
  }

  template <class T>
  NAN_METHOD(SassValueWrapper<T>::New) {
    NanScope();

    if (!args.IsConstructCall()) {
      unsigned argc = args.Length();
      std::vector<Handle<v8::Value>> argv;

      argv.reserve(argc);
      for (unsigned i = 0; i < argc; i++) {
        argv.push_back(args[i]);
      }

      NanReturnValue(NanNew(T::get_constructor())->NewInstance(argc, &argv[0]));
    }

    std::vector<Local<v8::Value>> localArgs(args.Length());

    for (auto i = 0; i < args.Length(); ++i) {
      localArgs[i] = args[i];
    }

    try {
      Sass_Value* value = T::construct(localArgs);
      T* obj = new T(value);
      sass_delete_value(value);

      NanSetInternalFieldPointer(args.This(), 0, obj);
      NanAssignPersistent(obj->js_object, args.This());
    } catch (const std::exception& e) {
      return NanThrowError(NanNew(e.what()));
    }

    NanReturnUndefined();
  }

  template <class T>
  T* SassValueWrapper<T>::unwrap(Local<Object> obj) {
    return static_cast<T*>(Factory::unwrap(obj));
  }
}


#endif
