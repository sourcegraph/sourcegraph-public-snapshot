package com.sourcegraph.cody.context.keyword;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class KeywordContextFetcherResult {
    private final @Nullable String repoName;
    private final @Nullable String revision;
    private final @NotNull String fileName;
    private final @NotNull String content;

    public KeywordContextFetcherResult(@Nullable String repoName, @Nullable String revision, @NotNull String fileName, @NotNull String content) {
        this.repoName = repoName;
        this.revision = revision;
        this.fileName = fileName;
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

    public @NotNull String getContent() {
        return content;
    }
}
