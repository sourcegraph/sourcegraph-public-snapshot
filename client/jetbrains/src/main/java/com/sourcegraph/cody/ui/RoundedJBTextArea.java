package com.sourcegraph.cody.ui;

import com.intellij.ui.ColorUtil;
import com.intellij.ui.components.JBTextArea;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import java.awt.geom.RoundRectangle2D;
import javax.swing.BorderFactory;

public class RoundedJBTextArea extends JBTextArea {

  private final int cornerRadius;

  public RoundedJBTextArea(int minRows, int cornerRadius) {
    super(minRows, 0);
    this.cornerRadius = cornerRadius;
    setOpaque(false);
    setBorder(BorderFactory.createEmptyBorder(4, 4, 4, 4));
  }

  @Override
  protected void paintComponent(Graphics g) {
    Graphics2D g2 = (Graphics2D) g.create();
    g2.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_ON);

    int width = getWidth();
    int height = getHeight();

    RoundRectangle2D.Float roundRect =
        new RoundRectangle2D.Float(0, 0, width - 1, height - 1, cornerRadius, cornerRadius);

    g2.setColor(getBackground());
    g2.fill(roundRect);
    g2.setColor(ColorUtil.brighter(getBackground(), 2));
    g2.draw(roundRect);

    g2.dispose();
    super.paintComponent(g);
  }
}
