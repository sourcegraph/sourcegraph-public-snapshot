package com.sourcegraph.cody.chat;

import com.intellij.ui.JBColor;
import com.intellij.util.ui.JBInsets;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.cody.completions.Speaker;
import java.awt.*;
import javax.swing.*;
import javax.swing.border.EmptyBorder;
import org.commonmark.node.Node;
import org.commonmark.parser.Parser;
import org.commonmark.renderer.html.HtmlRenderer;
import org.jetbrains.annotations.NotNull;

public class ChatBubble extends JPanel {
  private final int radius;

  public ChatBubble(int radius, @NotNull ChatMessage message) {
    super();

    boolean isHuman = message.getSpeaker() == Speaker.HUMAN;
    JBColor background = isHuman ? JBColor.BLUE : JBColor.GRAY;
    this.radius = radius;
    this.setBackground(background);
    this.setLayout(new BorderLayout());
    this.setBorder(new EmptyBorder(new JBInsets(10, 10, 10, 10)));

    JEditorPane pane = new JEditorPane();
    pane.setContentType("text/html");
    pane.setText(convertToHtml(message.getDisplayText()));
    pane.setEditable(false);
    pane.setFont(UIUtil.getLabelFont());
    pane.setBackground(background);
    pane.setForeground(isHuman ? JBColor.WHITE : JBColor.BLACK);
    pane.setComponentOrientation(
        isHuman ? ComponentOrientation.RIGHT_TO_LEFT : ComponentOrientation.LEFT_TO_RIGHT);
    pane.setBorder(BorderFactory.createEmptyBorder(0, 0, 0, 0));
    this.add(pane, BorderLayout.CENTER);
  }

  public void updateText(@NotNull String newMarkdownText) {
    JEditorPane pane = (JEditorPane) this.getComponent(0);
    pane.setText(convertToHtml(newMarkdownText));
  }

  private @NotNull String convertToHtml(@NotNull String markdown) {
    // Parse markdown
    Parser parser = Parser.builder().build();
    Node node = parser.parse(markdown);
    HtmlRenderer renderer = HtmlRenderer.builder().build();
    String messageAsHtml = renderer.render(node);

    // Build HTML
    return "<html data-gramm=\"false\"><head><style>p { margin:0; }</style></head><body>"
        + messageAsHtml
        + "</body></html>";
  }

  @Override
  protected void paintComponent(@NotNull Graphics g) {
    final Graphics2D g2d = (Graphics2D) g;
    g2d.setColor(getBackground());
    g2d.fillRoundRect(0, 0, this.getWidth() - 1, this.getHeight() - 1, this.radius, this.radius);
  }
}
