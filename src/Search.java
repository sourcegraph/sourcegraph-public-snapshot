import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.command.CommandProcessor;
import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.SelectionModel;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.editor.LogicalPosition;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.sun.org.apache.bcel.internal.generic.NEW;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.application.ApplicationInfo;

import java.io.*;
import java.awt.Desktop;
import java.net.URI;
import java.net.URLEncoder;

public class Search extends AnAction {
    @Override
    public void actionPerformed(AnActionEvent e) {
        Logger logger = Logger.getInstance(this.getClass());

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
        if (currentDoc == null) {
            return;
        }
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(currentDoc);
        if (currentFile == null) {
            return;
        }
        SelectionModel sel = editor.getSelectionModel();

        String q = sel.getSelectedText();
        if (q.equals("")) {
            return; // nothing to query
        }

        // Build the URL that we will open.
        String uri;
        String productName = ApplicationInfo.getInstance().getVersionName();
        String productVersion = ApplicationInfo.getInstance().getFullVersion();
        try {
            uri = Util.sourcegraphURL()+"-/editor"
                    + "?search=" + URLEncoder.encode(q, "UTF-8")
                    + "&editor=" + URLEncoder.encode("JetBrains", "UTF-8")
                    + "&version=" + URLEncoder.encode(Util.VERSION, "UTF-8")
                    + "&utm_product_name=" + URLEncoder.encode(productName, "UTF-8")
                    + "&utm_product_version=" + URLEncoder.encode(productVersion, "UTF-8");
        } catch (UnsupportedEncodingException err) {
            logger.debug("failed to build URL");
            err.printStackTrace();
            return;
        }

        // Open the URL in the browser.
        try {
            Desktop.getDesktop().browse(URI.create(uri));
        } catch (IOException err) {
            logger.debug("failed to open browser");
            err.printStackTrace();
        }
        return;
    }
}