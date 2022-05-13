package com.sourcegraph.config;

import com.intellij.openapi.project.Project;

import java.util.Objects;

public class ConfigUtil {
    public static String getDefaultBranchName(Project project) {
        String defaultBranch = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getDefaultBranchName();
        if (defaultBranch == null || defaultBranch.length() == 0) {
            return UserLevelConfig.getDefaultBranchName();
        }
        return defaultBranch;
    }

    public static String getRemoteUrlReplacements(Project project) {
        String replacements = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getRemoteUrlReplacements();
        if (replacements == null || replacements.length() == 0) {
            return UserLevelConfig.getRemoteUrlReplacements();
        }
        return replacements;
    }

    public static String getSourcegraphUrl(Project project) {
        String url = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getSourcegraphUrl();
        if (url == null || url.length() == 0) {
            return UserLevelConfig.getSourcegraphUrl();
        }
        return url.endsWith("/") ? url : url + "/";
    }

    public static String getVersion() {
        return "v1.2.2";
    }
}
