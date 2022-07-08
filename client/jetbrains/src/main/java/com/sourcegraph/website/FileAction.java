package com.sourcegraph.website;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.browser.URLBuilder;
import com.sourcegraph.find.PreviewContent;
import com.sourcegraph.git.GitUtil;
import com.sourcegraph.git.RepoInfo;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.util.Objects;

public abstract class FileAction extends DumbAwareAction {
    abstract void handleFileUri(@NotNull String uri);

    @Override
    public void actionPerformed(@NotNull AnActionEvent e) {
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

        // Get repo information.
        RepoInfo repoInfo = GitUtil.getRepoInfo(currentFile.getPath(), project);
        if (Objects.equals(repoInfo.remoteUrl, "")) {
            return;
        }

        String uri = URLBuilder.buildEditorFileUrl(project, repoInfo.remoteUrl, repoInfo.branchName, repoInfo.relativePath, getSelectionStartPosition(editor), getSelectionEndPosition(editor));

        handleFileUri(uri);
    }

    @Nullable
    private LogicalPosition getSelectionStartPosition(@NotNull Editor editor) {
        SelectionModel sel = editor.getSelectionModel();
        VisualPosition position = sel.getSelectionStartPosition();
        return position != null ? editor.visualToLogicalPosition(position) : null;
    }

    @Nullable
    private LogicalPosition getSelectionEndPosition(@NotNull Editor editor) {
        SelectionModel sel = editor.getSelectionModel();
        VisualPosition position = sel.getSelectionEndPosition();
        return position != null ? editor.visualToLogicalPosition(position) : null;
    }

    public void actionPerformedFromPreviewContent(@NotNull Project project, @Nullable PreviewContent previewContent, @Nullable LogicalPosition start, @Nullable LogicalPosition end) {
        if (previewContent == null) {
            return;
        }

        if (previewContent.getRepoUrl().isEmpty()) {
            return;
        }

        if (previewContent.getCommit() == null || previewContent.getCommit().isEmpty()) {
            return;
        }

        if (previewContent.getPath() == null || previewContent.getPath().isEmpty()) {
            return;
        }

        String uri = URLBuilder.buildSourcegraphBlobUrl(project, previewContent.getRepoUrl(), previewContent.getCommit(), previewContent.getPath(), start, end);

        handleFileUri(uri);
    }
}
