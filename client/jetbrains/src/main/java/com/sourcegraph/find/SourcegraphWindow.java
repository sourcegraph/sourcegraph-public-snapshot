package com.sourcegraph.find;

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
import com.intellij.ui.jcef.JBCefApp;
import com.intellij.util.ui.JBUI;
import com.sourcegraph.browser.SourcegraphJBCefBrowser;

import javax.swing.*;
import java.awt.*;
import java.util.Objects;

public class SourcegraphWindow implements Disposable {
    private final Project project;
    private final JPanel mainPanel;
    private final SourcegraphJBCefBrowser sourcegraphJBCefBrowser;
    private JBPopup popup;

    public SourcegraphWindow(Project project, JCEFService service) {
        this.project = project;

        mainPanel = new JPanel(new BorderLayout());
        mainPanel.setPreferredSize(JBUI.size(1200, 800));
        mainPanel.setBorder(PopupBorder.Factory.create(true, true));
        mainPanel.setFocusCycleRoot(true);

        EditorFactory editorFactory = EditorFactory.getInstance();
        JBPanel<JBPanelWithEmptyText> editorPanel = new JBPanelWithEmptyText(new BorderLayout());

        JPanel jcefPanel = new JPanel(new BorderLayout());
        /* Make sure JCEF is supported */
        if (!JBCefApp.isSupported()) {
            JLabel warningLabel = new JLabel("Unfortunately, the browser is not available on your system. Try running the IDE with the default OpenJDK.");
            jcefPanel.add(warningLabel);
            sourcegraphJBCefBrowser = null;
            return;
        }
        sourcegraphJBCefBrowser = service.getJcefWindow();

        JPanel topPanel = new JPanel(new BorderLayout());

        /* Add browser to panels */
        jcefPanel.add(Objects.requireNonNull(sourcegraphJBCefBrowser.getComponent()), BorderLayout.CENTER);
        topPanel.add(jcefPanel);


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

        mainPanel.add(splitter, BorderLayout.CENTER);
    }

    synchronized public void showPopup() {
        if (popup == null || popup.isDisposed()) {
            popup = createPopup();
            popup.showCenteredInCurrentWindow(project);
        }

        // If the popup is already shown, hitting alt + a gain should behave the same as the native find in files
        // feature and focus the search field.
        sourcegraphJBCefBrowser.focus();
    }

    private JBPopup createPopup() {
        return JBPopupFactory.getInstance().createComponentPopupBuilder(mainPanel, mainPanel)
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
        if (popup != null) {
            popup.dispose();
        }
    }
}
