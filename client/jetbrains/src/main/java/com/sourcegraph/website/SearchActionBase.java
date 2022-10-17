package com.sourcegraph.website;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.editor.Caret;
import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.SelectionModel;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.TextRange;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.common.BrowserOpener;
import com.sourcegraph.find.SourcegraphVirtualFile;
import com.sourcegraph.git.GitUtil;
import com.sourcegraph.git.RepoInfo;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public abstract class SearchActionBase extends DumbAwareAction {
    public void actionPerformedMode(@NotNull AnActionEvent event, @NotNull Scope scope) {
        final Project project = event.getProject();

        String selectedText = getSelectedText(project);

        if (selectedText == null || selectedText.length() == 0) {
            return;
        }
        //noinspection ConstantConditions selectedText != null, so the editor can't be null.
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(FileEditorManager.getInstance(project).getSelectedTextEditor().getDocument());
        assert currentFile != null; // selectedText != null, so this can't be null.

        String url;
        if (currentFile instanceof SourcegraphVirtualFile) {
            SourcegraphVirtualFile sourcegraphFile = (SourcegraphVirtualFile) currentFile;
            String repoUrl = (scope == Scope.REPOSITORY) ? sourcegraphFile.getRepoUrl() : null;
            url = URLBuilder.buildEditorSearchUrl(project, selectedText, repoUrl, null);
        } else {
            RepoInfo repoInfo = GitUtil.getRepoInfo(currentFile.getPath(), project);
            String remoteUrl = (scope == Scope.REPOSITORY) ? repoInfo.remoteUrl : null;
            String branchName = (scope == Scope.REPOSITORY) ? repoInfo.branchName : null;
            url = URLBuilder.buildEditorSearchUrl(project, selectedText, remoteUrl, branchName);
        }

        BrowserOpener.openInBrowser(project, url);
    }

    enum Scope {
        REPOSITORY,
        ANYWHERE
    }

    @Override
    public void update(@NotNull AnActionEvent e) {
        final Project project = e.getProject();
        if (project == null) {
            return;
        }
        String selectedText = getSelectedText(project);
        e.getPresentation().setEnabled(selectedText != null && selectedText.length() > 0);
    }

    @Nullable
    private String getSelectedText(@Nullable Project project) {
        if (project == null) {
            return null;
        }

        Editor editor = FileEditorManager.getInstance(project).getSelectedTextEditor();
        if (editor == null) {
            return null;
        }

        Document currentDocument = editor.getDocument();
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(currentDocument);
        if (currentFile == null) {
            return null;
        }

        SelectionModel selectionModel = editor.getSelectionModel();
        String selectedText = selectionModel.getSelectedText();
        if (selectedText != null && !selectedText.equals("")) {
            return selectedText;
        }

        // Get whole current line, trimmed
        Caret caret = editor.getCaretModel().getCurrentCaret();
        selectedText = currentDocument.getText(new TextRange(caret.getVisualLineStart(), caret.getVisualLineEnd())).trim();

        return !selectedText.equals("") ? selectedText : null;
    }
}
