package com.sourcegraph.website;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.application.ApplicationInfo;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.find.PreviewContent;
import com.sourcegraph.git.GitUtil;
import com.sourcegraph.git.RepoInfo;

import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.Objects;

public abstract class FileAction extends DumbAwareAction {

    abstract void handleFileUri(String uri);

    @Override
    public void actionPerformed(AnActionEvent e) {
        // Get project, editor, document, file, and position information.
        final Project project = e.getProject();
        if (project == null) {
            return;
        }
        Editor editor = FileEditorManager.getInstance(project).getSelectedTextEditor();
        if (editor == null) {
            return;
        }
        Document currentDoc = editor.getDocument();
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(currentDoc);
        if (currentFile == null) {
            return;
        }
        SelectionModel sel = editor.getSelectionModel();

        // Get repo information.
        RepoInfo repoInfo = GitUtil.getRepoInfo(currentFile.getPath(), project);
        if (Objects.equals(repoInfo.remoteUrl, "")) {
            return;
        }

        // Build the URL that we will open.
        String productName = ApplicationInfo.getInstance().getVersionName();
        String productVersion = ApplicationInfo.getInstance().getFullVersion();
        String uri;

        VisualPosition selectionStartPosition = sel.getSelectionStartPosition();
        VisualPosition selectionEndPosition = sel.getSelectionEndPosition();
        LogicalPosition start = selectionStartPosition != null ? editor.visualToLogicalPosition(selectionStartPosition) : null;
        LogicalPosition end = selectionEndPosition != null ? editor.visualToLogicalPosition(selectionEndPosition) : null;
        uri = ConfigUtil.getSourcegraphUrl(project) + "-/editor"
            + "?remote_url=" + URLEncoder.encode(repoInfo.remoteUrl, StandardCharsets.UTF_8)
            + "&branch=" + URLEncoder.encode(repoInfo.branchName, StandardCharsets.UTF_8)
            + "&file=" + URLEncoder.encode(repoInfo.relativePath, StandardCharsets.UTF_8)
            + "&editor=" + URLEncoder.encode("JetBrains", StandardCharsets.UTF_8)
            + "&version=v" + URLEncoder.encode(ConfigUtil.getPluginVersion(), StandardCharsets.UTF_8)
            + (start != null ? ("&start_row=" + URLEncoder.encode(Integer.toString(start.line), StandardCharsets.UTF_8)
            + "&start_col=" + URLEncoder.encode(Integer.toString(start.column), StandardCharsets.UTF_8)) : "")
            + (end != null ? ("&end_row=" + URLEncoder.encode(Integer.toString(end.line), StandardCharsets.UTF_8)
            + "&end_col=" + URLEncoder.encode(Integer.toString(end.column), StandardCharsets.UTF_8)) : "")
            + "&utm_product_name=" + URLEncoder.encode(productName, StandardCharsets.UTF_8)
            + "&utm_product_version=" + URLEncoder.encode(productVersion, StandardCharsets.UTF_8);

        handleFileUri(uri);
    }

    public void actionPerformedFromPreviewContent(Project project, PreviewContent previewContent, LogicalPosition start, LogicalPosition end) {
        String productName = ApplicationInfo.getInstance().getVersionName();
        String productVersion = ApplicationInfo.getInstance().getFullVersion();

        if (previewContent.getRepoUrl().isEmpty()) {
            return;
        }

        if (previewContent.getCommit() == null || previewContent.getCommit().isEmpty()) {
            return;
        }

        if (previewContent.getPath() == null || previewContent.getPath().isEmpty()) {
            return;
        }

        // Converting the information from the PreviewContent into a Sourcegraph blob URL
        String uri = ConfigUtil.getSourcegraphUrl(project)
            + previewContent.getRepoUrl()
            + "@"
            + previewContent.getCommit()
            + "/-/blob/"
            + previewContent.getPath()
            + "?"
            + (start != null ? ("L" + URLEncoder.encode(Integer.toString(start.line + 1), StandardCharsets.UTF_8)
            + ":" + URLEncoder.encode(Integer.toString(start.column + 1), StandardCharsets.UTF_8)) : "")
            + (end != null ? ("-" + URLEncoder.encode(Integer.toString(end.line + 1), StandardCharsets.UTF_8)
            + ":" + URLEncoder.encode(Integer.toString(end.column + 1), StandardCharsets.UTF_8)) : "")
            + "&editor=" + URLEncoder.encode("JetBrains", StandardCharsets.UTF_8)
            + "&version=v" + URLEncoder.encode(ConfigUtil.getPluginVersion(), StandardCharsets.UTF_8)
            + "&utm_product_name=" + URLEncoder.encode(productName, StandardCharsets.UTF_8)
            + "&utm_product_version=" + URLEncoder.encode(productVersion, StandardCharsets.UTF_8);

        handleFileUri(uri);
    }
}
