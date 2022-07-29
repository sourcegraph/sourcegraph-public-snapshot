package com.sourcegraph.website;

import com.google.common.base.Strings;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.find.PreviewContent;
import com.sourcegraph.find.SourcegraphVirtualFile;
import com.sourcegraph.git.GitUtil;
import com.sourcegraph.git.RepoInfo;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public abstract class FileActionBase extends DumbAwareAction {
    abstract protected void handleFileUri(@NotNull Project project, @NotNull String uri);

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
        Document currentDocument = editor.getDocument();
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(currentDocument);
        if (currentFile == null) {
            return;
        }

        if (currentFile instanceof SourcegraphVirtualFile) {
            SourcegraphVirtualFile sourcegraphFile = (SourcegraphVirtualFile) currentFile;
            handleFileUri(project, URLBuilder.buildSourcegraphBlobUrl(project, sourcegraphFile.getRepoUrl(), sourcegraphFile.getCommit(), sourcegraphFile.getRelativePath(), getSelectionStartPosition(editor), getSelectionEndPosition(editor)));
        } else {
            RepoInfo repoInfo = GitUtil.getRepoInfo(currentFile.getPath(), project);
            if (repoInfo.remoteUrl.equals("")) {
                return;
            }

            handleFileUri(project, URLBuilder.buildEditorFileUrl(project, repoInfo.remoteUrl, repoInfo.branchName, repoInfo.relativePath, getSelectionStartPosition(editor), getSelectionEndPosition(editor)));
        }
    }

    public void actionPerformedFromPreviewContent(@NotNull Project project, @Nullable PreviewContent previewContent, @Nullable LogicalPosition start, @Nullable LogicalPosition end) {
        if (previewContent == null
            || previewContent.getRepoUrl().isEmpty()
            || Strings.isNullOrEmpty(previewContent.getCommit())
            || Strings.isNullOrEmpty(previewContent.getPath())) {
            return;
        }

        handleFileUri(project, URLBuilder.buildSourcegraphBlobUrl(project, previewContent.getRepoUrl(), previewContent.getCommit(), previewContent.getPath(), start, end));
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
}
