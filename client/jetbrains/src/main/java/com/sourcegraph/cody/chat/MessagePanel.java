package com.sourcegraph.cody.chat;

import com.intellij.openapi.ui.VerticalFlowLayout;
import com.intellij.ui.ColorUtil;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.cody.ui.Colors;
import java.awt.*;
import javax.swing.*;
import javax.swing.border.Border;
import org.jetbrains.annotations.NotNull;

public class MessagePanel extends JPanel {

  private final boolean isHuman;
  private final int gradientWidth;

  public MessagePanel(Speaker speaker, int gradientWidth) {
    super();
    this.gradientWidth = gradientWidth;
    Border emptyBorder = BorderFactory.createEmptyBorder(0, 0, 0, 0);
    Color background = UIUtil.getPanelBackground();
    Border topBorder =
        BorderFactory.createMatteBorder(1, 0, 0, 0, ColorUtil.brighter(background, 2));
    Border bottomBorder =
        BorderFactory.createMatteBorder(0, 0, 1, 0, ColorUtil.brighter(background, 3));
    Border topAndBottomBorder = BorderFactory.createCompoundBorder(topBorder, bottomBorder);

    isHuman = speaker.equals(Speaker.HUMAN);
    this.setBorder(isHuman ? emptyBorder : topAndBottomBorder);
    this.setBackground(isHuman ? ColorUtil.darker(background, 2) : background);
    this.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false));
  }

  @Override
  protected void paintComponent(@NotNull Graphics g) {
    super.paintComponent(g);
    paintLeftBorderGradient(g);
  }

  private void paintLeftBorderGradient(Graphics g) {
    if (isHuman) return;
    int panelHeight = getHeight();
    int halfOfHeight = panelHeight / 2;

    GradientPaint firstPartGradient =
        new GradientPaint(0, 0, Colors.PURPLE, 0, halfOfHeight, Colors.ORANGE);
    GradientPaint secondPartGradient =
        new GradientPaint(0, halfOfHeight, Colors.ORANGE, 0, panelHeight, Colors.CYAN);
    final Graphics2D g2d = (Graphics2D) g;
    g2d.setPaint(firstPartGradient);
    g2d.fillRect(0, 0, gradientWidth, halfOfHeight);
    g2d.setPaint(secondPartGradient);
    g2d.fillRect(0, halfOfHeight, gradientWidth, panelHeight);
  }
}
