#include "ast.hpp"
#include "context.hpp"
#include "to_string.hpp"
#include <set>
#include <algorithm>
#include <iostream>

namespace Sass {
  using namespace std;

  static Null sass_null(Sass::Null(ParserState("null")));

  bool Compound_Selector::operator<(const Compound_Selector& rhs) const
  {
    To_String to_string;
    // ugly
    return const_cast<Compound_Selector*>(this)->perform(&to_string) <
           const_cast<Compound_Selector&>(rhs).perform(&to_string);
  }

  bool Complex_Selector::operator<(const Complex_Selector& rhs) const
  {
    To_String to_string;
    return const_cast<Complex_Selector*>(this)->perform(&to_string) <
           const_cast<Complex_Selector&>(rhs).perform(&to_string);
  }

  bool Complex_Selector::operator==(const Complex_Selector& rhs) const {
    // TODO: We have to access the tail directly using tail_ since ADD_PROPERTY doesn't provide a const version.

    const Complex_Selector* pOne = this;
    const Complex_Selector* pTwo = &rhs;

    // Consume any empty references at the beginning of the Complex_Selector
    if (pOne->combinator() == Complex_Selector::ANCESTOR_OF && pOne->head()->is_empty_reference()) {
      pOne = pOne->tail_;
    }
    if (pTwo->combinator() == Complex_Selector::ANCESTOR_OF && pTwo->head()->is_empty_reference()) {
      pTwo = pTwo->tail_;
    }

    while (pOne && pTwo) {
      if (pOne->combinator() != pTwo->combinator()) {
        return false;
      }

      if (*(pOne->head()) != *(pTwo->head())) {
        return false;
      }

      pOne = pOne->tail_;
      pTwo = pTwo->tail_;
    }

    return pOne == NULL && pTwo == NULL;
  }

  Compound_Selector* Compound_Selector::unify_with(Compound_Selector* rhs, Context& ctx)
  {
    Compound_Selector* unified = rhs;
    for (size_t i = 0, L = length(); i < L; ++i)
    {
      if (!unified) break;
      else          unified = (*this)[i]->unify_with(unified, ctx);
    }
    return unified;
  }

  bool Simple_Selector::operator==(const Simple_Selector& rhs) const
  {
    // Compare the string representations for equality.

    // Cast away const here. To_String should take a const object, but it doesn't.
    Simple_Selector* pLHS = const_cast<Simple_Selector*>(this);
    Simple_Selector* pRHS = const_cast<Simple_Selector*>(&rhs);

    To_String to_string;
    return pLHS->perform(&to_string) == pRHS->perform(&to_string);
  }

  bool Simple_Selector::operator<(const Simple_Selector& rhs) const {
    // Use the string representation for ordering.

    // Cast away const here. To_String should take a const object, but it doesn't.
    Simple_Selector* pLHS = const_cast<Simple_Selector*>(this);
    Simple_Selector* pRHS = const_cast<Simple_Selector*>(&rhs);

    To_String to_string;
    return pLHS->perform(&to_string) < pRHS->perform(&to_string);
  }

  Compound_Selector* Simple_Selector::unify_with(Compound_Selector* rhs, Context& ctx)
  {
    To_String to_string(&ctx);
    for (size_t i = 0, L = rhs->length(); i < L; ++i)
    { if (perform(&to_string) == (*rhs)[i]->perform(&to_string)) return rhs; }

    // check for pseudo elements because they need to come last
    size_t i, L;
    bool found = false;
    if (typeid(*this) == typeid(Pseudo_Selector) || typeid(*this) == typeid(Wrapped_Selector))
    {
      for (i = 0, L = rhs->length(); i < L; ++i)
      {
        if ((typeid(*(*rhs)[i]) == typeid(Pseudo_Selector) || typeid(*(*rhs)[i]) == typeid(Wrapped_Selector)) && (*rhs)[L-1]->is_pseudo_element())
        { found = true; break; }
      }
    }
    else
    {
      for (i = 0, L = rhs->length(); i < L; ++i)
      {
        if (typeid(*(*rhs)[i]) == typeid(Pseudo_Selector) || typeid(*(*rhs)[i]) == typeid(Wrapped_Selector))
        { found = true; break; }
      }
    }
    if (!found)
    {
      Compound_Selector* cpy = new (ctx.mem) Compound_Selector(*rhs);
      (*cpy) << this;
      return cpy;
    }
    Compound_Selector* cpy = new (ctx.mem) Compound_Selector(rhs->pstate());
    for (size_t j = 0; j < i; ++j)
    { (*cpy) << (*rhs)[j]; }
    (*cpy) << this;
    for (size_t j = i; j < L; ++j)
    { (*cpy) << (*rhs)[j]; }
    return cpy;
  }

