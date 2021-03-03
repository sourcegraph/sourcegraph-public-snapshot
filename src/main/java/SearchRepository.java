import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.SelectionModel;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.application.ApplicationInfo;

import javax.annotation.Nullable;
import java.io.*;
import java.awt.Desktop;
import java.net.URI;
import java.net.URLEncoder;

public class SearchRepository extends SearchActionBase {
    @Override
    public void actionPerformed(AnActionEvent e) {
        super.actionPerformedMode(e, "search.repository");
    }
}