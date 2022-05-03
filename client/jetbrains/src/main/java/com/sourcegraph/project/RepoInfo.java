package com.sourcegraph.project;

public class RepoInfo {
    public final String fileRel;
    public final String remoteURL;
    public final String branch;

    public RepoInfo(String sFileRel, String sRemoteURL, String sBranch) {
        fileRel = sFileRel;
        remoteURL = sRemoteURL;
        branch = sBranch;
    }
}
