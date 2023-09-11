package com.sourcegraph.cody.context;

import static com.sourcegraph.cody.chat.ChatUIConstants.TEXT_MARGIN;

import com.intellij.openapi.fileEditor.FileEditorManagerListener;
import com.intellij.openapi.project.Project;
import com.intellij.ui.SimpleColoredComponent;
import com.intellij.ui.SimpleTextAttributes;
import com.intellij.ui.components.JBLabel;
import com.intellij.util.ui.JBUI;
import java.awt.FlowLayout;
import javax.swing.Box;
import javax.swing.Icon;
import javax.swing.JPanel;
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

    this.setEmbeddingStatus(new NoGitRepositoryEmbeddingStatus());
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

  public void setOpenedFileName(@NotNull String fileName, @Nullable String filePath) {
    openedFileContent.setText(fileName);
    openedFileContent.setToolTipText(filePath);
  }
}
