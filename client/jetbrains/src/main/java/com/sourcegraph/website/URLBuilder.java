package com.sourcegraph.website;

import com.intellij.openapi.editor.LogicalPosition;
import com.intellij.openapi.project.Project;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.net.URI;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;

import static com.sourcegraph.common.RegexEscaper.escapeRegexChars;

public class URLBuilder {
    @NotNull
    public static String buildEditorFileUrl(@NotNull Project project, @NotNull String remoteUrl, @NotNull String branchName, @NotNull String relativePath, @Nullable LogicalPosition start, @Nullable LogicalPosition end) {
        return ConfigUtil.getSourcegraphUrl(project) + "-/editor"
            + "?remote_url=" + URLEncoder.encode(remoteUrl, StandardCharsets.UTF_8)
            + "&branch=" + URLEncoder.encode(branchName, StandardCharsets.UTF_8)
            + "&file=" + URLEncoder.encode(relativePath, StandardCharsets.UTF_8)
            + (start != null ? ("&start_row=" + URLEncoder.encode(Integer.toString(start.line), StandardCharsets.UTF_8)
            + "&start_col=" + URLEncoder.encode(Integer.toString(start.column), StandardCharsets.UTF_8)) : "")
            + (end != null ? ("&end_row=" + URLEncoder.encode(Integer.toString(end.line), StandardCharsets.UTF_8)
            + "&end_col=" + URLEncoder.encode(Integer.toString(end.column), StandardCharsets.UTF_8)) : "")
            + "&" + buildVersionParams();
    }

    @NotNull
    public static String buildEditorSearchUrl(@NotNull Project project, @NotNull String search, @Nullable String remoteUrl, @Nullable String remoteBranchName) {
        String url = ConfigUtil.getSourcegraphUrl(project) + "-/editor"
            + "?" + buildVersionParams()
            + "&search=" + URLEncoder.encode(search, StandardCharsets.UTF_8);

        if (remoteUrl != null) {
            url += "&search_remote_url=" + URLEncoder.encode(remoteUrl, StandardCharsets.UTF_8);
            if (remoteBranchName != null) {
                url += "&search_branch=" + URLEncoder.encode(remoteBranchName, StandardCharsets.UTF_8);
            }
        }

        return url;
    }

    @NotNull
    public static String buildDirectSearchUrl(@NotNull Project project, @NotNull String search, @Nullable String codeHost, @Nullable String repoName) {
        String repoFilter = (codeHost != null && repoName != null) ? "repo:^" + escapeRegexChars(codeHost + "/" + repoName) + "$" : null;
        return ConfigUtil.getSourcegraphUrl(project) + "/search"
            + "?patternType=literal"
            + "&q=" + URLEncoder.encode((repoFilter != null ? repoFilter + " " : "") + search, StandardCharsets.UTF_8);
    }

    @NotNull
    public static String buildCommitUrl(@NotNull String sourcegraphBase, @NotNull String revisionNumber, @NotNull String remoteUrl,
                                        @NotNull String productName, @NotNull String productVersion) {
        if (sourcegraphBase.equals("")) {
            throw new RuntimeException("Missing sourcegraph URI for commit uri.");
        } else if (revisionNumber.equals("")) {
            throw new RuntimeException("Missing revision number for commit uri.");
        } else if (remoteUrl.equals("")) {
            throw new RuntimeException("Missing remote URL for commit uri.");
        }

        // this is pretty hacky but to try to build the repo string we will just try to naively parse the git remote uri. Worst case scenario this 404s
        if (remoteUrl.startsWith("git")) {
            remoteUrl = remoteUrl.replace(".git", "").replaceFirst(":", "/").replace("git@", "https://");
        }

        URI remote = URI.create(remoteUrl);

        return sourcegraphBase +
            String.format("/%s%s", remote.getHost(), remote.getPath()) +
            String.format("/-/commit/%s", revisionNumber) +
            String.format("?editor=%s", URLEncoder.encode("JetBrains", StandardCharsets.UTF_8)) +
            String.format("&version=v%s", URLEncoder.encode(ConfigUtil.getPluginVersion(), StandardCharsets.UTF_8)) +
            String.format("&utm_product_name=%s", URLEncoder.encode(productName, StandardCharsets.UTF_8)) +
            String.format("&utm_product_version=%s", URLEncoder.encode(productVersion, StandardCharsets.UTF_8));
    }

    @NotNull
    // repoUrl should be like "github.com/sourcegraph/sourcegraph"
    public static String buildSourcegraphBlobUrl(@NotNull Project project,
                                                 @NotNull String repoUrl,
                                                 @Nullable String commit,
                                                 @NotNull String path,
                                                 @Nullable LogicalPosition start,
                                                 @Nullable LogicalPosition end) {
        return ConfigUtil.getSourcegraphUrl(project)
            + repoUrl + (commit != null ? "@" + commit : "") + "/-/blob/" + path
            + "?"
            + (start != null ? ("L" + URLEncoder.encode(Integer.toString(start.line + 1), StandardCharsets.UTF_8)
            + ":" + URLEncoder.encode(Integer.toString(start.column + 1), StandardCharsets.UTF_8)) : "")
            + (end != null ? ("-" + URLEncoder.encode(Integer.toString(end.line + 1), StandardCharsets.UTF_8)
            + ":" + URLEncoder.encode(Integer.toString(end.column + 1), StandardCharsets.UTF_8)) : "");
    }

    @NotNull
    private static String buildVersionParams() {
        return "editor=" + URLEncoder.encode("JetBrains", StandardCharsets.UTF_8)
            + "&version=v" + URLEncoder.encode(ConfigUtil.getPluginVersion(), StandardCharsets.UTF_8);
    }
}
