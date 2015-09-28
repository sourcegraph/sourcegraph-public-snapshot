#ifndef SASS_REMOVE_PLACEHOLDERS_H
#define SASS_REMOVE_PLACEHOLDERS_H

#pragma once

#include <iostream>

#include "ast.hpp"
#include "operation.hpp"

namespace Sass {

    using namespace std;

    class Context;

    class Remove_Placeholders : public Operation_CRTP<void, Remove_Placeholders> {

        Context&          ctx;

        void fallback_impl(AST_Node* n) {};

    public:
        Remove_Placeholders(Context&);
        virtual ~Remove_Placeholders() { }

        using Operation<void>::operator();

        void operator()(Block*);
        void operator()(Ruleset*);
        void operator()(Media_Block*);
        void operator()(At_Rule*);

        template <typename T>
        void clean_selector_list(T r);

        template <typename U>
        void fallback(U x) { return fallback_impl(x); }
    };

}

#endif
