package com.sourcegraph.find;

import com.intellij.codeInsight.highlighting.HighlightManager;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.actionSystem.ActionGroup;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.DefaultActionGroup;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.editor.colors.EditorColors;
import com.intellij.openapi.editor.impl.ContextMenuPopupHandler;
import com.intellij.openapi.editor.impl.EditorImpl;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.ui.components.JBPanelWithEmptyText;
import com.sourcegraph.Icons;
import com.sourcegraph.website.Copy;
import com.sourcegraph.website.FileAction;
import com.sourcegraph.website.OpenFile;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import javax.swing.*;
import java.awt.*;

public class PreviewPanel extends JBPanelWithEmptyText implements Disposable {
    private final String NO_PREVIEW_AVAILABLE_TEXT = "No preview available";
    @SuppressWarnings("FieldCanBeLocal") // It's nicer to have these here at the top
    private final String LOADING_TEXT = "Loading...";

    private final Project project;
    private JComponent editorComponent;
    private PreviewContent previewContent;
    private Editor editor;

    public PreviewPanel(Project project) {
        super(new BorderLayout());

        this.project = project;
        this.getEmptyText().setText(NO_PREVIEW_AVAILABLE_TEXT);
    }

    @Nullable
    public PreviewContent getPreviewContent() {
        return previewContent;
    }

    public void setContent(@Nullable PreviewContent previewContent) {
        String fileContent = previewContent != null ? previewContent.getContent() : null;
        if (previewContent == null || fileContent == null) {
            setState(State.NO_PREVIEW_AVAILABLE);
            return;
        }

        if (editorComponent != null && previewContent.equals(this.previewContent)) {
            setState(State.PREVIEW_AVAILABLE);
            return;
        }

        this.previewContent = previewContent;

        if (editorComponent != null) {
            remove(editorComponent);
        }
        if (editor != null) {
            EditorFactory.getInstance().releaseEditor(editor);
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

        ((EditorImpl) editor).installPopupHandler(new ContextMenuPopupHandler.Simple(this.createActionGroup()));

        setState(State.PREVIEW_AVAILABLE);

        editorComponent = editor.getComponent();
        add(editorComponent, BorderLayout.CENTER);
        validate();

        addAndScrollToHighlights(editor, previewContent.getAbsoluteOffsetAndLengths());
    }

    public void setState(@NotNull State state) {
        if (editorComponent != null) {
            editorComponent.setVisible(state == State.PREVIEW_AVAILABLE);
        }
        if (state == State.LOADING) {
            getEmptyText().setText(LOADING_TEXT);
        } else if (state == State.NO_PREVIEW_AVAILABLE) {
            getEmptyText().setText(NO_PREVIEW_AVAILABLE_TEXT);
        }
    }

    private void addAndScrollToHighlights(@NotNull Editor editor, @NotNull int[][] absoluteOffsetAndLengths) {
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

    @Override
    public void dispose() {
        if (editor != null) {
            EditorFactory.getInstance().releaseEditor(editor);
        }
    }

    private ActionGroup createActionGroup() {
        DefaultActionGroup group = new DefaultActionGroup();
        group.add(new DumbAwareAction("Open File in Editor", "Open file in editor", Icons.Logo) {
            @Override
            public void actionPerformed(@NotNull AnActionEvent e) {
                try {
                    if (getPreviewContent() != null) {
                        getPreviewContent().openInEditorOrBrowser();
                    }
                } catch (Exception ex) {
                    Logger logger = Logger.getInstance(SelectionMetadataPanel.class);
                    logger.error("Error opening file in editor: " + ex.getMessage());
                }
            }
        });
        group.add(new SimpleEditorFileAction("Open on Sourcegraph", new OpenFile(), editor));
        group.add(new SimpleEditorFileAction("Copy Sourcegraph File Link", new Copy(), editor));
        return group;
    }

    public enum State {
        LOADING,
        PREVIEW_AVAILABLE,
        NO_PREVIEW_AVAILABLE,
    }

    class SimpleEditorFileAction extends DumbAwareAction {
        FileAction action;
        Editor editor;

        SimpleEditorFileAction(String text, FileAction action, Editor editor) {
            super(text, text, Icons.Logo);
            this.action = action;
            this.editor = editor;
        }

        @Override
        public void actionPerformed(@NotNull AnActionEvent e) {
            SelectionModel sel = editor.getSelectionModel();
            VisualPosition selectionStartPosition = sel.getSelectionStartPosition();
            VisualPosition selectionEndPosition = sel.getSelectionEndPosition();
            LogicalPosition start = selectionStartPosition != null ? editor.visualToLogicalPosition(selectionStartPosition) : null;
            LogicalPosition end = selectionEndPosition != null ? editor.visualToLogicalPosition(selectionEndPosition) : null;

            action.actionPerformedFromPreviewContent(project, getPreviewContent(), start, end);
        }
    }
}
