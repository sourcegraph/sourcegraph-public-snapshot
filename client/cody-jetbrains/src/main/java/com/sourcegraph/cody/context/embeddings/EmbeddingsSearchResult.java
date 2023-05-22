package com.sourcegraph.cody.context.embeddings;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class EmbeddingsSearchResult {
    private final @Nullable String repoName;
    private final @Nullable String revision;
    private final @NotNull String fileName;
    private final int startLine;
    private final int endLine;
    private final @NotNull String content;

    public EmbeddingsSearchResult(@Nullable String repoName, @Nullable String revision, @NotNull String fileName, int startLine, int endLine, @NotNull String content) {
        this.repoName = repoName;
        this.revision = revision;
        this.fileName = fileName;
        this.startLine = startLine;
        this.endLine = endLine;
        this.content = content;
    }

    public @Nullable String getRepoName() {
        return repoName;
    }

    public @Nullable String getRevision() {
        return revision;
    }

    public @NotNull String getFileName() {
        return fileName;
    }

    public int getStartLine() {
        return startLine;
    }

    public int getEndLine() {
        return endLine;
    }

    public @NotNull String getContent() {
        return content;
    }
}
