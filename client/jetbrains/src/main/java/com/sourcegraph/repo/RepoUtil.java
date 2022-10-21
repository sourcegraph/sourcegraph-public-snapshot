package com.sourcegraph.repo;

import com.intellij.dvcs.repo.Repository;
import com.intellij.dvcs.repo.VcsRepositoryManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.vcsUtil.VcsUtil;
import com.sourcegraph.config.ConfigUtil;
import git4idea.repo.GitRepository;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class RepoUtil {
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
            String localBranchName = getLocalBranchName(file, project);
            branchName = localBranchName != null && doesRemoteBranchExist(file, project, localBranchName)
                ? localBranchName : ConfigUtil.getDefaultBranchName(project);

            remoteUrl = getRemoteRepoUrl(file, project);
            String r = ConfigUtil.getRemoteUrlReplacements(project);
            String[] replacements = r.trim().split("\\s*,\\s*");
            // Check if the entered values are pairs
            for (int i = 0; i < replacements.length && replacements.length % 2 == 0; i += 2) {
                remoteUrl = remoteUrl.replace(replacements[i], replacements[i + 1]);
            }
        } catch (Exception err) {
            Logger.getInstance(RepoUtil.class).info(err);
            err.printStackTrace();
        }
        return new RepoInfo(relativePath, remoteUrl, branchName);
    }

    @NotNull
    public static String getRemoteRepoUrl(@NotNull VirtualFile file, @NotNull Project project) throws Exception {
        Repository repository = VcsRepositoryManager.getInstance(project).getRepositoryForFile(file);
        if (repository == null) {
            throw new Exception("Could not find repository for file " + file.getPath());
        }
        if (repository instanceof GitRepository) {
            return GitUtil.getRemoteRepoUrl((GitRepository) repository, project);
        }
        throw new Exception("Unsupported VCS: " + repository.getVcs().getName());
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
    @Nullable
    private static String getLocalBranchName(@NotNull VirtualFile file, @NotNull Project project) {
        Repository repository = VcsRepositoryManager.getInstance(project).getRepositoryForFile(file);
        return repository != null ? repository.getCurrentBranchName() : null;
    }

    /**
     * @param branchName E.g. "main"
     */
    private static boolean doesRemoteBranchExist(@NotNull VirtualFile file, @NotNull Project project, @NotNull String branchName) {
        Repository repository = VcsRepositoryManager.getInstance(project).getRepositoryForFile(file);
        if (repository == null) {
            return false;
        }

        if (repository instanceof GitRepository) {
            return GitUtil.doesRemoteBranchExist((GitRepository) repository, branchName);
        }

        // Unknown VCS.
        return false;
    }
}
