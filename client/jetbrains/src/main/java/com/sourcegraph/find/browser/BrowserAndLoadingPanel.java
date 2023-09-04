package com.sourcegraph.find.browser;

import static com.intellij.ui.SimpleTextAttributes.STYLE_PLAIN;

import com.intellij.openapi.options.ShowSettingsUtil;
import com.intellij.openapi.project.Project;
import com.intellij.ui.SimpleTextAttributes;
import com.intellij.ui.components.JBPanelWithEmptyText;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.StatusText;
import com.sourcegraph.config.SettingsConfigurable;
import java.awt.*;
import javax.swing.*;
import org.apache.commons.lang.SystemUtils;
import org.apache.commons.lang.WordUtils;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * Inspired by <a
 * href="https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java">FindPopupPanel.java</a>
 */
public class BrowserAndLoadingPanel extends JLayeredPane {
  private final Project project;
  private final JBPanelWithEmptyText overlayPanel;
  private final JBPanelWithEmptyText jcefPanel;
  private boolean isBrowserVisible = false;
  private ConnectionAndAuthState connectionAndAuthState = ConnectionAndAuthState.LOADING;
  private String errorMessage = null;

  public BrowserAndLoadingPanel(Project project) {
    this.project = project;
    jcefPanel =
        new JBPanelWithEmptyText(new BorderLayout())
            .withEmptyText(
                "Unfortunately, the browser is not available on your system. Try running the IDE with the default OpenJDK.");

    overlayPanel = new JBPanelWithEmptyText();
    setConnectionAndAuthState(ConnectionAndAuthState.LOADING);

    // We need to use the add(Component, Object) overload of the add method to ensure that the
    // constraints are
    // properly set.
    add(overlayPanel, Integer.valueOf(1));
    add(jcefPanel, Integer.valueOf(2));
  }

  public void setBrowserSearchErrorMessage(@Nullable String errorMessage) {
    this.errorMessage = errorMessage;
    refreshUI();
  }

  public void setConnectionAndAuthState(@NotNull ConnectionAndAuthState state) {
    this.connectionAndAuthState = state;
    refreshUI();
  }

  private void refreshUI() {
    StatusText emptyText = overlayPanel.getEmptyText();
    isBrowserVisible =
        errorMessage == null
            && (connectionAndAuthState == ConnectionAndAuthState.AUTHENTICATED
                || connectionAndAuthState
                    == ConnectionAndAuthState.COULD_CONNECT_BUT_NOT_AUTHENTICATED);

    if (connectionAndAuthState == ConnectionAndAuthState.COULD_NOT_CONNECT) {
      emptyText.setText("Could not connect to Sourcegraph.");
      emptyText.appendLine(
          "Make sure your Sourcegraph URL and access token are correct to use search.");
      emptyText.appendLine(
          "Click here to configure your Sourcegraph Cody + Code Search settings.",
          new SimpleTextAttributes(STYLE_PLAIN, JBUI.CurrentTheme.Link.Foreground.ENABLED),
          __ ->
              ShowSettingsUtil.getInstance()
                  .showSettingsDialog(project, SettingsConfigurable.class));

    } else if (errorMessage != null) {
      String wrappedText = WordUtils.wrap("Error: " + errorMessage, 100);
      String[] lines = wrappedText.split(SystemUtils.LINE_SEPARATOR);
      emptyText.setText(lines[0]);
      for (int i = 1; i < lines.length; i++) {
        if (!lines[i].trim().isEmpty()) {
          emptyText.appendLine(lines[i]);
        }
      }
      emptyText.appendLine("");
      emptyText.appendLine(
          "If you believe this is a bug, please raise this at support@sourcegraph.com,");
      //noinspection DialogTitleCapitalization
      emptyText.appendLine(
          "mentioning the above error message and your Cody plugin and Cody App or Sourcegraph server version.");
      emptyText.appendLine("Sorry for the inconvenience.");

    } else if (connectionAndAuthState == ConnectionAndAuthState.LOADING) {
      emptyText.setText("Loading...");
    } else {
      // We need to do this because the "COULD_NOT_CONNECT" link is clickable even when the empty
      // text is hidden! :o
      emptyText.setText("");
    }
    revalidate();
    repaint();
  }

  public ConnectionAndAuthState getConnectionAndAuthState() {
    return connectionAndAuthState;
  }

  public boolean hasSearchError() {
    return errorMessage != null;
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

  public enum ConnectionAndAuthState {
    LOADING,
    AUTHENTICATED,
    COULD_NOT_CONNECT,
    COULD_CONNECT_BUT_NOT_AUTHENTICATED
  }
}
