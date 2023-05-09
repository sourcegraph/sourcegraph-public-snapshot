package com.sourcegraph.cody;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.VerticalFlowLayout;
import com.intellij.openapi.wm.ToolWindow;
import com.intellij.ui.components.JBScrollPane;
import com.intellij.ui.scale.JBUIScale;
import com.sourcegraph.cody.chat.Chat;
import com.sourcegraph.cody.chat.ChatBubble;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.completions.Speaker;
import com.sourcegraph.cody.editor.EditorContext;
import com.sourcegraph.cody.editor.EditorContextGetter;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;
import java.awt.event.AdjustmentListener;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;

class CodyToolWindowContent {
    private final @NotNull JPanel contentPanel = new JPanel();
    private final @NotNull JPanel messagesPanel;
    private final @NotNull JTextField messageField;
    private boolean needScrollingDown = true;

    public CodyToolWindowContent(@NotNull Project project, @NotNull ToolWindow toolWindow) {
        // Chat panel
        messagesPanel = new JPanel();
        messagesPanel.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 10, true, true));
        JBScrollPane chatPanel = new JBScrollPane(messagesPanel, JBScrollPane.VERTICAL_SCROLLBAR_AS_NEEDED, JBScrollPane.HORIZONTAL_SCROLLBAR_NEVER);

        // Scroll all the way down after each message
        AdjustmentListener scrollAdjustmentListener = e -> {
            if (needScrollingDown) {
                e.getAdjustable().setValue(e.getAdjustable().getMaximum());
                needScrollingDown = false;
            }
        };
        chatPanel.getVerticalScrollBar().addAdjustmentListener(scrollAdjustmentListener);

        // Controls panel
        JPanel controlsPanel = new JPanel();
        controlsPanel.setLayout(new BoxLayout(controlsPanel, BoxLayout.X_AXIS));
        messageField = new JTextField();
        controlsPanel.add(messageField);
        JButton sendButton = new JButton("Send");
        sendButton.addActionListener(e -> sendMessage(project));
        controlsPanel.add(sendButton);

        // Main content panel
        contentPanel.setLayout(new BorderLayout(0, 20));
        contentPanel.setBorder(BorderFactory.createEmptyBorder(10, 10, 10, 10));
        contentPanel.add(chatPanel, BorderLayout.CENTER);
        contentPanel.add(controlsPanel, BorderLayout.SOUTH);

        // Add welcome message
        var welcomeText = "Hello! I'm Cody. I can write code and answer questions for you. See [Cody documentation](https://docs.sourcegraph.com/cody) for help and tips.";
        addMessage(ChatMessage.createAssistantMessage(welcomeText));
    }

    public void addMessage(@NotNull ChatMessage message) {
        boolean isHuman = message.getSpeaker() == Speaker.HUMAN;

        // Bubble panel
        var bubblePanel = new JPanel();
        bubblePanel.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false));
        bubblePanel.setBorder(BorderFactory.createEmptyBorder(0, isHuman ? JBUIScale.scale(20) : 0, 0, !isHuman ? JBUIScale.scale(20) : 0));

        // Chat bubble
        ChatBubble bubble = new ChatBubble(10, message);
        bubblePanel.add(bubble, VerticalFlowLayout.TOP);
        messagesPanel.add(bubblePanel);
        messagesPanel.revalidate();
        messagesPanel.repaint();

        // Need this hacky solution to scroll all the way down after each message
        SwingUtilities.invokeLater(() -> {
            needScrollingDown = true;
            messagesPanel.revalidate();
            messagesPanel.repaint();
        });
    }

    private void sendMessage(@NotNull Project project) {
        // Build message
        EditorContext editorContext = EditorContextGetter.getEditorContext(project);
        var chat = new Chat();
        ArrayList<String> contextFiles = editorContext == null ? new ArrayList<>() : new ArrayList<>(Collections.singletonList(editorContext.getCurrentFileContent()));
        ChatMessage humanMessage = ChatMessage.createHumanMessage(messageField.getText(), contextFiles);
        addMessage(humanMessage);

        // Get and process assistant message
        ChatMessage assistantMessage;
        try {
            assistantMessage = chat.sendMessage(humanMessage);
        } catch (IOException e) {
            if (e.getMessage().equals("Connection refused")) {
                assistantMessage = ChatMessage.createAssistantMessage("I'm sorry, I can't connect to the server. Please make sure that the server is running and try again.");
            } else {
                assistantMessage = ChatMessage.createAssistantMessage("I'm sorry, I can't connect to the server. Please try again. The error message I got was: \"" + e.getMessage() + "\".");
            }
        } catch (InterruptedException e) {
            assistantMessage = ChatMessage.createAssistantMessage("Okay, I've canceled that.");
        }
        addMessage(assistantMessage);
    }

    public @NotNull JPanel getContentPanel() {
        return contentPanel;
    }
}
