package com.sourcegraph.cody;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.VerticalFlowLayout;
import com.intellij.ui.components.JBScrollPane;
import com.intellij.ui.components.JBTabbedPane;
import com.intellij.ui.scale.JBUIScale;
import com.sourcegraph.cody.chat.Chat;
import com.sourcegraph.cody.chat.ChatBubble;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.completions.Speaker;
import com.sourcegraph.cody.config.ConfigUtil;
import com.sourcegraph.cody.config.SettingsComponent;
import com.sourcegraph.cody.editor.EditorContextGetter;
import com.sourcegraph.cody.recipes.RecipeRunner;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;
import java.awt.event.AdjustmentListener;
import java.util.ArrayList;

class CodyToolWindowContent implements UpdatableChat {
    private final @NotNull JBTabbedPane tabbedPane = new JBTabbedPane();
    private final @NotNull JPanel contentPanel = new JPanel();
    private final @NotNull JPanel recipesPanel = new JPanel();
    private final @NotNull JPanel messagesPanel = new JPanel();
    private final @NotNull JTextField messageField;
    private boolean needScrollingDown = true;

    public CodyToolWindowContent(@NotNull Project project) {
        // Tabs
        tabbedPane.insertTab("Chat", null, contentPanel, null, 0);
        tabbedPane.insertTab("Recipes", null, recipesPanel, null, 1);

        // Recipes panel
        RecipeRunner recipeRunner = new RecipeRunner(project, this);
        JButton explainCodeDetailedButton = new JButton("Explain selected code (detailed)");
        explainCodeDetailedButton.addActionListener(e -> recipeRunner.runExplainCodeDetailed());
        JButton explainCodeHighLevelButton = new JButton("Explain selected code (high level)");
        explainCodeHighLevelButton.addActionListener(e -> recipeRunner.runExplainCodeHighLevel());
        JButton generateUnitTestButton = new JButton("Generate a unit test");
        generateUnitTestButton.addActionListener(e -> recipeRunner.runGenerateUnitTest());
        JButton generateDocstringButton = new JButton("Generate a docstring");
        generateDocstringButton.addActionListener(e -> recipeRunner.runGenerateDocstring());
        JButton improveVariableNamesButton = new JButton("Improve variable names");
        improveVariableNamesButton.addActionListener(e -> recipeRunner.runImproveVariableNames());
        JButton translateToLanguageButton = new JButton("Translate to different language");
        translateToLanguageButton.addActionListener(e -> recipeRunner.runTranslateToLanguage());
        JButton gitHistoryButton = new JButton("Summarize recent code changes");
        gitHistoryButton.addActionListener(e -> recipeRunner.runGitHistory());
        JButton findCodeSmellsButton = new JButton("Smell code");
        findCodeSmellsButton.addActionListener(e -> recipeRunner.runFindCodeSmells());
        JButton fixupButton = new JButton("Fixup code from inline instructions");
        fixupButton.addActionListener(e -> recipeRunner.runFixup());
        JButton contextSearchButton = new JButton("Codebase context search");
        contextSearchButton.addActionListener(e -> recipeRunner.runContextSearch());
        JButton releaseNotesButton = new JButton("Generate release notes");
        releaseNotesButton.addActionListener(e -> recipeRunner.runReleaseNotes());
        recipesPanel.add(explainCodeDetailedButton);
        recipesPanel.add(explainCodeHighLevelButton);
        recipesPanel.add(generateUnitTestButton);
        recipesPanel.add(generateDocstringButton);
        recipesPanel.add(improveVariableNamesButton);
        recipesPanel.add(translateToLanguageButton);
        recipesPanel.add(gitHistoryButton);
        recipesPanel.add(findCodeSmellsButton);
        recipesPanel.add(fixupButton);
        recipesPanel.add(contextSearchButton);
        recipesPanel.add(releaseNotesButton);

        // Chat panel
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
        messageField.addActionListener(e -> sendMessage(project)); // TODO: Disable the button while sending, then re-enable it
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

    public synchronized void addMessage(@NotNull ChatMessage message) {
        ApplicationManager.getApplication().invokeLater(() -> {
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
            ApplicationManager.getApplication().invokeLater(() -> {
                needScrollingDown = true;
                messagesPanel.revalidate();
                messagesPanel.repaint();
            });
        });
    }

    public synchronized void updateLastMessage(@NotNull ChatMessage message) {
        ApplicationManager.getApplication().invokeLater(() -> {
            if (messagesPanel.getComponentCount() > 0) {
                JPanel lastBubblePanel = (JPanel) messagesPanel.getComponent(messagesPanel.getComponentCount() - 1);
                ChatBubble lastBubble = (ChatBubble) lastBubblePanel.getComponent(0);
                lastBubble.updateText(message.getDisplayText());
                messagesPanel.revalidate();
                messagesPanel.repaint();
            }
        });
    }

    private void sendMessage(@NotNull Project project) {
        // Build message
        boolean isEnterprise = ConfigUtil.getInstanceType(project).equals(SettingsComponent.InstanceType.ENTERPRISE);
        String instanceUrl = isEnterprise ? ConfigUtil.getEnterpriseUrl(project) : "https://sourcegraph.com/";
        String accessToken = isEnterprise ? ConfigUtil.getEnterpriseAccessToken(project) : ConfigUtil.getDotcomAccessToken(project);

        var chat = new Chat("", instanceUrl, accessToken != null ? accessToken : "");
        ArrayList<String> contextFiles = EditorContextGetter.getEditorContext(project).getCurrentFileContentAsArrayList();
        ChatMessage humanMessage = ChatMessage.createHumanMessage(messageField.getText(), contextFiles);
        addMessage(humanMessage);

        // Get assistant message
        // Note: A separate thread is needed because it's a long-running task. If we did the back-end call
        //       in the main thread and then waited, we wouldn't see the messages streamed back to us.
        new Thread(() -> {
            chat.sendMessage(humanMessage, "", this); // TODO: Use prefix
        }).start();
    }

    public @NotNull JComponent getContentPanel() {
        return tabbedPane;
    }
}

