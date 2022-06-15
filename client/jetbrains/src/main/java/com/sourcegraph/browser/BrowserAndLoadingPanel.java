package com.sourcegraph.browser;

import com.intellij.ui.components.JBPanelWithEmptyText;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;

/**
 * Inspired by <a href="https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java">FindPopupPanel.java</a>
 */
public class BrowserAndLoadingPanel extends JLayeredPane {
    private boolean isBrowserVisible = false;
    private final JBPanelWithEmptyText jcefPanel;

    public BrowserAndLoadingPanel() {
        jcefPanel = new JBPanelWithEmptyText(new BorderLayout()).withEmptyText(
            "Unfortunately, the browser is not available on your system. Try running the IDE with the default OpenJDK.");

        JBPanelWithEmptyText overlayPanel = new JBPanelWithEmptyText();
        //noinspection DialogTitleCapitalization
        overlayPanel.getEmptyText().setText("Loading Sourcegraph...");

        add(overlayPanel, 0);
        add(jcefPanel, 1);
    }

    public void setBrowser(@NotNull SourcegraphJBCefBrowser browser) {
        jcefPanel.add(browser.getComponent());
    }

    @Override
    public void doLayout() {
        Component overlay = getComponent(0);
        Component browser = getComponent(1);
        if (isBrowserVisible) {
            browser.setBounds(0, 0, getWidth(), getHeight());
        } else {
            browser.setBounds(0, 0, 1, 1);
        }
        overlay.setBounds(0, 0, getWidth(), getHeight());
    }

    @Override
    public Dimension getPreferredSize() {
        return getBounds().getSize();
    }

    public void setBrowserVisible(boolean browserVisible) {
        isBrowserVisible = browserVisible;
        revalidate();
        repaint();
    }
}
