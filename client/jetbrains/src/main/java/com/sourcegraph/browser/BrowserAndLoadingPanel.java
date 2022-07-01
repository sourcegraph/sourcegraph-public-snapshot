package com.sourcegraph.browser;

import com.intellij.ui.components.JBPanelWithEmptyText;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;

/**
 * Inspired by <a href="https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java">FindPopupPanel.java</a>
 */
public class BrowserAndLoadingPanel extends JLayeredPane {
    private final JBPanelWithEmptyText overlayPanel;
    private boolean isBrowserVisible = false;
    private final JBPanelWithEmptyText jcefPanel;

    public BrowserAndLoadingPanel() {
        jcefPanel = new JBPanelWithEmptyText(new BorderLayout()).withEmptyText(
            "Unfortunately, the browser is not available on your system. Try running the IDE with the default OpenJDK.");

        overlayPanel = new JBPanelWithEmptyText();
        //noinspection DialogTitleCapitalization
        overlayPanel.getEmptyText().setText("Loading Sourcegraph...");

        // We need to use the add(Component, Object) overload of the add method to ensure that the constraints are
        // properly set.
        add(overlayPanel, Integer.valueOf(1));
        add(jcefPanel, Integer.valueOf(2));
    }

    public void setBrowser(@NotNull SourcegraphJBCefBrowser browser) {
        jcefPanel.add(browser.getComponent());
    }

    @Override
    public void doLayout() {
        if (isBrowserVisible) {
            jcefPanel.setBounds(0, 0, getWidth(), getHeight());
        } else {
            jcefPanel.setBounds(0, 0, 1, 1);
        }
        overlayPanel.setBounds(0, 0, getWidth(), getHeight());
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