  Compound_Selector* Type_Selector::unify_with(Compound_Selector* rhs, Context& ctx)
  {
    // TODO: handle namespaces

    // if the rhs is empty, just return a copy of this
    if (rhs->length() == 0) {
      Compound_Selector* cpy = new (ctx.mem) Compound_Selector(rhs->pstate());
      (*cpy) << this;
      return cpy;
    }

    // if this is a universal selector and rhs is not empty, just return the rhs
    if (name() == "*")
    { return new (ctx.mem) Compound_Selector(*rhs); }


    Simple_Selector* rhs_0 = (*rhs)[0];
    // otherwise, this is a tag name
    if (typeid(*rhs_0) == typeid(Type_Selector))
    {
      // if rhs is universal, just return this tagname + rhs's qualifiers
      if (static_cast<Type_Selector*>(rhs_0)->name() == "*")
      {
        Compound_Selector* cpy = new (ctx.mem) Compound_Selector(rhs->pstate());
        (*cpy) << this;
        for (size_t i = 1, L = rhs->length(); i < L; ++i)
        { (*cpy) << (*rhs)[i]; }
        return cpy;
      }
      // if rhs is another tag name and it matches this, return rhs
      else if (static_cast<Type_Selector*>(rhs_0)->name() == name())
      { return new (ctx.mem) Compound_Selector(*rhs); }
      // else the tag names don't match; return nil
      else
      { return 0; }
    }
    // else it's a tag name and a bunch of qualifiers -- just append them
    Compound_Selector* cpy = new (ctx.mem) Compound_Selector(rhs->pstate());
    (*cpy) << this;
    (*cpy) += rhs;
    return cpy;
  }

  Compound_Selector* Selector_Qualifier::unify_with(Compound_Selector* rhs, Context& ctx)
  {
    if (name()[0] == '#')
    {
      for (size_t i = 0, L = rhs->length(); i < L; ++i)
      {
        Simple_Selector* rhs_i = (*rhs)[i];
        if (typeid(*rhs_i) == typeid(Selector_Qualifier) &&
            static_cast<Selector_Qualifier*>(rhs_i)->name()[0] == '#' &&
            static_cast<Selector_Qualifier*>(rhs_i)->name() != name())
          return 0;
      }
    }
    rhs->has_line_break(has_line_break());
    return Simple_Selector::unify_with(rhs, ctx);
  }

  Compound_Selector* Pseudo_Selector::unify_with(Compound_Selector* rhs, Context& ctx)
  {
    if (is_pseudo_element())
    {
      for (size_t i = 0, L = rhs->length(); i < L; ++i)
      {
        Simple_Selector* rhs_i = (*rhs)[i];
        if (typeid(*rhs_i) == typeid(Pseudo_Selector) &&
            static_cast<Pseudo_Selector*>(rhs_i)->is_pseudo_element() &&
            static_cast<Pseudo_Selector*>(rhs_i)->name() != name())
        { return 0; }
      }
    }
    return Simple_Selector::unify_with(rhs, ctx);
  }

