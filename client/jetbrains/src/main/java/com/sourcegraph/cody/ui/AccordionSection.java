package com.sourcegraph.cody.ui;

import com.intellij.openapi.ui.VerticalFlowLayout;
import java.awt.*;
import javax.swing.*;

public class AccordionSection extends JPanel {
  private final JButton toggleButton;
  private final String sectionTitle;
  private JPanel contentPanel;

  public AccordionSection(String title) {
    setLayout(new BorderLayout());
    sectionTitle = title;

    toggleButton = new JButton(createToggleButtonHTML(title, true));
    toggleButton.setHorizontalAlignment(SwingConstants.LEFT);
    toggleButton.setBorderPainted(false);
    toggleButton.setFocusPainted(false);
    toggleButton.setContentAreaFilled(false);
    toggleButton.addActionListener(
        e -> {
          if (contentPanel.isVisible()) {
            contentPanel.setVisible(false);
            toggleButton.setText(createToggleButtonHTML(sectionTitle, true));
          } else {
            contentPanel.setVisible(true);
            toggleButton.setText(createToggleButtonHTML(sectionTitle, false));
          }
        });

    contentPanel = new JPanel();
    contentPanel.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false));
    contentPanel.setVisible(false);

    add(toggleButton, BorderLayout.NORTH);
    add(contentPanel, BorderLayout.CENTER);
  }

  private String createToggleButtonHTML(String title, boolean isCollapsed) {
    String symbol =
        isCollapsed ? "&#9658;" : "&#9660;"; // Unicode entities for right and down arrows
    return "<html><body style='text-align:left'>"
        + title
        + " <span style='float:right; color:gray'>"
        + symbol
        + "</span></body></html>";
  }

  public JPanel getContentPanel() {
    return contentPanel;
  }
}
