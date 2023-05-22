package com.sourcegraph.cody.editor;

import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.TextRange;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.cody.TruncationUtils;
import org.jetbrains.annotations.NotNull;

public class EditorContextGetter {
    @NotNull
    public static EditorContext getEditorContext(@NotNull Project project) {
        Editor editor = FileEditorManager.getInstance(project).getSelectedTextEditor();
        if (editor == null) {
            return new EditorContext();
        }
        @NotNull Document currentDocument = editor.getDocument();
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(currentDocument);
        if (currentFile == null) {
            return new EditorContext();
        }

        // Get preceding text
        int startLine = Math.max(0, currentDocument.getLineNumber(editor.getSelectionModel().getSelectionStart()) - TruncationUtils.SURROUNDING_LINES);
        int precedingTextStartOffset = currentDocument.getLineStartOffset(startLine);
        int precedingTextEndOffset = editor.getSelectionModel().getSelectionStart();
        String precedingText = currentDocument.getText(new TextRange(precedingTextStartOffset, precedingTextEndOffset));

        // Get selection
        String selection = editor.getSelectionModel().getSelectedText();

        // Get following text
        int endLine = Math.min(currentDocument.getLineCount() - 1, currentDocument.getLineNumber(editor.getSelectionModel().getSelectionEnd()) + TruncationUtils.SURROUNDING_LINES);
        int followingTextStartOffset = editor.getSelectionModel().getSelectionEnd();
        int followingTextEndOffset = currentDocument.getLineEndOffset(endLine);
        String followingText = currentDocument.getText(new TextRange(followingTextStartOffset, followingTextEndOffset));

        return new EditorContext(
            currentFile.getName(),
            currentDocument.getText(),
            precedingText,
            selection == null || selection.isEmpty() ? null : selection,
            followingText);
    }
}
