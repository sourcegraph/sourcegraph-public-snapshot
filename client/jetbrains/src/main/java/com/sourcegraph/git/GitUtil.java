package com.sourcegraph.git;

import com.intellij.dvcs.repo.Repository;
import com.intellij.dvcs.repo.VcsRepositoryManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.vcsUtil.VcsUtil;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStreamReader;

public class GitUtil {
    // repoInfo returns the Sourcegraph repository URI, and the file path
    // relative to the repository root. If the repository URI cannot be
    // determined, a RepoInfo with empty strings is returned.
    @NotNull
    public static RepoInfo getRepoInfo(@NotNull VirtualFile file, @NotNull Project project) {
        String relativePath = "";
        String remoteUrl = "";
        String branchName = "";
        try {
            String repoRootPath = getRepoRootPath(project, file);
            if (repoRootPath == null) {
                return new RepoInfo("", "", "");
            }

            // Determine file path, relative to repository root.
            relativePath = file.getPath().length() > repoRootPath.length()
                ? file.getPath().substring(repoRootPath.length() + 1) : "";

            // If the current branch doesn't exist on the remote, use the default branch.
            branchName = getCurrentBranchName(repoRootPath);
            if (!doesRemoteBranchExist(branchName, repoRootPath)) {
                branchName = ConfigUtil.getDefaultBranchName(project);
            }

            remoteUrl = getRemoteUrl(file, project);
            String r = ConfigUtil.getRemoteUrlReplacements(project);
            String[] replacements = r.trim().split("\\s*,\\s*");
            // Check if the entered values are pairs
            for (int i = 0; i < replacements.length && replacements.length % 2 == 0; i += 2) {
                remoteUrl = remoteUrl.replace(replacements[i], replacements[i + 1]);
            }
        } catch (Exception err) {
            Logger.getInstance(GitUtil.class).info(err);
            err.printStackTrace();
        }
        return new RepoInfo(relativePath, remoteUrl, branchName);
    }

    @NotNull
    public static String getRemoteUrl(@NotNull VirtualFile file, @NotNull Project project) throws Exception {
        Repository repository = VcsRepositoryManager.getInstance(project).getRepositoryForFile(file);
        if (repository == null) {
            //String currentBranchName = repository.getCurrentBranchName();
            throw new Exception("No repository found.");
        }

        String result = exec("git remote get-url " + "sourcegraph", repository.getRoot().getPath()).trim();
        if (result.isEmpty()) {
            result = exec("git remote get-url " + "origin", repository.getRoot().getPath()).trim();
        }
        if (result.isEmpty()) {
            throw new Exception("No configured git remote for \"sourcegraph\" or \"origin\".");
        }

        return result;
    }

    /**
     * Returns the repository root directory for any path within a repository.
     */
    @Nullable
    private static String getRepoRootPath(@NotNull Project project, @NotNull VirtualFile file) {
        VirtualFile vcsRoot = VcsUtil.getVcsRootFor(project, file);
        return vcsRoot != null ? vcsRoot.getPath() : null;
    }

    /**
     * Returns the current branch name of the repository.
     * In detached HEAD state and other exceptional cases it returns "HEAD".
     */
    @NotNull
    private static String getCurrentBranchName(@NotNull String path) throws IOException {
        return exec("git rev-parse --abbrev-ref HEAD", path).trim();
    }

    /**
     * @param branchName E.g. "main"
     */
    private static boolean doesRemoteBranchExist(@NotNull String branchName, @NotNull String repoDirectoryPath) throws IOException {
        return exec("git show-branch remotes/origin/" + branchName, repoDirectoryPath).length() > 0;
    }

    // exec executes the given command in the specified directory and returns
    // its stdout. Any stderr output is logged.
    private static String exec(@NotNull String command, @NotNull String directoryPath) throws IOException {
        Logger.getInstance(GitUtil.class).debug("exec cmd='" + command + "' dir=" + directoryPath);

        // Create the process.
        Process process = Runtime.getRuntime().exec(command, null, new File(directoryPath));
        BufferedReader stdout = new BufferedReader(new InputStreamReader(process.getInputStream()));
        BufferedReader stderr = new BufferedReader(new InputStreamReader(process.getErrorStream()));

        // Log any stderr output.
        Logger logger = Logger.getInstance(GitUtil.class);
        String s;
        while ((s = stderr.readLine()) != null) {
            logger.debug(s);
        }

        String out = "";
        //noinspection StatementWithEmptyBody
        for (String line; (line = stdout.readLine()) != null; out += line + "\n") ;
        return out;
    }
}
