package com.sourcegraph.find;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.Splitter;
import com.intellij.ui.OnePixelSplitter;
import com.intellij.ui.PopupBorder;
import com.intellij.ui.components.JBPanel;
import com.intellij.ui.jcef.JBCefApp;
import com.intellij.util.ui.JBUI;
import com.sourcegraph.browser.BrowserAndLoadingPanel;
import com.sourcegraph.browser.JSToJavaBridgeRequestHandler;
import com.sourcegraph.browser.SourcegraphJBCefBrowser;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.awt.*;

/**
 * Inspired by <a href="https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java">FindPopupPanel.java</a>
 */
public class FindPopupPanel extends JBPanel<FindPopupPanel> implements Disposable {
    private final SourcegraphJBCefBrowser browser;
    private final PreviewPanel previewPanel;
    private final BrowserAndLoadingPanel browserAndLoadingPanel;

    public FindPopupPanel(@NotNull Project project) {
        super(new BorderLayout());

        setPreferredSize(JBUI.size(1200, 800));
        setBorder(PopupBorder.Factory.create(true, true));
        setFocusCycleRoot(true);

        Splitter splitter = new OnePixelSplitter(true, 0.5f, 0.1f, 0.9f);
        add(splitter, BorderLayout.CENTER);

        previewPanel = new PreviewPanel(project);

        browserAndLoadingPanel = new BrowserAndLoadingPanel();
        JSToJavaBridgeRequestHandler requestHandler = new JSToJavaBridgeRequestHandler(project, this);
        browser = JBCefApp.isSupported() ? new SourcegraphJBCefBrowser(requestHandler) : null;
        if (browser != null) {
            browserAndLoadingPanel.setBrowser(browser);
        }
        splitter.setFirstComponent(browserAndLoadingPanel);
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

    public void setPreviewContent(@NotNull PreviewContent previewContent) {
        Logger logger = Logger.getInstance(FindPopupPanel.class);
        logger.info("setPreviewContent: ");
        previewPanel.setContent(previewContent);
    }

    public void clearPreviewContent() {
        previewPanel.clearContent();
    }

    public void setBrowserVisible(boolean visible) {
        browserAndLoadingPanel.setBrowserVisible(visible);
    }
}
