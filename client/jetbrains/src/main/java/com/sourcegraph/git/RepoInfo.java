package com.sourcegraph.git;

public class RepoInfo {
    public final String relativePath;
    public final String remoteUrl;
    public final String branchName;

    public RepoInfo(String relativePath, String remoteUrl, String branchName) {
        this.relativePath = relativePath;
        this.remoteUrl = remoteUrl;
        this.branchName = branchName;
    }
}
