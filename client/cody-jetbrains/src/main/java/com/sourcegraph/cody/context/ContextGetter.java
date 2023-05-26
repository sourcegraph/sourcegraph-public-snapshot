package com.sourcegraph.cody.context;

import com.sourcegraph.cody.context.embeddings.EmbeddingsSearcher;
import org.jetbrains.annotations.NotNull;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

// TODO: Use this class to get context
public class ContextGetter {
    private final @NotNull String codebase;

    public ContextGetter(@NotNull String codebase) {
        this.codebase = codebase;
    }

    public @NotNull List<ContextMessage> getContextMessages(@NotNull String query, int codeResultCount, int textResultCount, @NotNull String useContext) throws IOException {
        if (useContext.equals("embeddings")) {
            return EmbeddingsSearcher.getContextMessages(codebase, query, codeResultCount, textResultCount);
        } else {
            // TODO: Add keyword search if embeddings are not available
            //return KeywordSearcher.getContextMessages(query, codeResultCount, textResultCount);
            return new ArrayList<>();
        }
    }
}
