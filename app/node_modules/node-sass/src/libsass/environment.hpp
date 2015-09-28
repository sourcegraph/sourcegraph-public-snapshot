#ifndef SASS_ENVIRONMENT_H
#define SASS_ENVIRONMENT_H

#include <map>
#include <string>
#include <iostream>

#include "ast_fwd_decl.hpp"
#include "ast_def_macros.hpp"
#include "memory_manager.hpp"

namespace Sass {
  using std::string;
  using std::map;
  using std::cerr;
  using std::endl;

  template <typename T>
  class Environment {
    // TODO: test with unordered_map
    map<string, T> local_frame_;
    ADD_PROPERTY(Environment*, parent);

  public:
    Memory_Manager<AST_Node> mem;
    Environment() : local_frame_(map<string, T>()), parent_(0) { }

    // link parent to create a stack
    void link(Environment& env) { parent_ = &env; }
    void link(Environment* env) { parent_ = env; }

    // this is used to find the global frame
    // which is the second last on the stack
    bool is_lexical() const
    {
      return !! parent_ && parent_->parent_;
    }

    // only match the real root scope
    // there is still a parent around
    // not sure what it is actually use for
    // I guess we store functions etc. there
    bool is_root_scope() const
    {
      return parent_ && ! parent_->parent_;
    }

    // scope operates on the current frame

    map<string, T>& local_frame() {
      return local_frame_;
    }

    bool has_local(const string& key) const
    { return local_frame_.count(key); }

    T& get_local(const string& key)
    { return local_frame_[key]; }

    void set_local(const string& key, T val)
    { local_frame_[key] = val; }

    void del_local(const string& key)
    { local_frame_.erase(key); }


    // global operates on the global frame
    // which is the second last on the stack

    Environment* global_env()
    {
      Environment* cur = this;
      while (cur->is_lexical()) {
        cur = cur->parent_;
      }
      return cur;
    }

    bool has_global(const string& key)
    { return global_env()->has(key); }

    T& get_global(const string& key)
    { return (*global_env())[key]; }

    void set_global(const string& key, T val)
    { global_env()->local_frame_[key] = val; }

    void del_global(const string& key)
    { global_env()->local_frame_.erase(key); }

    // see if we have a lexical variable
    // move down the stack but stop before we
    // reach the global frame (is not included)
    bool has_lexical(const string& key) const
    {
      auto cur = this;
      while (cur->is_lexical()) {
        if (cur->has_local(key)) return true;
        cur = cur->parent_;
      }
      return false;
    }

    // see if we have a lexical we could update
    // either update already existing lexical value
    // or if flag is set, we create one if no lexical found
    void set_lexical(const string& key, T val)
    {
      auto cur = this;
      while (cur->is_lexical()) {
        if (cur->has_local(key)) {
          cur->set_local(key, val);
          return;
        }
        cur = cur->parent_;
      }
      set_local(key, val);
    }

    // look on the full stack for key
    // include all scopes available
    bool has(const string& key) const
    {
      auto cur = this;
      while (cur) {
        if (cur->has_local(key)) {
          return true;
        }
        cur = cur->parent_;
      }
      return false;
    }

    // use array access for getter and setter functions
    T& operator[](const string& key)
    {
      auto cur = this;
      while (cur) {
        if (cur->has_local(key)) {
          return cur->get_local(key);
        }
        cur = cur->parent_;
      }
      return get_local(key);
    }

    #ifdef DEBUG
    void print()
    {
      for (typename map<string, T>::iterator i = local_frame_.begin(); i != local_frame_.end(); ++i) {
        cerr << i->first << endl;
      }
      if (parent_) {
        cerr << "---" << endl;
        parent_->print();
      }
    }
    #endif

  };
}

#endif
