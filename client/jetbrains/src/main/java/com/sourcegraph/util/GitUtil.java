package com.sourcegraph.util;

import java.io.IOException;

public class GitUtil {
    /**
     * E.g. "origin" -> "git@github.com:foo/bar"
     */
    public static String getRemoteUrl(String repoDirectoryPath, String remoteName) throws Exception {
        String result = SourcegraphUtil.exec("git remote get-url " + remoteName, repoDirectoryPath).trim();
        if (result.isEmpty()) {
            throw new Exception("There is no such remote: \"" + remoteName + "\".");
        }
        return result;
    }

    /**
     * Returns the URL of the "sourcegraph" remote.
     * Falls back to the "origin" remote.
     * An exception is thrown if neither exists.
     */
    public static String getConfiguredRemoteUrl(String repoDirectoryPath) throws Exception {
        try {
            return getRemoteUrl(repoDirectoryPath, "sourcegraph");
        } catch (Exception e) {
            try {
                return getRemoteUrl(repoDirectoryPath, "origin");
            } catch (Exception e2) {
                throw new Exception("No configured git remote for \"sourcegraph\" or \"origin\".");
            }
        }
    }

    /**
     * Returns the repository root directory for any path within a repository.
     */
    public static String getRepoRootPath(String path) throws IOException {
        return SourcegraphUtil.exec("git rev-parse --show-toplevel", path).trim();
    }

    /**
     * Returns the current branch name of the repository.
     * In detached HEAD state and other exceptional cases it returns "HEAD".
     */
    public static String getCurrentBranchName(String path) throws IOException {
        return SourcegraphUtil.exec("git rev-parse --abbrev-ref HEAD", path).trim();
    }

    /**
     * @param branchName E.g. "main"
     */
    public static boolean doesRemoteBranchExist(String branchName, String repoDirectoryPath) throws IOException {
        return SourcegraphUtil.exec("git show-branch remotes/origin/" + branchName, repoDirectoryPath).length() > 0;
    }
}
