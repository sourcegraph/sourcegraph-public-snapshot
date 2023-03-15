package com.sourcegraph.vcs;

import org.jetbrains.annotations.NotNull;

public class RepoInfo {
    @NotNull
    public final VCSType vcsType;
    @NotNull
    public final String remoteUrl; // E.g. "git@github.com:sourcegraph/sourcegraph.git", with replacements already applied
    @NotNull
    public final String remoteBranchName; // E.g. "main"
    @NotNull
    public final String relativePath; // E.g. "/client/jetbrains/package.json"

    public RepoInfo(@NotNull VCSType vcsType, @NotNull String remoteUrl, @NotNull String remoteBranchName, @NotNull String relativePath) {
        this.vcsType = vcsType;
        this.remoteUrl = remoteUrl;
        this.remoteBranchName = remoteBranchName;
        this.relativePath = relativePath;
    }

    // E.g. "sourecegraph/sourcegraph"
    @NotNull
    public String getRepoName() {
        int colonIndex = remoteUrl.lastIndexOf(":");
        int dotIndex = remoteUrl.lastIndexOf(".");
        return remoteUrl.substring(colonIndex + 1, (dotIndex == -1 || colonIndex > dotIndex) ? remoteUrl.length() : dotIndex);
    }

    // E.g. "github.com"
    @NotNull
    public String getCodeHostUrl() {
        int atIndex = remoteUrl.indexOf("@");
        int colonIndex = remoteUrl.lastIndexOf(":");
        return remoteUrl.substring((atIndex == -1 && atIndex < colonIndex) ? 0 : atIndex + 1, colonIndex);
    }
}
