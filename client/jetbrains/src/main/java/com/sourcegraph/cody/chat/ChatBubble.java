package com.sourcegraph.cody.chat;

import java.awt.*;
import java.util.List;
import javax.swing.*;
import org.commonmark.ext.gfm.tables.TablesExtension;
import org.commonmark.node.*;
import org.commonmark.parser.Parser;
import org.jetbrains.annotations.NotNull;

public class ChatBubble extends JPanel {
  private static final int GRADIENT_WIDTH = 2;

  public ChatBubble(@NotNull ChatMessage message) {
    super();
    this.setLayout(new BorderLayout());

    JPanel messagePanel = buildMessagePanel(message);
    this.add(messagePanel);
  }

  @NotNull
  private JPanel buildMessagePanel(@NotNull ChatMessage message) {
    MessagePanel messagePanel = new MessagePanel(message, GRADIENT_WIDTH);
    Parser parser = Parser.builder().extensions(List.of(TablesExtension.create())).build();
    Node document = parser.parse(message.getDisplayText());
    MessageContentCreatorFromMarkdownNodes messageContentCreator =
        new MessageContentCreatorFromMarkdownNodes(messagePanel, GRADIENT_WIDTH);
    document.accept(messageContentCreator);
    return messagePanel;
  }

  public void updateText(@NotNull ChatMessage message) {
    JPanel newMessage = buildMessagePanel(message);
    this.remove(0);
    this.add(newMessage, BorderLayout.CENTER, 0);
    this.revalidate();
    this.repaint();
  }
}
