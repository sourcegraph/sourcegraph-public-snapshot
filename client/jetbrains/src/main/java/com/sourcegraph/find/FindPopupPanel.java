package com.sourcegraph.find;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.Splitter;
import com.intellij.ui.OnePixelSplitter;
import com.intellij.ui.PopupBorder;
import com.intellij.ui.jcef.JBCefApp;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.components.BorderLayoutPanel;
import com.sourcegraph.browser.BrowserAndLoadingPanel;
import com.sourcegraph.browser.JSToJavaBridgeRequestHandler;
import com.sourcegraph.browser.SourcegraphJBCefBrowser;
import org.jdesktop.swingx.util.OS;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.awt.*;
import java.util.Date;

/**
 * Inspired by <a href="https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java">FindPopupPanel.java</a>
 */
public class FindPopupPanel extends BorderLayoutPanel implements Disposable {
    private final SourcegraphJBCefBrowser browser;
    private final PreviewPanel previewPanel;
    private final BrowserAndLoadingPanel browserAndLoadingPanel;
    private final SelectionMetadataPanel selectionMetadataPanel;
    private final FooterPanel footerPanel;
    private Date lastPreviewUpdate;

    public FindPopupPanel(@NotNull Project project, @NotNull FindService findService) {
        super();

        setPreferredSize(JBUI.size(1200, 800));
        setBorder(PopupBorder.Factory.create(true, true));
        setFocusCycleRoot(true);

        Splitter splitter = new OnePixelSplitter(true, 0.5f, 0.1f, 0.9f);
        add(splitter, BorderLayout.CENTER);

        selectionMetadataPanel = new SelectionMetadataPanel();
        previewPanel = new PreviewPanel(project);
        footerPanel = new FooterPanel();

        BorderLayoutPanel bottomPanel = new BorderLayoutPanel();
        bottomPanel.add(selectionMetadataPanel, BorderLayout.NORTH);
        bottomPanel.add(previewPanel, BorderLayout.CENTER);
        bottomPanel.add(footerPanel, BorderLayout.SOUTH);

        browserAndLoadingPanel = new BrowserAndLoadingPanel(project);
        JSToJavaBridgeRequestHandler requestHandler = new JSToJavaBridgeRequestHandler(project, this, findService);
        browser = JBCefApp.isSupported() ? new SourcegraphJBCefBrowser(requestHandler) : null;
        if (browser != null) {
            browserAndLoadingPanel.setBrowser(browser);
        }

        BorderLayoutPanel topPanel = new BorderLayoutPanel();
        // The border is needed on macOS because without it, window and splitter resize don't work because the JCEF
        // doesn't properly pass the mouse events to Swing.
        // 4px is the minimum amount to make it work for the window resize, the splitter works without a padding.
        if (OS.isMacOSX()) {
            topPanel.setBorder(JBUI.Borders.empty(0, 4, 5, 4));
        }
        topPanel.add(browserAndLoadingPanel, BorderLayout.CENTER);
        topPanel.setMinimumSize(JBUI.size(750, 200));

        splitter.setFirstComponent(topPanel);
        splitter.setSecondComponent(bottomPanel);

        lastPreviewUpdate = new Date();
    }

    @Nullable
    public SourcegraphJBCefBrowser getBrowser() {
        return browser;
    }

    @Nullable
    public PreviewPanel getPreviewPanel() {
        return previewPanel;
    }

    public void indicateAuthenticationStatus(boolean wasServerAccessSuccessful, boolean authenticated) {
        browserAndLoadingPanel.setState(wasServerAccessSuccessful
            ? (authenticated ? BrowserAndLoadingPanel.State.AUTHENTICATED : BrowserAndLoadingPanel.State.COULD_CONNECT_BUT_NOT_AUTHENTICATED)
            : BrowserAndLoadingPanel.State.COULD_NOT_CONNECT);
        if (!wasServerAccessSuccessful) {
            selectionMetadataPanel.clearSelectionMetadataLabel();
            previewPanel.setState(PreviewPanel.State.NO_PREVIEW_AVAILABLE);
            footerPanel.setPreviewContent(null);
        } else {
            previewPanel.setState(PreviewPanel.State.PREVIEW_AVAILABLE);
            footerPanel.setPreviewContent(previewPanel.getPreviewContent());
        }
    }

    public void indicateLoadingIfInTime(@NotNull Date date) {
        if (lastPreviewUpdate.before(date)) {
            selectionMetadataPanel.clearSelectionMetadataLabel();
            previewPanel.setState(PreviewPanel.State.LOADING);
            footerPanel.setPreviewContent(null);
        }
    }

    public void setPreviewContentIfInTime(@NotNull PreviewContent previewContent) {
        if (lastPreviewUpdate.before(previewContent.getReceivedDateTime())) {
            this.lastPreviewUpdate = previewContent.getReceivedDateTime();
            selectionMetadataPanel.setSelectionMetadataLabel(previewContent);
            previewPanel.setContent(previewContent);
            footerPanel.setPreviewContent(previewContent);
        }
    }

    public void clearPreviewContentIfInTime(@NotNull Date date) {
        if (lastPreviewUpdate.before(date)) {
            this.lastPreviewUpdate = date;
            selectionMetadataPanel.clearSelectionMetadataLabel();
            previewPanel.setState(PreviewPanel.State.NO_PREVIEW_AVAILABLE);
            footerPanel.setPreviewContent(null);
        }
    }

    @Override
    public void dispose() {
        if (browser != null) {
            browser.dispose();
        }

        previewPanel.dispose();
    }
}
