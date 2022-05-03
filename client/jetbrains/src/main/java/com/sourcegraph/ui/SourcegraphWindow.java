package com.sourcegraph.ui;

import com.intellij.codeInsight.highlighting.HighlightManager;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.editor.colors.EditorColors;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.Splitter;
import com.intellij.openapi.ui.popup.JBPopup;
import com.intellij.openapi.ui.popup.JBPopupFactory;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.testFramework.LightVirtualFile;
import com.intellij.ui.OnePixelSplitter;
import com.intellij.ui.PopupBorder;
import com.intellij.ui.components.JBPanel;
import com.intellij.ui.components.JBPanelWithEmptyText;
import com.intellij.util.ui.JBUI;
import com.sourcegraph.service.JCEFService;

import javax.swing.*;
import java.awt.*;

public class SourcegraphWindow implements Disposable {
    private final Project project;
    private final JPanel panel;
    private final JCEFWindow jcefWindow;
    private JBPopup popup;

    public SourcegraphWindow(Project project) {
        this.project = project;

        panel = new JPanel(new BorderLayout());
        panel.setPreferredSize(JBUI.size(1200, 800));
        panel.setBorder(PopupBorder.Factory.create(true, true));
        panel.setFocusCycleRoot(true);

        EditorFactory editorFactory = EditorFactory.getInstance();
        JBPanel<JBPanelWithEmptyText> editorPanel = new JBPanelWithEmptyText(new BorderLayout());

        JCEFService service = project.getService(JCEFService.class);
        this.jcefWindow = service.getJcefWindow();

        JPanel topPanel = new JPanel(new BorderLayout());
        topPanel.add(this.jcefWindow.getContent());


        String contentTs = "let message: string = 'Hello, TypeScript!';\n" +
            "\n" +
            "let heading = document.createElement('h1');\n" +
            "heading.textContent = message;\n" +
            "\n" +
            "document.body.appendChild(heading);";
        VirtualFile virtualFile = new LightVirtualFile("helloWorld.ts", contentTs);
        Document document = editorFactory.createDocument(contentTs);

        Editor editor = editorFactory.createEditor(document, project, virtualFile, true, EditorKind.MAIN_EDITOR);
        EditorSettings settings = editor.getSettings();
        settings.setLineMarkerAreaShown(true);
        settings.setFoldingOutlineShown(false);
        settings.setAdditionalColumnsCount(0);
        settings.setAdditionalLinesCount(0);
        settings.setAnimatedScrolling(false);
        settings.setAutoCodeFoldingEnabled(false);
        editorPanel.add(editor.getComponent(), BorderLayout.CENTER);
        editorPanel.invalidate();
        editorPanel.validate();
        HighlightManager highlightManager = HighlightManager.getInstance(project);
        highlightManager.addOccurrenceHighlight(editor, 23, 41, EditorColors.SEARCH_RESULT_ATTRIBUTES, 0, null);

        Splitter splitter = new OnePixelSplitter(true, 0.5f, 0.1f, 0.9f);
        splitter.setFirstComponent(topPanel);
        splitter.setSecondComponent(editorPanel);

        panel.add(splitter, BorderLayout.CENTER);
    }

    synchronized public void showPopup() {
        if (this.popup == null || this.popup.isDisposed()) {
            this.popup = createPopup();
            this.popup.showCenteredInCurrentWindow(this.project);
        }

        // If the popup is already shown, hitting alt + a gain should behave the same as the native find in files
        // feature and focus the search field.
        this.jcefWindow.focus();
    }

    private JBPopup createPopup() {
        return JBPopupFactory.getInstance().createComponentPopupBuilder(panel, panel)
            .setTitle("Sourcegraph")
            .setCancelOnClickOutside(false)
            .setResizable(true)
            .setModalContext(false)
            .setRequestFocus(true)
            .setFocusable(true)
            .setMovable(true)
            .setBelongsToGlobalPopupStack(true)
            .setCancelOnOtherWindowOpen(true)
            .setCancelKeyEnabled(true)
            .setNormalWindowLevel(true)
            .createPopup();
    }

    @Override
    public void dispose() {
        if (this.popup != null) {
            this.popup.dispose();
        }
    }
}
