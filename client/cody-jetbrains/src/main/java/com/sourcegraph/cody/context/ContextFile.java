package com.sourcegraph.cody.context;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ContextFile {
    private final @NotNull String fileName;
    private final @Nullable String repoName;
    private final @Nullable String revision;

    public ContextFile(@NotNull String fileName, @Nullable String repoName, @Nullable String revision) {
        this.fileName = fileName;
        this.repoName = repoName;
        this.revision = revision;
    }

    public @NotNull String getFileName() {
        return fileName;
    }

    public @Nullable String getRepoName() {
        return repoName;
    }

    public @Nullable String getRevision() {
        return revision;
    }
}
