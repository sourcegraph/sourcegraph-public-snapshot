package com.sourcegraph.cody;

import com.intellij.openapi.project.DumbAware;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.VerticalFlowLayout;
import com.intellij.openapi.wm.ToolWindow;
import com.intellij.openapi.wm.ToolWindowFactory;
import com.intellij.ui.JBColor;
import com.intellij.ui.components.JBScrollPane;
import com.intellij.ui.content.Content;
import com.intellij.ui.content.ContentFactory;
import com.intellij.ui.scale.JBUIScale;
import com.intellij.uiDesigner.core.Spacer;
import com.intellij.util.ui.JBInsets;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.cody.chat.Chat;
import com.sourcegraph.cody.chat.ChatBubble;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.completions.Speaker;
import com.sourcegraph.cody.editor.EditorContext;
import com.sourcegraph.cody.editor.EditorContextGetter;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import javax.swing.border.EmptyBorder;
import java.awt.*;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;

public class CodyToolWindowFactory implements ToolWindowFactory, DumbAware {
    @Override
    public boolean isApplicable(@NotNull Project project) {
        return ToolWindowFactory.super.isApplicable(project);
    }

    @Override
    public void createToolWindowContent(@NotNull Project project, @NotNull ToolWindow toolWindow) {
        CodyToolWindowContent toolWindowContent = new CodyToolWindowContent(project, toolWindow);
        Content content = ContentFactory.SERVICE.getInstance().createContent(toolWindowContent.getContentPanel(), "", false);
        toolWindow.getContentManager().addContent(content);
    }

    private static class CodyToolWindowContent {
        private final @NotNull JPanel contentPanel = new JPanel();
        private final @NotNull JPanel messagesPanel;
        private final @NotNull JTextField messageField;
//        private final ArrayList<ChatMessage> messages = new ArrayList<>();

        public CodyToolWindowContent(@NotNull Project project, @NotNull ToolWindow toolWindow) {
            // Chat panel
            JBScrollPane chatPanel = new JBScrollPane();
            chatPanel.setHorizontalScrollBarPolicy(ScrollPaneConstants.HORIZONTAL_SCROLLBAR_NEVER);
            chatPanel.setLayout(new ScrollPaneLayout());

            messagesPanel = new JPanel();
            messagesPanel.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 10, true, true));

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
            contentPanel.add(chatPanel, BorderLayout.NORTH);
            contentPanel.add(controlsPanel, BorderLayout.SOUTH);

            // Add welcome message
            var welcomeText = "Hello! I'm Cody. I can write code and answer questions for you. See [Cody documentation](https://docs.sourcegraph.com/cody) for help and tips.";
            addMessage(ChatMessage.createAssistantMessage(welcomeText));
            addMessage(ChatMessage.createHumanMessage(welcomeText, new ArrayList<>()));
            addMessage(ChatMessage.createAssistantMessage(welcomeText));
        }

        public void addMessage(@NotNull ChatMessage message) {
            boolean isHuman = message.getSpeaker() == Speaker.HUMAN;

            // Bubble panel
            var bubblePanel = new JPanel();
            bubblePanel.setLayout(new GridBagLayout());
            messagesPanel.add(bubblePanel);

            GridBagConstraints c = new GridBagConstraints();

            // Spacer
            Spacer spacer = new Spacer();
            spacer.setMinimumSize(new Dimension(0, 400));
            spacer.setMaximumSize(new Dimension(bubblePanel.getWidth() - JBUIScale.scale(400), 400));
            c.weightx = 0.2;
            c.anchor = isHuman ? FlowLayout.LEFT : FlowLayout.RIGHT;
            bubblePanel.add(spacer, c);

            // Chat bubble
            JBColor backgroundColor = isHuman ? JBColor.BLUE : JBColor.GRAY;
            ChatBubble bubble = new ChatBubble();
            System.out.println("bubblePanel sizes:");
            System.out.println("  width: " + bubblePanel.getWidth());
            System.out.println("  height: " + bubblePanel.getHeight());
            bubble.setBorder(new EmptyBorder(new JBInsets(10, 10, 10, 10)));
            JTextArea bubbleTextArea = new JTextArea(message.getDisplayText());
            bubbleTextArea.setFont(UIUtil.getLabelFont());
            bubbleTextArea.setLineWrap(true);
            bubbleTextArea.setWrapStyleWord(true);
            bubbleTextArea.setBackground(JBColor.RED);
            bubbleTextArea.setComponentOrientation(isHuman ? ComponentOrientation.RIGHT_TO_LEFT : ComponentOrientation.LEFT_TO_RIGHT);
            bubble.add(bubbleTextArea, BorderLayout.NORTH);
            System.out.println("bubble sizes:");
            System.out.println("  width: " + bubble.getWidth());
            System.out.println("  height: " + bubble.getHeight());
            c.weightx = 0.8;
            c.anchor = isHuman ? FlowLayout.RIGHT : FlowLayout.LEFT;
            bubblePanel.add(bubble, c);

            // Assemble
            messagesPanel.revalidate();
            messagesPanel.repaint();
        }

        private void sendMessage(@NotNull Project project) {
            EditorContext editorContext = EditorContextGetter.getEditorContext(project);
            var chat = new Chat();
            ArrayList<String> contextFiles = editorContext == null ? new ArrayList<>() : new ArrayList<>(Collections.singletonList(editorContext.getCurrentFileContent()));
            ChatMessage humanMessage = ChatMessage.createHumanMessage(messageField.getText(), contextFiles);
            addMessage(humanMessage);
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
}
