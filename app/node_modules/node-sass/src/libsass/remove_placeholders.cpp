#include "remove_placeholders.hpp"
#include "context.hpp"
#include "inspect.hpp"
#include "to_string.hpp"
#include <iostream>

namespace Sass {

    Remove_Placeholders::Remove_Placeholders(Context& ctx)
    : ctx(ctx)
    { }

    template<typename T>
    void Remove_Placeholders::clean_selector_list(T r) {

        // Create a new selector group without placeholders
        Selector_List* sl = static_cast<Selector_List*>(r->selector());

        if (sl) {
            Selector_List* new_sl = new (ctx.mem) Selector_List(sl->pstate());

            for (size_t i = 0, L = sl->length(); i < L; ++i) {
                if (!(*sl)[i]->has_placeholder()) {
                    *new_sl << (*sl)[i];
                }
            }

            // Set the new placeholder selector list
            r->selector(new_sl);
        }

        // Iterate into child blocks
        Block* b = r->block();

        for (size_t i = 0, L = b->length(); i < L; ++i) {
            Statement* stm = (*b)[i];
            stm->perform(this);
        }
    }

    void Remove_Placeholders::operator()(Block* b) {
        for (size_t i = 0, L = b->length(); i < L; ++i) {
            (*b)[i]->perform(this);
        }
    }

    void Remove_Placeholders::operator()(Ruleset* r) {
        clean_selector_list(r);
    }

    void Remove_Placeholders::operator()(Media_Block* m) {
        clean_selector_list(m);
    }

    void Remove_Placeholders::operator()(At_Rule* a) {
        if (a->block()) a->block()->perform(this);
    }

}