  bool Compound_Selector::is_superselector_of(Compound_Selector* rhs)
  {
    To_String to_string;

    Simple_Selector* lbase = base();
    Simple_Selector* rbase = rhs->base();

    // Check if pseudo-elements are the same between the selectors

    set<string> lpsuedoset, rpsuedoset;
    for (size_t i = 0, L = length(); i < L; ++i)
    {
      if ((*this)[i]->is_pseudo_element()) {
        string pseudo((*this)[i]->perform(&to_string));
        pseudo = pseudo.substr(pseudo.find_first_not_of(":")); // strip off colons to ensure :after matches ::after since ruby sass is forgiving
        lpsuedoset.insert(pseudo);
      }
    }
    for (size_t i = 0, L = rhs->length(); i < L; ++i)
    {
      if ((*rhs)[i]->is_pseudo_element()) {
        string pseudo((*rhs)[i]->perform(&to_string));
        pseudo = pseudo.substr(pseudo.find_first_not_of(":")); // strip off colons to ensure :after matches ::after since ruby sass is forgiving
        rpsuedoset.insert(pseudo);
      }
    }
    if (lpsuedoset != rpsuedoset) {
      return false;
    }

    // Check the Simple_Selectors

    set<string> lset, rset;

    if (!lbase) // no lbase; just see if the left-hand qualifiers are a subset of the right-hand selector
    {
      for (size_t i = 0, L = length(); i < L; ++i)
      {
        Selector* lhs = (*this)[i];
        // very special case for wrapped matches selector
        if (Wrapped_Selector* wrapped = dynamic_cast<Wrapped_Selector*>(lhs)) {
          if (wrapped->name() == ":matches(" || wrapped->name() == ":-moz-any(") {
            if (Selector_List* list = dynamic_cast<Selector_List*>(wrapped->selector())) {
              if (Compound_Selector* comp = dynamic_cast<Compound_Selector*>(rhs)) {
                if (list->is_superselector_of(comp)) return true;
              }
            }
          }
        }
        // match from here on as strings
        lset.insert(lhs->perform(&to_string));
      }
      for (size_t i = 0, L = rhs->length(); i < L; ++i)
      { rset.insert((*rhs)[i]->perform(&to_string)); }
      return includes(rset.begin(), rset.end(), lset.begin(), lset.end());
    }
    else { // there's an lbase
      for (size_t i = 1, L = length(); i < L; ++i)
      { lset.insert((*this)[i]->perform(&to_string)); }
      if (rbase)
      {
        if (lbase->perform(&to_string) != rbase->perform(&to_string)) // if there's an rbase, make sure they match
        { return false; }
        else // the bases do match, so compare qualifiers
        {
          for (size_t i = 1, L = rhs->length(); i < L; ++i)
          { rset.insert((*rhs)[i]->perform(&to_string)); }
          return includes(rset.begin(), rset.end(), lset.begin(), lset.end());
        }
      }
    }
    // catch-all
    return false;
  }

  bool Compound_Selector::operator==(const Compound_Selector& rhs) const {
    To_String to_string;

    // Check if pseudo-elements are the same between the selectors

    set<string> lpsuedoset, rpsuedoset;
    for (size_t i = 0, L = length(); i < L; ++i)
    {
      if ((*this)[i]->is_pseudo_element()) {
        string pseudo((*this)[i]->perform(&to_string));
        pseudo = pseudo.substr(pseudo.find_first_not_of(":")); // strip off colons to ensure :after matches ::after since ruby sass is forgiving
        lpsuedoset.insert(pseudo);
      }
    }
    for (size_t i = 0, L = rhs.length(); i < L; ++i)
    {
      if (rhs[i]->is_pseudo_element()) {
        string pseudo(rhs[i]->perform(&to_string));
        pseudo = pseudo.substr(pseudo.find_first_not_of(":")); // strip off colons to ensure :after matches ::after since ruby sass is forgiving
        rpsuedoset.insert(pseudo);
      }
    }
    if (lpsuedoset != rpsuedoset) {
      return false;
    }

    // Check the base

    const Simple_Selector* const lbase = base();
    const Simple_Selector* const rbase = rhs.base();

    if ((lbase && !rbase) ||
      (!lbase && rbase) ||
      ((lbase && rbase) && (*lbase != *rbase))) {
      return false;
    }


    // Check the rest of the SimpleSelectors
    // Use string representations. We can't create a set of Simple_Selector pointers because std::set == std::set is going to call ==
    // on the pointers to determine equality. I don't know of a way to pass in a comparison object. The one you can specify as part of
    // the template type is used for ordering, but not equality. We also can't just put in non-pointer Simple_Selectors because the
    // class is intended to be subclassed, and we'd get splicing.

    set<string> lset, rset;

    for (size_t i = 0, L = length(); i < L; ++i)
    { lset.insert((*this)[i]->perform(&to_string)); }
    for (size_t i = 0, L = rhs.length(); i < L; ++i)
    { rset.insert(rhs[i]->perform(&to_string)); }

    return lset == rset;
  }

