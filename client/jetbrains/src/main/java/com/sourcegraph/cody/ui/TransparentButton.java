package com.sourcegraph.cody.ui;

import java.awt.AlphaComposite;
import java.awt.BasicStroke;
import java.awt.Color;
import java.awt.Dimension;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.geom.Rectangle2D;
import javax.swing.JButton;
import org.jetbrains.annotations.NotNull;

public class TransparentButton extends JButton {
  private final @NotNull Color textColor;
  private final float opacity = 0.7f;
  private final int cornerRadius = 5;
  private final int horizontalPadding = 10;
  private final int verticalPadding = 5;

  public TransparentButton(@NotNull String text, @NotNull Color textColor) {
    super(text);
    this.textColor = textColor;
    setContentAreaFilled(false);
    setFocusPainted(false);
    setBorderPainted(false);

    // Calculate the preferred size based on the size of the text
    FontMetrics fm = getFontMetrics(getFont());
    int width = fm.stringWidth(getText()) + horizontalPadding * 2;
    int height = fm.getHeight() + verticalPadding * 2;
    setPreferredSize(new Dimension(width, height));
  }

  @Override
  protected void paintComponent(Graphics g) {
    Graphics2D g2 = (Graphics2D) g.create();
    g2.setComposite(AlphaComposite.SrcOver.derive(opacity));
    g2.setColor(getBackground());
    g2.fillRoundRect(0, 0, getWidth(), getHeight(), cornerRadius, cornerRadius);

    g2.setColor(getForeground());
    g2.setStroke(new BasicStroke(1));
    g2.drawRoundRect(0, 0, getWidth() - 1, getHeight() - 1, cornerRadius, cornerRadius);
    g2.dispose();

    g.setColor(textColor);
    FontMetrics fm = g.getFontMetrics();
    Rectangle2D rect = fm.getStringBounds(getText(), g);

    int textHeight = (int) rect.getHeight();
    int textWidth = (int) rect.getWidth();
    int panelHeight = this.getHeight();
    int panelWidth = this.getWidth();

    // Center text horizontally and vertically
    int x = (panelWidth - textWidth) / 2;
    int y = (panelHeight - textHeight) / 2 + fm.getAscent();

    g.drawString(getText(), x, y);
  }
}
