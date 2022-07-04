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
        SelectionModel sel = editor.getSelectionModel();

        // Get repo information.
        RepoInfo repoInfo = GitUtil.getRepoInfo(currentFile.getPath(), project);
        if (Objects.equals(repoInfo.remoteUrl, "")) {
            return;
        }

        VisualPosition selectionStartPosition = sel.getSelectionStartPosition();
        VisualPosition selectionEndPosition = sel.getSelectionEndPosition();
        LogicalPosition start = selectionStartPosition != null ? editor.visualToLogicalPosition(selectionStartPosition) : null;
        LogicalPosition end = selectionEndPosition != null ? editor.visualToLogicalPosition(selectionEndPosition) : null;

        String uri = URLBuilder.buildEditorFileUrl(project, repoInfo.remoteUrl, repoInfo.branchName, repoInfo.relativePath, start, end);

        handleFileUri(uri);
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
