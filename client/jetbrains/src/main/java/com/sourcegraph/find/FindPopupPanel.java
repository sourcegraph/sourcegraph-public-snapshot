package com.sourcegraph.find;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.ide.CopyPasteManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.Splitter;
import com.intellij.ui.OnePixelSplitter;
import com.intellij.ui.PopupBorder;
import com.intellij.ui.jcef.JBCefApp;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.components.BorderLayoutPanel;
import com.sourcegraph.Icons;
import com.sourcegraph.find.browser.BrowserAndLoadingPanel;
import com.sourcegraph.find.browser.JSToJavaBridgeRequestHandler;
import com.sourcegraph.find.browser.JavaToJSBridge;
import com.sourcegraph.find.browser.SourcegraphJBCefBrowser;
import java.awt.*;
import java.awt.datatransfer.StringSelection;
import java.util.Date;
import javax.swing.*;
import org.jdesktop.swingx.util.OS;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * Inspired by <a
 * href="https://sourcegraph.com/github.com/JetBrains/intellij-community/-/blob/platform/lang-impl/src/com/intellij/find/impl/FindPopupPanel.java">FindPopupPanel.java</a>
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

    setPreferredSize(JBUI.size(1000, 700));
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
    JSToJavaBridgeRequestHandler requestHandler =
        new JSToJavaBridgeRequestHandler(project, this, findService);
    browser = JBCefApp.isSupported() ? new SourcegraphJBCefBrowser(requestHandler) : null;
    if (browser == null) {
      showNoBrowserErrorNotification();
      Logger logger = Logger.getInstance(JSToJavaBridgeRequestHandler.class);
      logger.warn("JCEF browser is not supported!");
    } else {
      browserAndLoadingPanel.setBrowser(browser);
    }
    // The border is needed on macOS because without it, window and splitter resize don't work
    // because the JCEF
    // doesn't properly pass the mouse events to Swing.
    // 4px is the minimum amount to make it work for the window resize, the splitter works without a
    // padding.
    JPanel browserContainerForOptionalBorder = new JPanel(new BorderLayout());
    if (OS.isMacOSX()) {
      browserContainerForOptionalBorder.setBorder(JBUI.Borders.empty(0, 4, 5, 4));
    }
    browserContainerForOptionalBorder.add(browserAndLoadingPanel, BorderLayout.CENTER);

    HeaderPanel headerPanel = new HeaderPanel();

    BorderLayoutPanel topPanel = new BorderLayoutPanel();
    topPanel.add(headerPanel, BorderLayout.NORTH);
    topPanel.add(browserContainerForOptionalBorder, BorderLayout.CENTER);
    topPanel.setMinimumSize(JBUI.size(750, 200));

    splitter.setFirstComponent(topPanel);
    splitter.setSecondComponent(bottomPanel);

    lastPreviewUpdate = new Date();

    UIManager.addPropertyChangeListener(
        propertyChangeEvent -> {
          if (propertyChangeEvent.getPropertyName().equals("lookAndFeel")) {
            SwingUtilities.updateComponentTreeUI(this);
          }
        });
  }

  private void showNoBrowserErrorNotification() {
    Notification notification =
        new Notification(
            "Sourcegraph errors",
            "Sourcegraph",
            "Your IDE doesn't support JCEF. You won't be able to use \"Find with Sourcegraph\". If you believe this is an error, please raise this at support@sourcegraph.com, specifying your OS and IDE version.",
            NotificationType.ERROR);
    AnAction copyEmailAddressAction =
        new DumbAwareAction("Copy Support Email Address") {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            CopyPasteManager.getInstance()
                .setContents(new StringSelection("support@sourcegraph.com"));
            notification.expire();
          }
        };
    AnAction dismissAction =
        new DumbAwareAction("Dismiss") {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            notification.expire();
          }
        };
    notification.setIcon(Icons.CodyLogo);
    notification.addAction(copyEmailAddressAction);
    notification.addAction(dismissAction);
    Notifications.Bus.notify(notification);
  }

  @Nullable
  public PreviewPanel getPreviewPanel() {
    return previewPanel;
  }

  @Nullable
  public JavaToJSBridge getJavaToJSBridge() {
    return browser != null ? browser.getJavaToJSBridge() : null;
  }

  public BrowserAndLoadingPanel.ConnectionAndAuthState getConnectionAndAuthState() {
    return browserAndLoadingPanel.getConnectionAndAuthState();
  }

  public boolean browserHasSearchError() {
    return browserAndLoadingPanel.hasSearchError();
  }

  public void indicateAuthenticationStatus(
      boolean wasServerAccessSuccessful, boolean authenticated) {
    browserAndLoadingPanel.setConnectionAndAuthState(
        wasServerAccessSuccessful
            ? (authenticated
                ? BrowserAndLoadingPanel.ConnectionAndAuthState.AUTHENTICATED
                : BrowserAndLoadingPanel.ConnectionAndAuthState.COULD_CONNECT_BUT_NOT_AUTHENTICATED)
            : BrowserAndLoadingPanel.ConnectionAndAuthState.COULD_NOT_CONNECT);

    if (wasServerAccessSuccessful) {
      previewPanel.setState(PreviewPanel.State.PREVIEW_AVAILABLE);
      footerPanel.setPreviewContent(previewPanel.getPreviewContent());
    } else {
      selectionMetadataPanel.clearSelectionMetadataLabel();
      previewPanel.setState(PreviewPanel.State.NO_PREVIEW_AVAILABLE);
      footerPanel.setPreviewContent(null);
    }
  }

  public void indicateSearchError(@NotNull String errorMessage, @NotNull Date date) {
    if (lastPreviewUpdate.before(date)) {
      this.lastPreviewUpdate = date;
      browserAndLoadingPanel.setBrowserSearchErrorMessage(errorMessage);
      selectionMetadataPanel.clearSelectionMetadataLabel();
      previewPanel.setState(PreviewPanel.State.NO_PREVIEW_AVAILABLE);
      footerPanel.setPreviewContent(null);
    }
  }

  public void indicateLoadingIfInTime(@NotNull Date date) {
    if (lastPreviewUpdate.before(date)) {
      this.lastPreviewUpdate = date;
      selectionMetadataPanel.clearSelectionMetadataLabel();
      previewPanel.setState(PreviewPanel.State.LOADING);
      footerPanel.setPreviewContent(null);
    }
  }

  public void setPreviewContentIfInTime(@NotNull PreviewContent previewContent) {
    if (lastPreviewUpdate.before(previewContent.getReceivedDateTime())) {
      this.lastPreviewUpdate = previewContent.getReceivedDateTime();
      browserAndLoadingPanel.setBrowserSearchErrorMessage(null);
      selectionMetadataPanel.setSelectionMetadataLabel(previewContent);
      previewPanel.setContent(previewContent);
      footerPanel.setPreviewContent(previewContent);
    }
  }

  public void clearPreviewContentIfInTime(@NotNull Date date) {
    if (lastPreviewUpdate.before(date)) {
      this.lastPreviewUpdate = date;
      browserAndLoadingPanel.setBrowserSearchErrorMessage(null);
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
