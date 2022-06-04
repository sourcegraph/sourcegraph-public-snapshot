package com.sourcegraph.find;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.Splitter;
import com.intellij.ui.OnePixelSplitter;
import com.intellij.ui.PopupBorder;
import com.intellij.ui.components.JBPanel;
import com.intellij.ui.components.JBPanelWithEmptyText;
import com.intellij.ui.jcef.JBCefApp;
import com.intellij.util.ui.JBUI;
import com.sourcegraph.browser.JSToJavaBridgeRequestHandler;
import com.sourcegraph.browser.SourcegraphJBCefBrowser;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.awt.*;

/**
 * Inspired by <a href="https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java">FindPopupPanel.java</a>
 */
public class FindPopupPanel extends JBPanel<FindPopupPanel> implements Disposable {
    private SourcegraphJBCefBrowser browser;
    private final Project project;
    private final Splitter splitter;
    private PreviewPanel previewPanel;

    public FindPopupPanel(@NotNull Project project) {
        super(new BorderLayout());

        this.project = project;

        setPreferredSize(JBUI.size(1200, 800));
        setBorder(PopupBorder.Factory.create(true, true));
        setFocusCycleRoot(true);

        splitter = new OnePixelSplitter(true, 0.5f, 0.1f, 0.9f);
        add(splitter, BorderLayout.CENTER);

        createPreviewPanel();
        createBrowserPanel();
    }

    private void createBrowserPanel() {
        JBPanelWithEmptyText overlayPanel = new JBPanelWithEmptyText();
        //noinspection DialogTitleCapitalization
        overlayPanel.getEmptyText().setText("Loading Sourcegraph...");

        JBPanelWithEmptyText jcefPanel = new JBPanelWithEmptyText(new BorderLayout()).withEmptyText("Unfortunately, the browser is not available on your system. Try running the IDE with the default OpenJDK.");

        BrowserAndLoadingPanel topPanel = new BrowserAndLoadingPanel(jcefPanel, overlayPanel);

        browser = JBCefApp.isSupported() ? new SourcegraphJBCefBrowser(new JSToJavaBridgeRequestHandler(project, previewPanel, topPanel)) : null;
        if (browser != null) {
            jcefPanel.add(browser.getComponent());
        }
        splitter.setFirstComponent(topPanel);
    }

    private void createPreviewPanel() {
        previewPanel = new PreviewPanel(project);
        splitter.setSecondComponent(previewPanel);
    }

    @Nullable
    public SourcegraphJBCefBrowser getBrowser() {
        return browser;
    }

    @Nullable
    public PreviewPanel getPreviewPanel() {
        return previewPanel;
    }

    @Override
    public void dispose() {
        if (browser != null) {
            browser.dispose();
        }

        previewPanel.dispose();
    }
}
