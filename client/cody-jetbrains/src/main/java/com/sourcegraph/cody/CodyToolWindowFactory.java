package com.sourcegraph.cody;

import com.intellij.openapi.project.DumbAware;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.wm.ToolWindow;
import com.intellij.openapi.wm.ToolWindowFactory;
import com.intellij.ui.JBColor;
import com.intellij.ui.content.Content;
import com.intellij.ui.content.ContentFactory;
import com.sourcegraph.cody.chat.Chat;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.completions.Speaker;
import com.sourcegraph.cody.editor.EditorContext;
import com.sourcegraph.cody.editor.EditorContextGetter;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
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
        private final @NotNull JPanel chatPanel;
        private final @NotNull JTextField messageField;
//        private final ArrayList<ChatMessage> messages = new ArrayList<>();

        public CodyToolWindowContent(@NotNull Project project, @NotNull ToolWindow toolWindow) {
            // Chat panel
            chatPanel = new JPanel();

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
        }

        public void addMessage(@NotNull ChatMessage message) {
            JTextArea textArea = new JTextArea(message.getText());
            textArea.setBorder(BorderFactory.createEmptyBorder(10, 10, 10, 10));
            textArea.setBackground(message.getSpeaker() == Speaker.HUMAN ? JBColor.LIGHT_GRAY : JBColor.WHITE);
            textArea.setLineWrap(true);
            textArea.setWrapStyleWord(true);
            String layoutConstraint = message.getSpeaker() == Speaker.HUMAN ? BorderLayout.EAST : BorderLayout.WEST;
            chatPanel.add(textArea, layoutConstraint);
            chatPanel.revalidate();
            chatPanel.repaint();
        }

        private void sendMessage(@NotNull Project project) {
            EditorContext editorContext = EditorContextGetter.getEditorContext(project);
            var chat = new Chat();
            ArrayList<String> contextFiles = editorContext == null ? new ArrayList<>() : new ArrayList<>(Collections.singletonList(editorContext.getCurrentFileContent()));

            try {
                var assistantMessage = chat.sendMessage(ChatMessage.createHumanMessage(messageField.getText(), contextFiles));
                addMessage(assistantMessage);
            } catch (IOException e) { // TODO: Add real error handling
                System.out.println("IO Exception");
                e.printStackTrace();
            } catch (InterruptedException e) {
                System.out.println("Interrupted");
                e.printStackTrace();
            }
        }

        public @NotNull JPanel getContentPanel() {
            return contentPanel;
        }
    }
}
