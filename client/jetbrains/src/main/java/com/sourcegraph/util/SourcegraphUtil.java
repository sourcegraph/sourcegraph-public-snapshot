package com.sourcegraph.util;

import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.sourcegraph.project.RepoInfo;

public class SourcegraphUtil {
    public static final String VERSION = "v1.2.2";

    // repoInfo returns the Sourcegraph repository URI, and the file path
    // relative to the repository root. If the repository URI cannot be
    // determined, a RepoInfo with empty strings is returned.
    public static RepoInfo getRepoInfo(String filePath, Project project) {
        String relativePath = "";
        String remoteUrl = "";
        String branchName = "";
        try {
            String defaultBranchNameSetting = ConfigUtil.getDefaultBranchNameSetting(project);
            String repoRootPath = GitUtil.getRepoRootPath(filePath);

            // Determine file path, relative to repository root.
            relativePath = filePath.substring(repoRootPath.length() + 1);

            // TODO: It’d make more sense to default to the current branch if it exists on the remote, and only fall back to the default branch if it doesn’t.
            branchName = defaultBranchNameSetting != null ? defaultBranchNameSetting : GitUtil.getCurrentBranchName(repoRootPath);
            // If there’s no default branch name setting and the current branch doesn’t exist on the remote, use the default branch.
            if (!GitUtil.doesRemoteBranchExist(branchName, repoRootPath) && defaultBranchNameSetting == null) {
                branchName = "master"; // TODO: Make this dynamic!
            }

            remoteUrl = GitUtil.getConfiguredRemoteUrl(repoRootPath);
            // replace remoteURL if config option is not null
            String r = ConfigUtil.getRemoteUrlReplacements(project);
            if (r != null) {
                String[] replacements = r.trim().split("\\s*,\\s*");
                // Check if the entered values are pairs
                for (int i = 0; i < replacements.length && replacements.length % 2 == 0; i += 2) {
                    remoteUrl = remoteUrl.replace(replacements[i], replacements[i + 1]);
                }
            }
        } catch (Exception err) {
            Logger.getInstance(SourcegraphUtil.class).info(err);
            err.printStackTrace();
        }
        return new RepoInfo(relativePath, remoteUrl, branchName);
    }

}
