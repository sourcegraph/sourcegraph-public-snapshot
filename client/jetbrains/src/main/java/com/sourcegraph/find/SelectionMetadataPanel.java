package com.sourcegraph.find;

import com.intellij.icons.AllIcons;
import com.intellij.openapi.actionSystem.KeyboardShortcut;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.keymap.KeymapUtil;
import com.intellij.ui.components.labels.LinkLabel;
import java.awt.*;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.util.Objects;
import javax.swing.*;
import org.jetbrains.annotations.NotNull;

public class SelectionMetadataPanel extends JPanel {
  private LinkLabel<String> selectionMetadataLabel = null;
  private final JLabel externalLinkLabel;
  private final JLabel openShortcutLabel;
  private PreviewContent previewContent;

  public SelectionMetadataPanel() {
    super(new FlowLayout(FlowLayout.LEFT, 0, 8));

    selectionMetadataLabel =
        new LinkLabel<>(
            "",
            null,
            (aSource, aLinkData) -> {
              if (previewContent != null) {
                try {
                  previewContent.openInEditorOrBrowser();
                } catch (Exception e) {
                  Logger logger = Logger.getInstance(SelectionMetadataPanel.class);
                  logger.warn(
                      "Error opening file in editor: \"" + selectionMetadataLabel.getText() + "\"",
                      e);
                }
              }
            });

    selectionMetadataLabel.setBorder(BorderFactory.createEmptyBorder(0, 5, 0, 0));
    externalLinkLabel = new JLabel("", AllIcons.Ide.External_link_arrow, SwingConstants.LEFT);
    externalLinkLabel.setVisible(false);
    KeyboardShortcut altEnterShortcut =
        new KeyboardShortcut(
            KeyStroke.getKeyStroke(KeyEvent.VK_ENTER, InputEvent.ALT_DOWN_MASK), null);
    String altEnterShortcutText = KeymapUtil.getShortcutText(altEnterShortcut);
    openShortcutLabel = new JLabel(altEnterShortcutText);
    openShortcutLabel.setBorder(BorderFactory.createEmptyBorder(0, 8, 0, 0));
    openShortcutLabel.setEnabled(false);
    openShortcutLabel.setVisible(false);

    add(selectionMetadataLabel);
    add(externalLinkLabel);
    add(openShortcutLabel);
  }

  public void clearSelectionMetadataLabel() {
    previewContent = null;
    selectionMetadataLabel.setText("");
    externalLinkLabel.setVisible(false);
    openShortcutLabel.setVisible(false);
  }

  public void setSelectionMetadataLabel(@NotNull PreviewContent previewContent) {
    this.previewContent = previewContent;
    String metadataText = getMetadataText(previewContent);
    selectionMetadataLabel.setText(metadataText);
    externalLinkLabel.setVisible(!previewContent.opensInEditor());
    openShortcutLabel.setToolTipText(
        "Press "
            + openShortcutLabel.getText()
            + " to open the selected file"
            + (previewContent.opensInEditor() ? " in the editor." : " in your browser."));
    openShortcutLabel.setVisible(true);
  }

  @NotNull
  private String getMetadataText(@NotNull PreviewContent previewContent) {
    if (Objects.equals(previewContent.getResultType(), "file")
        || Objects.equals(previewContent.getResultType(), "path")) {
      return previewContent.getRepoUrl() + ":" + previewContent.getPath();
    } else if (Objects.equals(previewContent.getResultType(), "repo")) {
      return previewContent.getRepoUrl();
    } else if (Objects.equals(previewContent.getResultType(), "symbol")) {
      return previewContent.getSymbolName() + " (" + previewContent.getSymbolContainerName() + ")";
    } else if (Objects.equals(previewContent.getResultType(), "diff")) {
      return previewContent.getCommitMessagePreview() != null
          ? previewContent.getCommitMessagePreview()
          : "";
    } else if (Objects.equals(previewContent.getResultType(), "commit")) {
      return previewContent.getCommitMessagePreview() != null
          ? previewContent.getCommitMessagePreview()
          : "";
    } else {
      return "";
    }
  }
}
