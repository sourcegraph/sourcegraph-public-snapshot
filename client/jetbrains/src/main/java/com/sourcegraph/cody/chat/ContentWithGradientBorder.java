package com.sourcegraph.cody.chat;

import com.sourcegraph.cody.ui.Colors;
import java.awt.GradientPaint;
import java.awt.Graphics;
import java.awt.Graphics2D;
import javax.swing.BoxLayout;
import javax.swing.JPanel;
import org.jetbrains.annotations.NotNull;

public class ContentWithGradientBorder extends JPanel {
  private final int gradientWidth;

  public ContentWithGradientBorder(int gradientWidth) {
    super();
    this.gradientWidth = gradientWidth;
    this.setLayout(new BoxLayout(this, BoxLayout.Y_AXIS));
  }

  @Override
  protected void paintComponent(@NotNull Graphics g) {
    super.paintComponent(g);
    paintBorderGradient(g);
  }

  private void paintBorderGradient(Graphics g) {
    int panelHeight = getHeight();
    int panelWidth = getWidth();
    int halfOfHeight = panelHeight / 2;

    GradientPaint firstPartGradient =
        new GradientPaint(0, 0, Colors.PURPLE, 0, halfOfHeight, Colors.ORANGE);
    GradientPaint secondPartGradient =
        new GradientPaint(0, halfOfHeight, Colors.ORANGE, 0, panelHeight, Colors.CYAN);
    final Graphics2D g2d = (Graphics2D) g;
    // top border
    g2d.setPaint(Colors.PURPLE);
    g2d.fillRect(0, 0, panelWidth, gradientWidth);
    // left border
    g2d.setPaint(firstPartGradient);
    g2d.fillRect(0, 0, gradientWidth, halfOfHeight);
    g2d.setPaint(secondPartGradient);
    g2d.fillRect(0, halfOfHeight, gradientWidth, panelHeight);
    // right border
    g2d.setPaint(firstPartGradient);
    g2d.fillRect(panelWidth - gradientWidth, 0, gradientWidth, halfOfHeight);
    g2d.setPaint(secondPartGradient);
    g2d.fillRect(panelWidth - gradientWidth, halfOfHeight, gradientWidth, panelHeight);
    // bottom border
    g2d.setPaint(Colors.CYAN);
    g2d.fillRect(0, panelHeight - gradientWidth, panelWidth, gradientWidth);
  }
}
