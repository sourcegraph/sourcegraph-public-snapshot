package com.sourcegraph.cody.vscode;

import javax.annotation.Nullable;

public class CompletionEvent {

    public final Params params;

    public CompletionEvent(Params params) {
        this.params = params;
    }

    @Override
    public String toString() {
        return "CompletionEvent{params=" + params + '}';
    }

    static class Params {

        @Nullable
        public final ContextSummary contextSummary;

        public Params(ContextSummary contextSummary) {
            this.contextSummary = contextSummary;
        }

        @Override
        public String toString() {
            return "Params{contextSummary=" + contextSummary + '}';
        }
    }
}
