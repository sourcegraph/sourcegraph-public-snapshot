package com.sourcegraph.website;

import com.intellij.openapi.editor.LogicalPosition;
import com.intellij.openapi.project.Project;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;

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
    public static String buildEditorSearchUrl(@NotNull Project project, @NotNull String search, @Nullable String remoteUrl, @Nullable String branchName) {
        String url = ConfigUtil.getSourcegraphUrl(project) + "-/editor"
            + "?" + buildVersionParams()
            + "&search=" + URLEncoder.encode(search, StandardCharsets.UTF_8);

        if (remoteUrl != null) {
            url += "&search_remote_url=" + URLEncoder.encode(remoteUrl, StandardCharsets.UTF_8);
            if (branchName != null) {
                url += "&search_branch=" + URLEncoder.encode(branchName, StandardCharsets.UTF_8);
            }
        }

        return url;
    }

    @NotNull
    public static String buildSourcegraphBlobUrl(@NotNull Project project, @NotNull String repoUrl, @NotNull String commit, @NotNull String path, @Nullable LogicalPosition start, @Nullable LogicalPosition end) {
        return ConfigUtil.getSourcegraphUrl(project)
            + repoUrl + "@" + commit + "/-/blob/" + path
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
