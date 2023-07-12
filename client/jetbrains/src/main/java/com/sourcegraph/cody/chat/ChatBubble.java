package com.sourcegraph.cody.chat;

import java.awt.*;
import java.util.List;
import javax.swing.*;
import org.commonmark.ext.gfm.tables.TablesExtension;
import org.commonmark.node.*;
import org.commonmark.parser.Parser;
import org.jetbrains.annotations.NotNull;

public class ChatBubble extends JPanel {

  public ChatBubble(@NotNull ChatMessage message, @NotNull JPanel parentPanel) {
    super();
    this.setLayout(new BorderLayout());

    JPanel messagePanel = buildMessagePanel(message, parentPanel);
    this.add(messagePanel);
  }

  @NotNull
  private JPanel buildMessagePanel(@NotNull ChatMessage message, @NotNull JPanel parentPanel) {
    /* Create panel */
    MessagePanel messagePanel =
        new MessagePanel(message.getSpeaker(), ChatUIConstants.ASSISTANT_MESSAGE_GRADIENT_WIDTH);

    /* Convert markdown-formatted chat message to Swing components */
    Parser parser = Parser.builder().extensions(List.of(TablesExtension.create())).build();
    Node document = parser.parse(message.getDisplayText());
    MessageContentCreatorFromMarkdownNodes messageContentCreator =
        new MessageContentCreatorFromMarkdownNodes(
            messagePanel,
            parentPanel,
            message.getSpeaker(),
            ChatUIConstants.ASSISTANT_MESSAGE_GRADIENT_WIDTH);
    document.accept(messageContentCreator);

    return messagePanel;
  }

  /**
   * This is useful when receiving the streamed response. In the background, it removes the last
   * message and adds the updated one.
   */
  public void updateText(@NotNull ChatMessage message, @NotNull JPanel parentPanel) {
    JPanel newMessage = buildMessagePanel(message, parentPanel);
    this.remove(0);
    this.add(newMessage, BorderLayout.CENTER, 0);
    this.revalidate();
    this.repaint();
  }
}