  bool Complex_Selector_Pointer_Compare::operator() (const Complex_Selector* const pLeft, const Complex_Selector* const pRight) const {
    return *pLeft < *pRight;
  }

  bool Complex_Selector::is_superselector_of(Compound_Selector* rhs)
  {
    return base()->is_superselector_of(rhs);
  }

  bool Complex_Selector::is_superselector_of(Complex_Selector* rhs)
  {
    Complex_Selector* lhs = this;
    To_String to_string;
    // check for selectors with leading or trailing combinators
    if (!lhs->head() || !rhs->head())
    { return false; }
    Complex_Selector* l_innermost = lhs->innermost();
    if (l_innermost->combinator() != Complex_Selector::ANCESTOR_OF && !l_innermost->tail())
    { return false; }
    Complex_Selector* r_innermost = rhs->innermost();
    if (r_innermost->combinator() != Complex_Selector::ANCESTOR_OF && !r_innermost->tail())
    { return false; }
    // more complex (i.e., longer) selectors are always more specific
    size_t l_len = lhs->length(), r_len = rhs->length();
    if (l_len > r_len)
    { return false; }

    if (l_len == 1)
    { return lhs->head()->is_superselector_of(rhs->base()); }

    // we have to look one tail deeper, since we cary the
    // combinator around for it (which is important here)
    if (rhs->tail() && lhs->tail() && combinator() != Complex_Selector::ANCESTOR_OF) {
      Complex_Selector* lhs_tail = lhs->tail();
      Complex_Selector* rhs_tail = rhs->tail();
      if (lhs_tail->combinator() != rhs_tail->combinator()) return false;
      if (!lhs_tail->head()->is_superselector_of(rhs_tail->head())) return false;
    }


    bool found = false;
    Complex_Selector* marker = rhs;
    for (size_t i = 0, L = rhs->length(); i < L; ++i) {
      if (i == L-1)
      { return false; }
      if (lhs->head()->is_superselector_of(marker->head()))
      { found = true; break; }
      marker = marker->tail();
    }
    if (!found)
    { return false; }

    /*
      Hmm, I hope I have the logic right:

      if lhs has a combinator:
        if !(marker has a combinator) return false
        if !(lhs.combinator == '~' ? marker.combinator != '>' : lhs.combinator == marker.combinator) return false
        return lhs.tail-without-innermost.is_superselector_of(marker.tail-without-innermost)
      else if marker has a combinator:
        if !(marker.combinator == ">") return false
        return lhs.tail.is_superselector_of(marker.tail)
      else
        return lhs.tail.is_superselector_of(marker.tail)
    */
    if (lhs->combinator() != Complex_Selector::ANCESTOR_OF)
    {
      if (marker->combinator() == Complex_Selector::ANCESTOR_OF)
      { return false; }
      if (!(lhs->combinator() == Complex_Selector::PRECEDES ? marker->combinator() != Complex_Selector::PARENT_OF : lhs->combinator() == marker->combinator()))
      { return false; }
      return lhs->tail()->is_superselector_of(marker->tail());
    }
    else if (marker->combinator() != Complex_Selector::ANCESTOR_OF)
    {
      if (marker->combinator() != Complex_Selector::PARENT_OF)
      { return false; }
      return lhs->tail()->is_superselector_of(marker->tail());
    }
    else
    {
      return lhs->tail()->is_superselector_of(marker->tail());
    }
    // catch-all
    return false;
  }

