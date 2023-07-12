package com.sourcegraph.cody.ui;

import java.awt.AlphaComposite;
import java.awt.Graphics;
import java.awt.Graphics2D;
import javax.swing.JButton;

public class TransparentButton extends JButton {
  private float transparency = 0.7f;

  public TransparentButton(String text) {
    super(text);
    setOpaque(false);
  }

  @Override
  protected void paintComponent(Graphics g) {
    Graphics2D g2 = (Graphics2D) g.create();
    g2.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, transparency));
    super.paintComponent(g2);
    g2.dispose();
  }
}
