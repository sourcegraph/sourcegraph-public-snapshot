package com.sourcegraph.cody.editor;

import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import org.jetbrains.annotations.NotNull;

public class EditorContextGetter {
    @NotNull
    public static EditorContext getEditorContext(@NotNull Project project) {
        Editor editor = FileEditorManager.getInstance(project).getSelectedTextEditor();
        if (editor == null) {
            return new EditorContext();
        }
        Document currentDocument = editor.getDocument();
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(currentDocument);
        if (currentFile == null) {
            return new EditorContext();
        }
        String selection = editor.getSelectionModel().getSelectedText();
        return new EditorContext(
            currentFile.getName(),
            currentDocument.getText(),
            selection == null || selection.isEmpty() ? null : selection);
    }
}
