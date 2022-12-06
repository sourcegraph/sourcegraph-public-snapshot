package com.sourcegraph.website;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.application.ApplicationManager;
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
import com.sourcegraph.vcs.RepoInfo;
import com.sourcegraph.vcs.RepoUtil;
import com.sourcegraph.vcs.VCSType;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.util.logging.Level;
import java.util.logging.Logger;

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

        if (currentFile instanceof SourcegraphVirtualFile) {
            String url;
            SourcegraphVirtualFile sourcegraphFile = (SourcegraphVirtualFile) currentFile;
            String repoUrl = (scope == Scope.REPOSITORY) ? sourcegraphFile.getRepoUrl() : null;
            url = URLBuilder.buildEditorSearchUrl(project, selectedText, repoUrl, null);
            BrowserOpener.openInBrowser(project, url);
        } else {
            // This cannot run on EDT (Event Dispatch Thread) because it may block for a long time.
            ApplicationManager.getApplication().executeOnPooledThread(
                () -> {
                    String url;
                    RepoInfo repoInfo = RepoUtil.getRepoInfo(project, currentFile);
                    String remoteUrl = (scope == Scope.REPOSITORY) ? repoInfo.remoteUrl : null;
                    String remoteBranchName = (scope == Scope.REPOSITORY) ? repoInfo.remoteBranchName : null;
                    if (repoInfo.vcsType == VCSType.PERFORCE) {
                        // Our "editor" backend doesn't support Perforce, but we have all the info we need, so we'll go to the final URL directly.
                        String codeHostUrl = (scope == Scope.REPOSITORY) ? repoInfo.getCodeHostUrl() : null;
                        String repoName = (scope == Scope.REPOSITORY) ? repoInfo.getRepoName() : null;
                        url = URLBuilder.buildDirectSearchUrl(project, selectedText, codeHostUrl, repoName);
                    } else {
                        url = URLBuilder.buildEditorSearchUrl(project, selectedText, remoteUrl, remoteBranchName);
                    }
                    BrowserOpener.openInBrowser(project, url);
                }
            );
        }
    }

    protected enum Scope {
        REPOSITORY,
        ANYWHERE
    }

    @Override
    public void update(@NotNull AnActionEvent event) {
        final Project project = event.getProject();
        if (project == null) {
            return;
        }
        // This must run on EDT (Event Dispatch Thread) because it interacts with the editor.
        ApplicationManager.getApplication().invokeLater(() -> {
            try {
                String selectedText = getSelectedText(project);
                event.getPresentation().setEnabled(selectedText != null && selectedText.length() > 0);
            } catch (Exception exception) {
                Logger logger = Logger.getLogger(SearchActionBase.class.getName());
                logger.log(Level.WARNING, "Problem while getting selected text", exception);
                event.getPresentation().setEnabled(false);
            }
        });
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
