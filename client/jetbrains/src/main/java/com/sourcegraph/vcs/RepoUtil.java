package com.sourcegraph.vcs;

import com.intellij.dvcs.repo.Repository;
import com.intellij.dvcs.repo.VcsRepositoryManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.vcsUtil.VcsUtil;
import com.sourcegraph.common.ErrorNotification;
import com.sourcegraph.config.ConfigUtil;
import git4idea.repo.GitRepository;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;
import org.jetbrains.idea.perforce.perforce.PerforceAuthenticationException;
import org.jetbrains.idea.perforce.perforce.PerforceSettings;

import java.io.File;
import java.util.Optional;

public class RepoUtil {
    // repoInfo returns the Sourcegraph repository URI, and the file path
    // relative to the repository root. If the repository URI cannot be
    // determined, a RepoInfo with empty strings is returned.
    @NotNull
    public static RepoInfo getRepoInfo(@NotNull Project project, @NotNull VirtualFile file) {
        VCSType vcsType = getVcsType(project, file);
        String relativePath = "";
        String remoteUrl = "";
        String remoteBranchName = "";
        try {
            String repoRootPath = getRepoRootPath(project, file);
            if (repoRootPath == null) {
                return new RepoInfo(vcsType, "", "", "");
            }

            // Determine file path, relative to repository root.
            relativePath = file.getPath().length() > repoRootPath.length()
                ? file.getPath().substring(repoRootPath.length() + 1) : "";
            if (vcsType == VCSType.PERFORCE && relativePath.indexOf('/') != -1) {
                relativePath = relativePath.substring(relativePath.indexOf("/") + 1);
            }

            // If the current branch doesn't exist on the remote, use the default branch.
            remoteBranchName = Optional.ofNullable(getRemoteBranchName(project, file))
                .orElse(ConfigUtil.getDefaultBranchName(project));

            remoteUrl = getRemoteRepoUrl(project, file);
            remoteUrl = doReplacements(project, remoteUrl);
        } catch (Exception err) {
            String message;
            if (err instanceof PerforceAuthenticationException) {
                message = "Perforce authentication error: " + err.getMessage();
            } else {
                message = "Error determining repository info: " + err.getMessage();
            }
            ErrorNotification.show(project, message);
            Logger.getInstance(RepoUtil.class).warn(message);
            err.printStackTrace();
        }
        return new RepoInfo(vcsType, remoteUrl, remoteBranchName, relativePath);
    }

    private static String doReplacements(@NotNull Project project, @NotNull String remoteUrl) {
        String remoteUrlWithReplacements = remoteUrl;
        String r = ConfigUtil.getRemoteUrlReplacements(project);
        String[] replacements = r.trim().split("\\s*,\\s*");
        if (replacements.length % 2 == 0) {
            for (int i = 0; i < replacements.length; i += 2) {
                remoteUrlWithReplacements = remoteUrlWithReplacements.replace(replacements[i], replacements[i + 1]);
            }
        }
        return remoteUrlWithReplacements;
    }

    @NotNull
    // Returned format: git@github.com:sourcegraph/sourcegraph.git
    public static String getRemoteRepoUrl(@NotNull Project project, @NotNull VirtualFile file) throws Exception {
        Repository repository = VcsRepositoryManager.getInstance(project).getRepositoryForFile(file);
        VCSType vcsType = getVcsType(project, file);

        if (vcsType == VCSType.GIT && repository != null) {
            return GitUtil.getRemoteRepoUrl((GitRepository) repository, project);
        }

        if (vcsType == VCSType.PERFORCE) {
            return PerforceUtil.getRemoteRepoUrl(project, file);
        }

        if (repository == null) {
            throw new Exception("Could not find repository for file " + file.getPath());
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
     * @return Like "main"
     */
    @Nullable
    private static String getRemoteBranchName(@NotNull Project project, @NotNull VirtualFile file) {
        Repository repository = VcsRepositoryManager.getInstance(project).getRepositoryForFile(file);
        if (repository == null) {
            return null;
        }

        if (repository instanceof GitRepository) {
            return GitUtil.getRemoteBranchName((GitRepository) repository);
        }

        // Unknown VCS.
        return null;
    }

    public static VCSType getVcsType(@NotNull Project project, @NotNull VirtualFile file) {
        Repository repository = VcsRepositoryManager.getInstance(project).getRepositoryForFile(file);

        try {
            Class.forName("git4idea.repo.GitRepository", false, RepoUtil.class.getClassLoader());
            if (repository instanceof GitRepository) {
                return VCSType.GIT;
            }
        } catch (ClassNotFoundException e) {
            // Git plugin is not installed.
        }

        try {
            Class.forName("org.jetbrains.idea.perforce.perforce.PerforceSettings", false, RepoUtil.class.getClassLoader());
            if (PerforceSettings.getSettings(project).getConnectionForFile(new File(file.getPath())) != null) {
                return VCSType.PERFORCE;
            }
        } catch (ClassNotFoundException e) {
            // Perforce plugin is not installed.
        }

        return VCSType.UNKNOWN;
    }

}
