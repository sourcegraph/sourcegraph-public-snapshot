package com.sourcegraph.cody.ui;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.ui.ColorUtil;
import com.intellij.ui.components.JBTextArea;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import java.awt.geom.RoundRectangle2D;
import javax.swing.BorderFactory;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.BadLocationException;

public class RoundedJBTextArea extends JBTextArea {

  private final int cornerRadius;
  private final int maxRows;

  public RoundedJBTextArea(int minRows, int maxRows, int cornerRadius) {
    super(minRows, 0);
    this.cornerRadius = cornerRadius;
    this.maxRows = maxRows;
    setOpaque(false);
    setBorder(BorderFactory.createEmptyBorder(4, 4, 4, 4));

    updateRowHeightAndMaxSize();

    getDocument().addDocumentListener(new DocumentListener() {
      @Override
      public void insertUpdate(DocumentEvent e) {
        updateSize();
      }

      @Override
      public void removeUpdate(DocumentEvent e) {
        updateSize();
      }

      @Override
      public void changedUpdate(DocumentEvent e) {
        updateSize();
      }

      private void updateSize() {
        int totalLines = 0;
        try {
          for (int i = 0; i < getLineCount(); i++) {
            int lineStart = getLineStartOffset(i);
            int lineEnd = getLineEndOffset(i);
            String line = getDocument().getText(lineStart, lineEnd - lineStart);
            int lineWidth = getFontMetrics(getFont()).stringWidth(line);
            totalLines += lineWidth / getWidth() + 1;
          }
        } catch (BadLocationException e) {
          e.printStackTrace();
        }
        totalLines = Math.max(Math.min(totalLines, maxRows), minRows);
        setRows(totalLines);

        // Post a runnable to the EDT to refresh the parent after revalidate
        ApplicationManager.getApplication().invokeLater(() -> {
          revalidate();
          Container parent = getParent();
          if (parent != null) {
            parent.repaint();
          }
        });
      }
    });
  }

  @Override
  public void setFont(Font font) {
    super.setFont(font);
    updateRowHeightAndMaxSize();
  }

  private void updateRowHeightAndMaxSize() {
    FontMetrics fm = getFontMetrics(getFont());
    int rowHeight = fm.getHeight();
    setMaximumSize(new Dimension(getMaximumSize().width, rowHeight * maxRows));
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