  size_t Complex_Selector::length()
  {
    // TODO: make this iterative
    if (!tail()) return 1;
    return 1 + tail()->length();
  }

  Compound_Selector* Complex_Selector::base()
  {
    if (!tail()) return head();
    else return tail()->base();
  }

  Complex_Selector* Complex_Selector::context(Context& ctx)
  {
    if (!tail()) return 0;
    if (!head()) return tail()->context(ctx);
    Complex_Selector* cpy = new (ctx.mem) Complex_Selector(pstate(), combinator(), head(), tail()->context(ctx));
    cpy->media_block(media_block());
    cpy->last_block(last_block());
    return cpy;
  }

  Complex_Selector* Complex_Selector::innermost()
  {
    if (!tail()) return this;
    else         return tail()->innermost();
  }

  Complex_Selector::Combinator Complex_Selector::clear_innermost()
  {
    Combinator c;
    if (!tail() || tail()->length() == 1)
    { c = combinator(); combinator(ANCESTOR_OF); tail(0); }
    else
    { c = tail()->clear_innermost(); }
    return c;
  }

  void Complex_Selector::set_innermost(Complex_Selector* val, Combinator c)
  {
    if (!tail())
    { tail(val); combinator(c); }
    else
    { tail()->set_innermost(val, c); }
  }

  Complex_Selector* Complex_Selector::clone(Context& ctx) const
  {
    Complex_Selector* cpy = new (ctx.mem) Complex_Selector(*this);
    if (tail()) cpy->tail(tail()->clone(ctx));
    return cpy;
  }

  Complex_Selector* Complex_Selector::cloneFully(Context& ctx) const
  {
    Complex_Selector* cpy = new (ctx.mem) Complex_Selector(*this);

    if (head()) {
      cpy->head(head()->clone(ctx));
    }

    if (tail()) {
      cpy->tail(tail()->cloneFully(ctx));
    }

    return cpy;
  }

  Compound_Selector* Compound_Selector::clone(Context& ctx) const
  {
    Compound_Selector* cpy = new (ctx.mem) Compound_Selector(*this);
    return cpy;
  }



  /* not used anymore - remove?
  Selector_Placeholder* Selector::find_placeholder()
  {
    return 0;
  }*/

  void Selector_List::adjust_after_pushing(Complex_Selector* c)
  {
    if (c->has_reference())   has_reference(true);
    if (c->has_placeholder()) has_placeholder(true);

#ifdef DEBUG
    To_String to_string;
    this->mCachedSelector(this->perform(&to_string));
#endif
  }

  // it's a superselector if every selector of the right side
  // list is a superselector of the given left side selector
  bool Complex_Selector::is_superselector_of(Selector_List *sub)
  {
    // Check every rhs selector against left hand list
    for(size_t i = 0, L = sub->length(); i < L; ++i) {
      if (!is_superselector_of((*sub)[i])) return false;
    }
    return true;
  }

  // it's a superselector if every selector of the right side
  // list is a superselector of the given left side selector
  bool Selector_List::is_superselector_of(Selector_List *sub)
  {
    // Check every rhs selector against left hand list
    for(size_t i = 0, L = sub->length(); i < L; ++i) {
      if (!is_superselector_of((*sub)[i])) return false;
    }
    return true;
  }

  // it's a superselector if every selector on the right side
  // is a superselector of any one of the left side selectors
  bool Selector_List::is_superselector_of(Compound_Selector *sub)
  {
    // Check every lhs selector against right hand
    for(size_t i = 0, L = length(); i < L; ++i) {
      if ((*this)[i]->is_superselector_of(sub)) return true;
    }
    return false;
  }

