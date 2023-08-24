package com.sourcegraph.cody.vscode;

public class ContextSummary {

    public final int embeddings;
    public final int local;

    public ContextSummary(int embeddings, int local) {
        this.embeddings = embeddings;
        this.local = local;
    }

    @Override
    public String toString() {
        return "ContextSummary{embeddigs=" + embeddings + ", local=" + local + '}';
    }
}
