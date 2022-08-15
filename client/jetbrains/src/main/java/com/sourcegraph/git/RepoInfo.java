package com.sourcegraph.git;

import org.jetbrains.annotations.NotNull;

public class RepoInfo {
    @NotNull
    public final String relativePath; // E.g. "/client/jetbrains/package.json"
    @NotNull
    public final String remoteUrl; // E.g. "git@github.com:sourcegraph/sourcegraph"
    @NotNull
    public final String branchName; // E.g. "main"

    public RepoInfo(@NotNull String relativePath, @NotNull String remoteUrl, @NotNull String branchName) {
        this.relativePath = relativePath;
        this.remoteUrl = remoteUrl;
        this.branchName = branchName;
    }
}
