package com.sourcegraph.find;

import com.intellij.codeInsight.highlighting.HighlightManager;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.editor.colors.EditorColors;
import com.intellij.openapi.project.Project;
import com.intellij.ui.components.JBPanelWithEmptyText;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import javax.swing.*;
import java.awt.*;

public class PreviewPanel extends JBPanelWithEmptyText implements Disposable {
    private final Project project;
    private JComponent editorComponent;
    private PreviewContent previewContent;
    private Editor editor;

    public PreviewPanel(Project project) {
        super(new BorderLayout());

        this.project = project;
        this.getEmptyText().setText("(No preview available)");
    }

    @Nullable
    public PreviewContent getPreviewContent() {
        return previewContent;
    }

    public void setContent(@NotNull PreviewContent previewContent) {
        if (editorComponent != null && previewContent.equals(this.previewContent)) {
            return;
        }

        this.previewContent = previewContent;
        String fileContent = previewContent.getContent();

        /* If no content, just show "No preview available" */
        if (fileContent == null) {
            clearContent();
            return;
        }

        if (previewContent.getVirtualFile() == null) {
            clearContent();
            return;
        }

        ApplicationManager.getApplication().invokeLater(() -> {
            if (editorComponent != null) {
                remove(editorComponent);
            }
            EditorFactory editorFactory = EditorFactory.getInstance();
            Document document = editorFactory.createDocument(fileContent);
            document.setReadOnly(true);

            editor = editorFactory.createEditor(document, project, previewContent.getVirtualFile(), true, EditorKind.MAIN_EDITOR);

            EditorSettings settings = editor.getSettings();
            settings.setLineMarkerAreaShown(true);
            settings.setFoldingOutlineShown(false);
            settings.setAdditionalColumnsCount(0);
            settings.setAdditionalLinesCount(0);
            settings.setAnimatedScrolling(false);
            settings.setAutoCodeFoldingEnabled(false);

            editorComponent = editor.getComponent();
            add(editorComponent, BorderLayout.CENTER);

            revalidate();
            repaint();

            addAndScrollToHighlights(editor, previewContent.getAbsoluteOffsetAndLengths());
        });
    }

    private void addAndScrollToHighlights(Editor editor, int[][] absoluteOffsetAndLengths) {
        int firstOffset = -1;
        HighlightManager highlightManager = HighlightManager.getInstance(project);
        for (int[] offsetAndLength : absoluteOffsetAndLengths) {
            if (firstOffset == -1) {
                firstOffset = offsetAndLength[0] + offsetAndLength[1];
            }

            highlightManager.addOccurrenceHighlight(editor, offsetAndLength[0], offsetAndLength[0] + offsetAndLength[1], EditorColors.TEXT_SEARCH_RESULT_ATTRIBUTES, 0, null);
        }

        if (firstOffset != -1) {
            editor.getScrollingModel().scrollTo(editor.offsetToLogicalPosition(firstOffset), ScrollType.CENTER);
        }
    }

    public void clearContent() {
        if (editorComponent != null) {
            ApplicationManager.getApplication().invokeLater(() -> {
                remove(editorComponent);
                editorComponent = null;
                revalidate();
                repaint();
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
