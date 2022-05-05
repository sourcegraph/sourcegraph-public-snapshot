package com.sourcegraph.find;

import com.intellij.codeInsight.highlighting.HighlightManager;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.editor.colors.EditorColors;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.Splitter;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.testFramework.LightVirtualFile;
import com.intellij.ui.OnePixelSplitter;
import com.intellij.ui.PopupBorder;
import com.intellij.ui.components.JBPanel;
import com.intellij.ui.components.JBPanelWithEmptyText;
import com.intellij.ui.jcef.JBCefApp;
import com.intellij.util.ui.JBUI;
import com.sourcegraph.browser.SourcegraphJBCefBrowser;
import org.jetbrains.annotations.NotNull;

import javax.annotation.Nullable;
import java.awt.*;

/**
 * Inspired by <a href="https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java">FindPopupPanel.java</a>
 */
public class FindPopupPanel extends JBPanel<FindPopupPanel> implements Disposable {
    private final SourcegraphJBCefBrowser browser;

    public FindPopupPanel(Project project) {
        super(new BorderLayout());

        setPreferredSize(JBUI.size(1200, 800));
        setBorder(PopupBorder.Factory.create(true, true));
        setFocusCycleRoot(true);

        // Create splitter
        Splitter splitter = new OnePixelSplitter(true, 0.5f, 0.1f, 0.9f);
        add(splitter, BorderLayout.CENTER);

        JBPanel<JBPanelWithEmptyText> jcefPanel = new JBPanelWithEmptyText(new BorderLayout()).withEmptyText("Unfortunately, the browser is not available on your system. Try running the IDE with the default OpenJDK.");
        browser = JBCefApp.isSupported() ? new SourcegraphJBCefBrowser() : null;
        if (browser != null) {
            jcefPanel.add(browser.getComponent(), BorderLayout.CENTER);
        }

        JBPanel<JBPanelWithEmptyText> previewPanel = createPreviewPanel(project);

        splitter.setFirstComponent(jcefPanel);
        splitter.setSecondComponent(previewPanel);
    }

    @Nullable
    public SourcegraphJBCefBrowser getBrowser() {
        return browser;
    }

    @NotNull
    private JBPanel<JBPanelWithEmptyText> createPreviewPanel(Project project) {
        EditorFactory editorFactory = EditorFactory.getInstance();

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

        HighlightManager highlightManager = HighlightManager.getInstance(project);
        highlightManager.addOccurrenceHighlight(editor, 23, 41, EditorColors.SEARCH_RESULT_ATTRIBUTES, 0, null);

        JBPanel<JBPanelWithEmptyText> editorPanel = new JBPanelWithEmptyText(new BorderLayout()).withEmptyText("Type search query to find on Sourcegraph");
        editorPanel.add(editor.getComponent(), BorderLayout.CENTER);
        editorPanel.invalidate();
        editorPanel.validate();

        return editorPanel;
    }

    @Override
    public void dispose() {
        if (browser != null) {
            browser.dispose();
        }
    }
}
