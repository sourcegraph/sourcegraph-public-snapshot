package com.sourcegraph.find;

import com.intellij.icons.AllIcons;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.actionSystem.KeyboardShortcut;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.keymap.KeymapUtil;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.Splitter;
import com.intellij.ui.OnePixelSplitter;
import com.intellij.ui.PopupBorder;
import com.intellij.ui.components.JBPanel;
import com.intellij.ui.jcef.JBCefApp;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.components.BorderLayoutPanel;
import com.sourcegraph.browser.BrowserAndLoadingPanel;
import com.sourcegraph.browser.JSToJavaBridgeRequestHandler;
import com.sourcegraph.browser.SourcegraphJBCefBrowser;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import javax.swing.*;
import java.awt.*;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.util.Objects;

/**
 * Inspired by <a href="https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java">FindPopupPanel.java</a>
 */
public class FindPopupPanel extends JBPanel<FindPopupPanel> implements Disposable {
    private final SourcegraphJBCefBrowser browser;
    private final PreviewPanel previewPanel;
    private final BrowserAndLoadingPanel browserAndLoadingPanel;
    private JLabel selectionMetadataLabel;
    private JLabel externalLinkLabel;
    private JLabel openShortcutLabel;

    public FindPopupPanel(@NotNull Project project) {
        super(new BorderLayout());

        setPreferredSize(JBUI.size(1200, 800));
        setBorder(PopupBorder.Factory.create(true, true));
        setFocusCycleRoot(true);

        Splitter splitter = new OnePixelSplitter(true, 0.5f, 0.1f, 0.9f);
        add(splitter, BorderLayout.CENTER);

        previewPanel = new PreviewPanel(project);
        JPanel bottomPanel = createBottomPanel();

        browserAndLoadingPanel = new BrowserAndLoadingPanel();
        JSToJavaBridgeRequestHandler requestHandler = new JSToJavaBridgeRequestHandler(project, this);
        browser = JBCefApp.isSupported() ? new SourcegraphJBCefBrowser(requestHandler) : null;
        if (browser != null) {
            browserAndLoadingPanel.setBrowser(browser);
        }
        splitter.setFirstComponent(browserAndLoadingPanel);
        splitter.setSecondComponent(bottomPanel);
    }

    @NotNull
    private BorderLayoutPanel createBottomPanel() {
        BorderLayoutPanel bottomPanel = new BorderLayoutPanel();
        JPanel selectionMetadataPanel = new JPanel(new FlowLayout(FlowLayout.LEFT, 5, 10));

        selectionMetadataLabel = new JLabel();
        externalLinkLabel = new JLabel("", AllIcons.Ide.External_link_arrow, SwingConstants.LEFT);
        KeyboardShortcut altEnterShortcut = new KeyboardShortcut(KeyStroke.getKeyStroke(KeyEvent.VK_ENTER, InputEvent.ALT_DOWN_MASK), null);
        String altEnterShortcutText = KeymapUtil.getShortcutText(altEnterShortcut);
        openShortcutLabel = new JLabel(altEnterShortcutText);

        selectionMetadataPanel.add(selectionMetadataLabel);
        selectionMetadataPanel.add(externalLinkLabel);
        selectionMetadataPanel.add(openShortcutLabel);
        bottomPanel.add(selectionMetadataPanel, BorderLayout.NORTH);
        bottomPanel.add(previewPanel, BorderLayout.CENTER);
        return bottomPanel;
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
        logger.info("setPreviewContent called.");
        previewPanel.setContent(previewContent);
    }

    public void clearPreviewContent() {
        previewPanel.clearContent();
    }

    public void setBrowserVisible(boolean visible) {
        browserAndLoadingPanel.setBrowserVisible(visible);
    }

    public void clearSelectionMetadataLabel() {
        selectionMetadataLabel.setText("");
        externalLinkLabel.setVisible(false);
        openShortcutLabel.setVisible(false);
    }

    public void setSelectionMetadataLabel(@NotNull PreviewContent previewContent) {
        String metadataText = getMetadataText(previewContent);
        selectionMetadataLabel.setText(metadataText);
        externalLinkLabel.setVisible(!previewContent.opensInEditor());
        openShortcutLabel.setToolTipText("Press " + openShortcutLabel.getText() + " to open the selected file" +
            (previewContent.opensInEditor() ? " in the editor." : " in your browser."));
        openShortcutLabel.setVisible(true);
    }

    @NotNull
    private String getMetadataText(@NotNull PreviewContent previewContent) {
        if (Objects.equals(previewContent.getResultType(), "file") || Objects.equals(previewContent.getResultType(), "path")) {
            return previewContent.getRepoUrl() + ":" + previewContent.getPath();
        } else if (Objects.equals(previewContent.getResultType(), "repo")) {
            return previewContent.getRepoUrl();
        } else if (Objects.equals(previewContent.getResultType(), "symbol")) {
            return previewContent.getSymbolName() + " (" + previewContent.getSymbolContainerName() + ")";
        } else if (Objects.equals(previewContent.getResultType(), "diff")) {
            return previewContent.getCommitMessagePreview() != null ? previewContent.getCommitMessagePreview() : "";
        } else if (Objects.equals(previewContent.getResultType(), "commit")) {
            return previewContent.getCommitMessagePreview() != null ? previewContent.getCommitMessagePreview() : "";
        } else {
            return "";
        }
    }
}
