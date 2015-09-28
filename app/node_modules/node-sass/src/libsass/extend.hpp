#ifndef SASS_EXTEND_H
#define SASS_EXTEND_H

#include <map>
#include <set>
#include <vector>
#include <iostream>

#include "ast.hpp"
#include "operation.hpp"
#include "subset_map.hpp"

namespace Sass {
  using namespace std;

  class Context;

  typedef Subset_Map<string, pair<Complex_Selector*, Compound_Selector*> > ExtensionSubsetMap;

  class Extend : public Operation_CRTP<void, Extend> {

    Context&            ctx;
    ExtensionSubsetMap& subset_map;

    void fallback_impl(AST_Node* n) { };

  public:
    Extend(Context&, ExtensionSubsetMap&);
    virtual ~Extend() { }

    using Operation<void>::operator();

    void operator()(Block*);
    void operator()(Ruleset*);
    void operator()(Feature_Block*);
    void operator()(Media_Block*);
    void operator()(At_Rule*);

    template <typename U>
    void fallback(U x) { return fallback_impl(x); }
  };

}

#endif
