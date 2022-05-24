package com.sourcegraph.find;

import com.intellij.codeInsight.daemon.DaemonCodeAnalyzer;
import com.intellij.codeInsight.highlighting.HighlightManager;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.editor.colors.EditorColors;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.fileEditor.OpenFileDescriptor;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.psi.PsiFile;
import com.intellij.psi.PsiManager;
import com.intellij.testFramework.LightVirtualFile;
import com.intellij.ui.components.JBPanelWithEmptyText;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;

public class PreviewPanel extends JBPanelWithEmptyText {
    private final Project project;
    private JComponent editorComponent;
    private VirtualFile virtualFile;
    private String fileName;
    private String fileContent;
    private int lineNumber;

    public PreviewPanel(Project project) {
        super(new BorderLayout());

        this.project = project;
        this.getEmptyText().setText("Type search query to find on Sourcegraph");
    }

    public void setContent(@NotNull PreviewContent previewContent, boolean openInEditor) {
        if (editorComponent != null &&
            fileName.equals(previewContent.getFileName()) &&
            fileContent.equals(previewContent.getContent()) &&
            lineNumber == previewContent.getLineNumber() &&
            !openInEditor) {
            return;
        }

        fileName = previewContent.getFileName();
        fileContent = previewContent.getContent();
        lineNumber = previewContent.getLineNumber();

        ApplicationManager.getApplication().invokeLater(() -> {
            if (editorComponent != null) {
                remove(editorComponent);
            }
            EditorFactory editorFactory = EditorFactory.getInstance();
            virtualFile = new LightVirtualFile(fileName, fileContent);
            Document document = editorFactory.createDocument(fileContent);
            document.setReadOnly(true);
            Editor editor = editorFactory.createEditor(document, project, virtualFile, true, EditorKind.MAIN_EDITOR);

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

            if (openInEditor) {
                // Open file in editor
                OpenFileDescriptor openFileDescriptor = new OpenFileDescriptor(project, virtualFile, 0);
                FileEditorManager.getInstance(project).openTextEditor(openFileDescriptor, true);

                // Suppress code issues
                PsiFile file = PsiManager.getInstance(project).findFile(virtualFile);
                if (file != null) {
                    DaemonCodeAnalyzer.getInstance(project).setHighlightingEnabled(file, false);
                }
            }
        });
    }

    private void addHighlights(Editor editor, @NotNull int[][] absoluteOffsetAndLengths) {
        HighlightManager highlightManager = HighlightManager.getInstance(project);
        for (int[] offsetAndLength : absoluteOffsetAndLengths) {
            highlightManager.addOccurrenceHighlight(editor, offsetAndLength[0], offsetAndLength[0] + offsetAndLength[1], EditorColors.SEARCH_RESULT_ATTRIBUTES, 0, null);
        }
    }

    public void clearContent() {
        if (editorComponent != null) {
            ApplicationManager.getApplication().invokeLater(() -> remove(editorComponent));
        }
    }
}
