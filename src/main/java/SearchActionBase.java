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

public abstract class SearchActionBase extends AnAction {
    public void actionPerformedMode(AnActionEvent e, String mode) {
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

        // Get repo information.
        RepoInfo repoInfo = Util.repoInfo(currentFile.getPath());

        String q = sel.getSelectedText();
        if (q == null || q.equals("")) {
            return; // nothing to query
        }

        // Build the URL that we will open.
        String uri;
        String productName = ApplicationInfo.getInstance().getVersionName();
        String productVersion = ApplicationInfo.getInstance().getFullVersion();
        try {
            uri = Util.sourcegraphURL(project)+"-/editor"
                    + "?editor=" + URLEncoder.encode("JetBrains", "UTF-8")
                    + "&version=" + URLEncoder.encode(Util.VERSION, "UTF-8")
                    + "&utm_product_name=" + URLEncoder.encode(productName, "UTF-8")
                    + "&utm_product_version=" + URLEncoder.encode(productVersion, "UTF-8")
                    + "&search=" + URLEncoder.encode(q, "UTF-8");

            if (mode == "search.repository") {
                uri += "&search_remote_url=" + URLEncoder.encode(repoInfo.remoteURL, "UTF-8")
                        + "&search_branch=" + URLEncoder.encode(repoInfo.branch, "UTF-8");
            }

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

    @Override
    public void update(AnActionEvent e) {
        final Project project = e.getProject();
        if (project == null) {
            return;
        }
        String selectedText = getSelectedText(project);
        e.getPresentation().setEnabled(selectedText != null && selectedText.length() > 0);
    }

    @Nullable
    private String getSelectedText(Project project) {
        Editor editor = FileEditorManager.getInstance(project).getSelectedTextEditor();
        if (editor == null) {
            return null;
        }
        Document currentDoc = editor.getDocument();
        if (currentDoc == null) {
            return null;
        }
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(currentDoc);
        if (currentFile == null) {
            return null;
        }
        SelectionModel sel = editor.getSelectionModel();

        return sel.getSelectedText();
    }
}