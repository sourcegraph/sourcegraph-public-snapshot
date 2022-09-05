package com.sourcegraph.find.browser;

import com.intellij.openapi.options.ShowSettingsUtil;
import com.intellij.openapi.project.Project;
import com.intellij.ui.SimpleTextAttributes;
import com.intellij.ui.components.JBPanelWithEmptyText;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.StatusText;
import com.sourcegraph.config.SettingsConfigurable;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;

import static com.intellij.ui.SimpleTextAttributes.STYLE_PLAIN;

/**
 * Inspired by <a href="https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java">FindPopupPanel.java</a>
 */
public class BrowserAndLoadingPanel extends JLayeredPane {
    private final Project project;
    private final JBPanelWithEmptyText overlayPanel;
    private final JBPanelWithEmptyText jcefPanel;
    private boolean isBrowserVisible = false;

    public BrowserAndLoadingPanel(Project project) {
        this.project = project;
        jcefPanel = new JBPanelWithEmptyText(new BorderLayout()).withEmptyText(
            "Unfortunately, the browser is not available on your system. Try running the IDE with the default OpenJDK.");

        overlayPanel = new JBPanelWithEmptyText();
        setState(State.LOADING);

        // We need to use the add(Component, Object) overload of the add method to ensure that the constraints are
        // properly set.
        add(overlayPanel, Integer.valueOf(1));
        add(jcefPanel, Integer.valueOf(2));
    }

    public void setState(State state) {
        StatusText emptyText = overlayPanel.getEmptyText();

        isBrowserVisible = state == State.AUTHENTICATED || state == State.COULD_CONNECT_BUT_NOT_AUTHENTICATED;
        if (state == State.LOADING) {
            emptyText.setText("Loading...");
        } else if (state == State.COULD_NOT_CONNECT) {
            emptyText.setText("Could not connect to Sourcegraph.");
            emptyText.appendLine("Make sure your Sourcegraph URL and access token are correct to use search.");
            emptyText.appendLine("Click here to configure your Sourcegraph settings.",
                new SimpleTextAttributes(STYLE_PLAIN, JBUI.CurrentTheme.Link.Foreground.ENABLED),
                __ -> ShowSettingsUtil.getInstance().showSettingsDialog(project, SettingsConfigurable.class)
            );
        } else {
            // We need to do this because the "COULD_NOT_CONNECT" link is clickable even when the empty text is hidden! :o
            emptyText.setText("");
        }
        revalidate();
        repaint();
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

    public enum State {
        LOADING,
        AUTHENTICATED,
        COULD_NOT_CONNECT,
        COULD_CONNECT_BUT_NOT_AUTHENTICATED
    }
}
