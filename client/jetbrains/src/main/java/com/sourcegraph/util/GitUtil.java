package com.sourcegraph.util;

import com.intellij.openapi.diagnostic.Logger;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStreamReader;

public class GitUtil {
    /**
     * E.g. "origin" -> "git@github.com:foo/bar"
     */
    public static String getRemoteUrl(String repoDirectoryPath, String remoteName) throws Exception {
        String result = exec("git remote get-url " + remoteName, repoDirectoryPath).trim();
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
        return exec("git rev-parse --show-toplevel", path).trim();
    }

    /**
     * Returns the current branch name of the repository.
     * In detached HEAD state and other exceptional cases it returns "HEAD".
     */
    public static String getCurrentBranchName(String path) throws IOException {
        return exec("git rev-parse --abbrev-ref HEAD", path).trim();
    }

    /**
     * @param branchName E.g. "main"
     */
    public static boolean doesRemoteBranchExist(String branchName, String repoDirectoryPath) throws IOException {
        return exec("git show-branch remotes/origin/" + branchName, repoDirectoryPath).length() > 0;
    }

    // exec executes the given command in the specified directory and returns
    // its stdout. Any stderr output is logged.
    private static String exec(String command, String directoryPath) throws IOException {
        Logger.getInstance(SourcegraphUtil.class).debug("exec cmd='" + command + "' dir=" + directoryPath);

        // Create the process.
        Process p = Runtime.getRuntime().exec(command, null, new File(directoryPath));
        BufferedReader stdout = new BufferedReader(new InputStreamReader(p.getInputStream()));
        BufferedReader stderr = new BufferedReader(new InputStreamReader(p.getErrorStream()));

        // Log any stderr output.
        Logger logger = Logger.getInstance(SourcegraphUtil.class);
        String s;
        while ((s = stderr.readLine()) != null) {
            logger.debug(s);
        }

        String out = "";
        //noinspection StatementWithEmptyBody
        for (String l; (l = stdout.readLine()) != null; out += l + "\n") ;
        return out;
    }
}
