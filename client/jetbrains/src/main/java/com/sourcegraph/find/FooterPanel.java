package com.sourcegraph.find;

import com.intellij.openapi.actionSystem.KeyboardShortcut;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.keymap.KeymapUtil;
import com.intellij.ui.components.JBPanel;
import java.awt.*;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import javax.swing.*;
import org.jetbrains.annotations.Nullable;

public class FooterPanel extends JBPanel<FooterPanel> {
  private final JButton openButton;
  private final JLabel openShortcutLabel;
  @Nullable private PreviewContent previewContent;

  public FooterPanel() {
    super(new FlowLayout(FlowLayout.RIGHT));

    KeyboardShortcut altEnterShortcut =
        new KeyboardShortcut(
            KeyStroke.getKeyStroke(KeyEvent.VK_ENTER, InputEvent.ALT_DOWN_MASK), null);
    String altEnterShortcutText = KeymapUtil.getShortcutText(altEnterShortcut);
    openShortcutLabel = new JLabel(altEnterShortcutText);
    openShortcutLabel.setEnabled(false);
    openShortcutLabel.setVisible(false);

    openButton = new JButton("");
    openButton.addActionListener(
        event -> {
          if (previewContent != null) {
            try {
              previewContent.openInEditorOrBrowser();
            } catch (Exception e) {
              Logger logger = Logger.getInstance(FooterPanel.class);
              logger.warn(
                  "Error while opening preview content externally: "
                      + e.getClass().getName()
                      + ": "
                      + e.getMessage());
            }
          }
        });

    add(openShortcutLabel);
    add(openButton);

    setPreviewContent(null);
  }

  public void setPreviewContent(@Nullable PreviewContent previewContent) {
    this.previewContent = previewContent;
    openShortcutLabel.setVisible(previewContent != null);
    openButton.setEnabled(previewContent != null);
    openButton.setText(
        (previewContent == null || previewContent.opensInEditor())
            ? "Open in Editor"
            : "Open in Browser");
  }
}
