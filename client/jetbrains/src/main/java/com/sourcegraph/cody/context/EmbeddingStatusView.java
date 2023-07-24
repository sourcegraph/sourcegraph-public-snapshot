package com.sourcegraph.cody.context;

import static com.sourcegraph.cody.chat.ChatUIConstants.TEXT_MARGIN;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.fileEditor.FileEditorManagerListener;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.serviceContainer.AlreadyDisposedException;
import com.intellij.ui.ColorUtil;
import com.intellij.ui.SimpleColoredComponent;
import com.intellij.ui.SimpleTextAttributes;
import com.intellij.ui.components.JBLabel;
import com.intellij.util.concurrency.EdtExecutorService;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.cody.context.embeddings.EmbeddingsStatusLoader;
import com.sourcegraph.cody.editor.EditorUtil;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.vcs.RepoUtil;
import java.awt.FlowLayout;
import java.util.concurrent.TimeUnit;
import javax.swing.BorderFactory;
import javax.swing.Box;
import javax.swing.Icon;
import javax.swing.JPanel;
import javax.swing.border.Border;
import javax.swing.border.EmptyBorder;
import org.apache.commons.lang.StringUtils;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class EmbeddingStatusView extends JPanel {

  private final @NotNull SimpleColoredComponent embeddingStatusContent;
  private final @NotNull JBLabel openedFileContent;
  private final @NotNull Project project;
  private @NotNull EmbeddingStatus embeddingStatus;

  public EmbeddingStatusView(@NotNull Project project) {
    super();
    this.project = project;
    this.setLayout(new FlowLayout(FlowLayout.LEFT));
    Border topBorder =
        BorderFactory.createMatteBorder(
            1, 0, 0, 0, ColorUtil.brighter(UIUtil.getPanelBackground(), 3));
    this.setBorder(topBorder);
    Box innerPanel = Box.createHorizontalBox();
    embeddingStatusContent = new SimpleColoredComponent();
    openedFileContent = new JBLabel();
    embeddingStatus = new EmbeddingStatusNotAvailableYet();
    updateViewBasedOnStatus();
    innerPanel.add(embeddingStatusContent);
    innerPanel.add(Box.createHorizontalStrut(5));
    innerPanel.add(openedFileContent);
    innerPanel.setBorder(new EmptyBorder(JBUI.insets(TEXT_MARGIN, TEXT_MARGIN, 0, TEXT_MARGIN)));
    this.add(innerPanel);

    EdtExecutorService.getScheduledExecutorInstance()
        .scheduleWithFixedDelay(this::updateEmbeddingStatusView, 1, 10, TimeUnit.SECONDS);
    project
        .getMessageBus()
        .connect()
        .subscribe(
            FileEditorManagerListener.FILE_EDITOR_MANAGER,
            new CurrentlyOpenFileListener(project, this));
  }

  private void updateViewBasedOnStatus() {
    embeddingStatusContent.clear();
    embeddingStatusContent.append(
        embeddingStatus.getMainText(), SimpleTextAttributes.REGULAR_ATTRIBUTES);
    Icon icon = embeddingStatus.getIcon();
    if (icon != null) {
      embeddingStatusContent.setIcon(icon);
    }
    String tooltip = embeddingStatus.getTooltip(project);
    if (StringUtils.isNotEmpty(tooltip)) {
      embeddingStatusContent.setToolTipText(tooltip);
    }
  }

  public void setEmbeddingStatus(EmbeddingStatus embeddingStatus) {
    this.embeddingStatus = embeddingStatus;
    updateViewBasedOnStatus();
  }

  private void updateEmbeddingStatusView() {
    VirtualFile currentFile;
    try {
      currentFile = EditorUtil.getCurrentFile(project);
    } catch (AlreadyDisposedException e) {
      return;
    }

    ApplicationManager.getApplication()
        .executeOnPooledThread(
            () -> {
              String repositoryName = RepoUtil.findRepositoryName(project, currentFile);
              String instanceUrl = ConfigUtil.getSourcegraphUrl(project);
              String accessToken = ConfigUtil.getProjectAccessToken(project);

              String accessTokenOrEmpty = accessToken != null ? accessToken : "";
              String repoId =
                  repositoryName != null
                      ? new EmbeddingsStatusLoader(
                              instanceUrl,
                              accessTokenOrEmpty,
                              ConfigUtil.getCustomRequestHeaders(project))
                          .getRepoId(repositoryName)
                      : null;
              ApplicationManager.getApplication()
                  .invokeLater(
                      () -> {
                        if (repositoryName == null) {
                          this.setEmbeddingStatus(new NoGitRepositoryEmbeddingStatus());
                        } else if (repoId == null) {
                          this.setEmbeddingStatus(
                              new RepositoryMissingEmbeddingStatus(repositoryName));
                        } else {
                          this.setEmbeddingStatus(
                              new RepositoryIndexedEmbeddingStatus(repositoryName));
                        }
                      });
            });
  }

  public void setOpenedFileName(@NotNull String fileName, @Nullable String filePath) {
    openedFileContent.setText(fileName);
    openedFileContent.setToolTipText(filePath);
  }
}
