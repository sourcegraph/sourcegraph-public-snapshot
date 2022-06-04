package com.sourcegraph.find;

import com.intellij.codeInsight.daemon.DaemonCodeAnalyzer;
import com.intellij.codeInsight.highlighting.HighlightManager;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.editor.colors.EditorColors;
import com.intellij.openapi.externalSystem.service.execution.NotSupportedException;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.fileEditor.OpenFileDescriptor;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.psi.PsiFile;
import com.intellij.psi.PsiManager;
import com.intellij.testFramework.LightVirtualFile;
import com.intellij.ui.components.JBPanelWithEmptyText;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import javax.swing.*;
import java.awt.*;
import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.Objects;

public class PreviewPanel extends JBPanelWithEmptyText implements Disposable {
    private final Project project;
    private JComponent editorComponent;

    private PreviewContent previewContent;
    private VirtualFile virtualFile;
    private Editor editor;

    public PreviewPanel(Project project) {
        super(new BorderLayout());

        this.project = project;
        this.getEmptyText().setText("(No preview available)");
    }

    public void setContent(@NotNull PreviewContent previewContent) {
        if (editorComponent != null && previewContent.equals(this.previewContent)) {
            return;
        }

        this.previewContent = previewContent;
        String fileContent = previewContent.getContent();

        /* If no content, just show “No preview available” */
        if (fileContent == null) {
            clearContent();
            return;
        }

        ApplicationManager.getApplication().invokeLater(() -> {
            if (editorComponent != null) {
                remove(editorComponent);
            }
            EditorFactory editorFactory = EditorFactory.getInstance();
            virtualFile = new LightVirtualFile(this.previewContent.getFileName(), fileContent);
            Document document = editorFactory.createDocument(fileContent);
            document.setReadOnly(true);
            editor = editorFactory.createEditor(document, project, virtualFile, true, EditorKind.MAIN_EDITOR);

            EditorSettings settings = editor.getSettings();
            settings.setLineMarkerAreaShown(true);
            settings.setFoldingOutlineShown(false);
            settings.setAdditionalColumnsCount(0);
            settings.setAdditionalLinesCount(0);
            settings.setAnimatedScrolling(false);
            settings.setAutoCodeFoldingEnabled(false);

            editorComponent = editor.getComponent();
            add(editorComponent, BorderLayout.CENTER);

            addHighlights(editor, previewContent.getAbsoluteOffsetAndLengths());

            invalidate(); // TODO: Is this needed? What does it do? Maybe use revalidate()? If needed then document
            validate();
        });
    }

    public void openInEditorOrBrowser() throws URISyntaxException, IOException, NotSupportedException {
        openInEditorOrBrowser(this.previewContent);
    }

    public void openInEditorOrBrowser(@Nullable PreviewContent previewContent) throws URISyntaxException, IOException, NotSupportedException {
        if (previewContent == null) {
            return;
        }

        if (previewContent.getFileName().length() == 0) {
            openInBrowser(previewContent);
        } else {
            openInEditor(previewContent);
        }
    }

    private void openInEditor(@NotNull PreviewContent previewContent) {
        // Open file in editor
        virtualFile = new LightVirtualFile(this.previewContent.getFileName(), Objects.requireNonNull(previewContent.getContent()));
        OpenFileDescriptor openFileDescriptor = new OpenFileDescriptor(project, virtualFile, 0);
        FileEditorManager.getInstance(project).openTextEditor(openFileDescriptor, true);

        // Suppress code issues
        PsiFile file = PsiManager.getInstance(project).findFile(virtualFile);
        if (file != null) {
            DaemonCodeAnalyzer.getInstance(project).setHighlightingEnabled(file, false);
        }
    }

    private void openInBrowser(@NotNull PreviewContent previewContent) throws URISyntaxException, IOException, NotSupportedException {
        // Source: https://stackoverflow.com/questions/5226212/how-to-open-the-default-webbrowser-using-java
        if (Desktop.isDesktopSupported() && Desktop.getDesktop().isSupported(Desktop.Action.BROWSE)) {
            String sourcegraphUrl = ConfigUtil.getSourcegraphUrl(this.project);
            Desktop.getDesktop().browse(new URI(sourcegraphUrl + "/" + previewContent.getRelativeUrl()));
        } else {
            throw new NotSupportedException("Can't open link. Desktop is not supported.");
        }
    }

    private void addHighlights(Editor editor, @NotNull int[][] absoluteOffsetAndLengths) {
        HighlightManager highlightManager = HighlightManager.getInstance(project);
        for (int[] offsetAndLength : absoluteOffsetAndLengths) {
            highlightManager.addOccurrenceHighlight(editor, offsetAndLength[0], offsetAndLength[0] + offsetAndLength[1], EditorColors.SEARCH_RESULT_ATTRIBUTES, 0, null);
        }
    }

    public void clearContent() {
        if (editorComponent != null) {
            ApplicationManager.getApplication().invokeLater(() -> {
                remove(editorComponent);
                editorComponent = null;
                virtualFile = null;
            });
        }
    }

    @Override
    public void dispose() {
        if (editor != null) {
            EditorFactory.getInstance().releaseEditor(editor);
        }
    }
}