  // it's a superselector if every selector on the right side
  // is a superselector of any one of the left side selectors
  bool Selector_List::is_superselector_of(Complex_Selector *sub)
  {
    // Check every lhs selector against right hand
    for(size_t i = 0, L = length(); i < L; ++i) {
      if ((*this)[i]->is_superselector_of(sub)) return true;
    }
    return false;
  }

  /* not used anymore - remove?
  Selector_Placeholder* Selector_List::find_placeholder()
  {
    if (has_placeholder()) {
      for (size_t i = 0, L = length(); i < L; ++i) {
        if ((*this)[i]->has_placeholder()) return (*this)[i]->find_placeholder();
      }
    }
    return 0;
  }*/

  /* not used anymore - remove?
  Selector_Placeholder* Complex_Selector::find_placeholder()
  {
    if (has_placeholder()) {
      if (head() && head()->has_placeholder()) return head()->find_placeholder();
      else if (tail() && tail()->has_placeholder()) return tail()->find_placeholder();
    }
    return 0;
  }*/

  /* not used anymore - remove?
  Selector_Placeholder* Compound_Selector::find_placeholder()
  {
    if (has_placeholder()) {
      for (size_t i = 0, L = length(); i < L; ++i) {
        if ((*this)[i]->has_placeholder()) return (*this)[i]->find_placeholder();
      }
      // return this;
    }
    return 0;
  }*/

  /* not used anymore - remove?
  Selector_Placeholder* Selector_Placeholder::find_placeholder()
  {
    return this;
  }*/

  vector<string> Compound_Selector::to_str_vec()
  {
    To_String to_string;
    vector<string> result;
    result.reserve(length());
    for (size_t i = 0, L = length(); i < L; ++i)
    { result.push_back((*this)[i]->perform(&to_string)); }
    return result;
  }

  Compound_Selector* Compound_Selector::minus(Compound_Selector* rhs, Context& ctx)
  {
    To_String to_string(&ctx);
    Compound_Selector* result = new (ctx.mem) Compound_Selector(pstate());

    // not very efficient because it needs to preserve order
    for (size_t i = 0, L = length(); i < L; ++i)
    {
      bool found = false;
      string thisSelector((*this)[i]->perform(&to_string));
      for (size_t j = 0, M = rhs->length(); j < M; ++j)
      {
        if (thisSelector == (*rhs)[j]->perform(&to_string))
        {
          found = true;
          break;
        }
      }
      if (!found) (*result) << (*this)[i];
    }

    return result;
  }

  void Compound_Selector::mergeSources(SourcesSet& sources, Context& ctx)
  {
    for (SourcesSet::iterator iterator = sources.begin(), endIterator = sources.end(); iterator != endIterator; ++iterator) {
      this->sources_.insert((*iterator)->clone(ctx));
    }
  }

  /* not used anymore - remove?
  vector<Compound_Selector*> Complex_Selector::to_vector()
  {
    vector<Compound_Selector*> result;
    Compound_Selector* h = head();
    Complex_Selector* t = tail();
    if (h) result.push_back(h);
    while (t)
    {
      h = t->head();
      t = t->tail();
      if (h) result.push_back(h);
    }
    return result;
  }*/

  Number::Number(ParserState pstate, double val, string u, bool zero)
  : Expression(pstate),
    value_(val),
    zero_(zero),
    numerator_units_(vector<string>()),
    denominator_units_(vector<string>()),
    hash_(0)
  {
    size_t l = 0, r = 0;
    if (!u.empty()) {
      bool nominator = true;
      while (true) {
        r = u.find_first_of("*/", l);
        string unit(u.substr(l, r - l));
        if (nominator) numerator_units_.push_back(unit);
        else denominator_units_.push_back(unit);
        if (u[r] == '/') nominator = false;
        if (r == string::npos) break;
        else l = r + 1;
      }
    }
    concrete_type(NUMBER);
  }

