package com.sourcegraph.cody.editor;

import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class EditorUtil {
  @Nullable
  public static VirtualFile getCurrentFile(@NotNull Project project) {
    VirtualFile currentFile = null;
    Editor editor = FileEditorManager.getInstance(project).getSelectedTextEditor();
    if (editor != null) {
      Document currentDocument = editor.getDocument();
      currentFile = FileDocumentManager.getInstance().getFile(currentDocument);
    }
    return currentFile;
  }
}
