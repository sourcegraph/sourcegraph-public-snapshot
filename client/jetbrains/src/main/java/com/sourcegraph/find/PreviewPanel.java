package com.sourcegraph.find;

import com.intellij.codeInsight.daemon.DaemonCodeAnalyzer;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.editor.*;
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

    public PreviewPanel(Project project) {
        super(new BorderLayout());

        this.project = project;
        this.getEmptyText().setText("Type search query to find on Sourcegraph");
    }

    public void setContent(@NotNull String fileName, @NotNull String fileContent) {
        if (editorComponent == null || !this.fileName.equals(fileName) || !this.fileContent.equals(fileContent)) {
            this.fileName = fileName;
            this.fileContent = fileContent;

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
                invalidate();
                validate();
            });
        }

        //HighlightManager highlightManager = HighlightManager.getInstance(project);
        //highlightManager.addOccurrenceHighlight(editor, 23, 41, EditorColors.SEARCH_RESULT_ATTRIBUTES, 0, null);

        // invalidate(); // TODO: Is this needed? What does it do? Maybe use revalidate()? If needed then document this.
        // validate();
    }

    public void clearContent() {
        if (editorComponent != null) {
            ApplicationManager.getApplication().invokeLater(() -> remove(editorComponent));
        }
    }

    public void setContentAndOpenInEditor(String fileName, String content) {
        ApplicationManager.getApplication().invokeLater(() -> {
            setContent(fileName, content);

            // Open file in editor
            OpenFileDescriptor openFileDescriptor = new OpenFileDescriptor(project, virtualFile, 0);
            FileEditorManager.getInstance(project).openTextEditor(openFileDescriptor, true);

            // Suppress code issues
            PsiFile file = PsiManager.getInstance(project).findFile(virtualFile);
            if (file != null) {
                DaemonCodeAnalyzer.getInstance(project).setHighlightingEnabled(file, false);
            }
        });
    }
}