  string Number::unit() const
  {
    stringstream u;
    for (size_t i = 0, S = numerator_units_.size(); i < S; ++i) {
      if (i) u << '*';
      u << numerator_units_[i];
    }
    if (!denominator_units_.empty()) u << '/';
    for (size_t i = 0, S = denominator_units_.size(); i < S; ++i) {
      if (i) u << '*';
      u << denominator_units_[i];
    }
    return u.str();
  }

  bool Number::is_unitless()
  { return numerator_units_.empty() && denominator_units_.empty(); }

  void Number::normalize(const string& prefered)
  {

    // first make sure same units cancel each other out
    // it seems that a map table will fit nicely to do this
    // we basically construct exponents for each unit
    // has the advantage that they will be pre-sorted
    map<string, int> exponents;

    // initialize by summing up occurences in unit vectors
    for (size_t i = 0, S = numerator_units_.size(); i < S; ++i) ++ exponents[numerator_units_[i]];
    for (size_t i = 0, S = denominator_units_.size(); i < S; ++i) -- exponents[denominator_units_[i]];

    // the final conversion factor
    double factor = 1;

    // get the first entry of numerators
    // forward it when entry is converted
    vector<string>::iterator nom_it = numerator_units_.begin();
    vector<string>::iterator nom_end = numerator_units_.end();
    vector<string>::iterator denom_it = denominator_units_.begin();
    vector<string>::iterator denom_end = denominator_units_.end();

    // main normalization loop
    // should be close to optimal
    while (denom_it != denom_end)
    {
      // get and increment afterwards
      const string denom = *(denom_it ++);
      // skip already canceled out unit
      if (exponents[denom] >= 0) continue;
      // skip all units we don't know how to convert
      if (string_to_unit(denom) == UNKNOWN) continue;
      // now search for nominator
      while (nom_it != nom_end)
      {
        // get and increment afterwards
        const string nom = *(nom_it ++);
        // skip already canceled out unit
        if (exponents[nom] <= 0) continue;
        // skip all units we don't know how to convert
        if (string_to_unit(nom) == UNKNOWN) continue;
        // we now have two convertable units
        // add factor for current conversion
        factor *= conversion_factor(nom, denom);
        // update nominator/denominator exponent
        -- exponents[nom]; ++ exponents[denom];
        // inner loop done
        break;
      }
    }

    // now we can build up the new unit arrays
    numerator_units_.clear();
    denominator_units_.clear();

    // build them by iterating over the exponents
    for (auto exp : exponents)
    {
      // maybe there is more effecient way to push
      // the same item multiple times to a vector?
      for(size_t i = 0, S = abs(exp.second); i < S; ++i)
      {
        // opted to have these switches in the inner loop
        // makes it more readable and should not cost much
        if (exp.second < 0) denominator_units_.push_back(exp.first);
        else if (exp.second > 0) numerator_units_.push_back(exp.first);
      }
    }

    // apply factor to value_
    // best precision this way
    value_ *= factor;

    // maybe convert to other unit
    // easier implemented on its own
    try { convert(prefered); }
    catch (incompatibleUnits& err)
    { error(err.what(), pstate()); }
    catch (...) { throw; }

  }

