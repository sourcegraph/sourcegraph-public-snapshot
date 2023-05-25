package com.sourcegraph.cody.context.embeddings;

import com.sourcegraph.cody.context.ContextFile;
import org.jetbrains.annotations.NotNull;

import java.util.List;

public class GroupedResults {
    private final @NotNull ContextFile file;
    private final @NotNull List<String> snippets;

    public GroupedResults(@NotNull ContextFile file, @NotNull List<String> snippets) {
        this.file = file;
        this.snippets = snippets;
    }

    public @NotNull ContextFile getFile() {
        return file;
    }

    public @NotNull List<String> getSnippets() {
        return snippets;
    }
}
