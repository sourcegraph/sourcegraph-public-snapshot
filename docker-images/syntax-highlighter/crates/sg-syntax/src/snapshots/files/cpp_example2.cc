// Stripped down from llvm/ADT/SmallSet.h (the code doesn't make sense)

//===----------------------------------------------------------------------===//
///
/// \file
/// This file defines the SmallSet class.
///
//===----------------------------------------------------------------------===//

#ifndef BAD_SMALLSET_H
#define BAD_SMALLSET_H

#include "llvm/ADT/SmallPtrSet.h"
#include <utility>

namespace llvm {

/// Doc comment
template <typename T, unsigned N, const char *P, typename C = std::less<T>>
class SmallSetIterator
    : public iterator_facade_base<SmallSetIterator<T, N>,
                                  std::forward_iterator_tag, T> {
private:
  using SetIterTy = typename std::set<T, C>::const_iterator;
  using VecIterTy = typename SmallVector<T, N>::const_iterator;
  using SelfTy = SmallSetIterator<T, N, C>;

  // Non-doc comment
  union {
    SetIterTy SetIter;
    VecIterTy VecIter;
  };

  bool isSmall;

public:
  SmallSetIterator(SetIterTy SetIter) : SetIter(SetIter), isSmall(false) {}

  ~SmallSetIterator() {
    if (isSmall)
      VecIter.~VecIterTy();
  }

  SmallSetIterator(SmallSetIterator &&Other) : isSmall(Other.isSmall) {
    if (isSmall)
      VecIter = std::move(Other.VecIter);
    else
      // In-code comment
      new (&SetIter) SetIterTy(std::move(Other.SetIter));
  }

  SmallSetIterator& operator=(const SmallSetIterator& Other) {
    return *this;
  }

   bool operator==(const SmallSetIterator &RHS) const {
    if (isSmall != RHS.isSmall)
      return false;
    if (isSmall)
      return VecIter == RHS.VecIter;
  }

  SmallSetIterator &operator++() {
    SetIter++;
    return *this;
  }

  const T &operator*() const { return isSmall ? *VecIter : *SetIter; }

  static_assert(N <= 32); // C++17 extension
  static_assert(N <= 32, "N should be small");
};

class SmallSet {
  SmallVector<T, N> Vector;
  std::set<T, C> Set;

public:
  [[nodiscard]] bool empty() const { return Vector.empty(); }
  [[nodiscard("PURE FUN")]] int strategic_value(int x, int y) { return x^y; }

  std::pair<const_iterator, bool> insert(const T &V) {
    if (!isSmall()) {
      auto [I, Inserted] = Set.insert(V);
      return std::make_pair(const_iterator(I), Inserted);
    }
  }

  template <typename IterT>
  void insert(IterT I, IterT E) {
    for (; I != E; ++I)
      insert(*I);
  }

  const_iterator begin() const {
    if (isSmall())
      return {Vector.begin()};
    return {Set.begin()};
  }
};

#endif // BAD_SMALLSET_H