  void Number::convert(const string& prefered)
  {
    // abort if unit is empty
    if (prefered.empty()) return;

    // first make sure same units cancel each other out
    // it seems that a map table will fit nicely to do this
    // we basically construct exponents for each unit
    // has the advantage that they will be pre-sorted
    map<string, int> exponents;

    // initialize by summing up occurences in unit vectors
    for (size_t i = 0, S = numerator_units_.size(); i < S; ++i) ++ exponents[numerator_units_[i]];
    for (size_t i = 0, S = denominator_units_.size(); i < S; ++i) -- exponents[denominator_units_[i]];

    // the final conversion factor
    double factor = 1;

    vector<string>::iterator denom_it = denominator_units_.begin();
    vector<string>::iterator denom_end = denominator_units_.end();

    // main normalization loop
    // should be close to optimal
    while (denom_it != denom_end)
    {
      // get and increment afterwards
      const string denom = *(denom_it ++);
      // check if conversion is needed
      if (denom == prefered) continue;
      // skip already canceled out unit
      if (exponents[denom] >= 0) continue;
      // skip all units we don't know how to convert
      if (string_to_unit(denom) == UNKNOWN) continue;
      // we now have two convertable units
      // add factor for current conversion
      factor *= conversion_factor(denom, prefered);
      // update nominator/denominator exponent
      ++ exponents[denom]; -- exponents[prefered];
    }

    vector<string>::iterator nom_it = numerator_units_.begin();
    vector<string>::iterator nom_end = numerator_units_.end();

    // now search for nominator
    while (nom_it != nom_end)
    {
      // get and increment afterwards
      const string nom = *(nom_it ++);
      // check if conversion is needed
      if (nom == prefered) continue;
      // skip already canceled out unit
      if (exponents[nom] <= 0) continue;
      // skip all units we don't know how to convert
      if (string_to_unit(nom) == UNKNOWN) continue;
      // we now have two convertable units
      // add factor for current conversion
      factor *= conversion_factor(nom, prefered);
      // update nominator/denominator exponent
      -- exponents[nom]; ++ exponents[prefered];
    }

    // now we can build up the new unit arrays
    numerator_units_.clear();
    denominator_units_.clear();

    // build them by iterating over the exponents
    for (auto exp : exponents)
    {
      // maybe there is more effecient way to push
      // the same item multiple times to a vector?
      for(size_t i = 0, S = abs(exp.second); i < S; ++i)
      {
        // opted to have these switches in the inner loop
        // makes it more readable and should not cost much
        if (exp.second < 0) denominator_units_.push_back(exp.first);
        else if (exp.second > 0) numerator_units_.push_back(exp.first);
      }
    }

    // apply factor to value_
    // best precision this way
    value_ *= factor;

  }

  // useful for making one number compatible with another
  string Number::find_convertible_unit() const
  {
    for (size_t i = 0, S = numerator_units_.size(); i < S; ++i) {
      string u(numerator_units_[i]);
      if (string_to_unit(u) != UNKNOWN) return u;
    }
    for (size_t i = 0, S = denominator_units_.size(); i < S; ++i) {
      string u(denominator_units_[i]);
      if (string_to_unit(u) != UNKNOWN) return u;
    }
    return string();
  }


  bool Number::operator== (Expression* rhs) const
  {
    try
    {
      Number l(pstate_, value_, unit());
      Number& r = dynamic_cast<Number&>(*rhs);
      l.normalize(find_convertible_unit());
      r.normalize(find_convertible_unit());
      return l.unit() == r.unit() &&
             l.value() == r.value();
    }
    catch (std::bad_cast&) {}
    catch (...) { throw; }
    return false;
  }

  bool Number::operator== (Expression& rhs) const
  {
    return operator==(&rhs);
  }

  bool List::operator==(Expression* rhs) const
  {
    try
    {
      List* r = dynamic_cast<List*>(rhs);
      if (!r || length() != r->length()) return false;
      if (separator() != r->separator()) return false;
      for (size_t i = 0, L = r->length(); i < L; ++i)
        if (*elements()[i] != *(*r)[i]) return false;
      return true;
    }
    catch (std::bad_cast&) {}
    catch (...) { throw; }
    return false;
  }

  bool List::operator== (Expression& rhs) const
  {
    return operator==(&rhs);
  }

  size_t List::size() const {
    if (!is_arglist_) return length();
    // arglist expects a list of arguments
    // so we need to break before keywords
    for (size_t i = 0, L = length(); i < L; ++i) {
      if (Argument* arg = dynamic_cast<Argument*>((*this)[i])) {
        if (!arg->name().empty()) return i;
      }
    }
    return length();
  }

  Expression* Hashed::at(Expression* k) const
  {
    if (elements_.count(k))
    { return elements_.at(k); }
    else { return &sass_null; }
  }

}

