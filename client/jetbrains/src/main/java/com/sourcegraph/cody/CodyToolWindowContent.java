package com.sourcegraph.cody;

import static com.intellij.openapi.util.SystemInfoRt.isMac;
import static java.awt.event.InputEvent.CTRL_DOWN_MASK;
import static java.awt.event.InputEvent.META_DOWN_MASK;
import static java.awt.event.KeyEvent.VK_ENTER;
import static javax.swing.KeyStroke.getKeyStroke;

import com.intellij.ide.ui.laf.darcula.ui.DarculaButtonUI;
import com.intellij.ide.ui.laf.darcula.ui.DarculaTextAreaUI;
import com.intellij.openapi.actionSystem.*;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.VerticalFlowLayout;
import com.intellij.ui.components.JBScrollPane;
import com.intellij.ui.components.JBTabbedPane;
import com.intellij.ui.components.JBTextArea;
import com.intellij.util.ui.JBUI;
import com.sourcegraph.cody.chat.Chat;
import com.sourcegraph.cody.chat.ChatBubble;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.editor.EditorContextGetter;
import com.sourcegraph.cody.recipes.*;
import com.sourcegraph.cody.ui.RoundedJBTextArea;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.SettingsComponent;
import java.awt.*;
import java.awt.event.AdjustmentListener;
import java.util.ArrayList;
import java.util.Collections;
import javax.swing.*;
import javax.swing.border.EmptyBorder;
import javax.swing.plaf.ButtonUI;
import javax.swing.plaf.basic.BasicTextAreaUI;
import org.jetbrains.annotations.NotNull;

class CodyToolWindowContent implements UpdatableChat {
  private static final int CHAT_TAB_INDEX = 0;
  private static final int RECIPES_TAB_INDEX = 1;
  private final @NotNull JBTabbedPane tabbedPane = new JBTabbedPane();
  private final @NotNull JPanel messagesPanel = new JPanel();
  private final @NotNull JBTextArea promptInput;
  private final @NotNull JButton sendButton;
  private final @NotNull Project project;
  private boolean needScrollingDown = true;

  public CodyToolWindowContent(@NotNull Project project) {
    this.project = project;
    // Tabs
    @NotNull JPanel contentPanel = new JPanel();
    tabbedPane.insertTab("Chat", null, contentPanel, null, CHAT_TAB_INDEX);
    @NotNull JPanel recipesPanel = new JPanel(new GridLayout(0, 1));
    recipesPanel.setLayout(new BoxLayout(recipesPanel, BoxLayout.Y_AXIS));
    tabbedPane.insertTab("Recipes", null, recipesPanel, null, RECIPES_TAB_INDEX);

    // Recipes panel
    RecipeRunner recipeRunner = new RecipeRunner(this.project, this);
    JButton explainCodeDetailedButton = createWideButton("Explain selected code (detailed)");
    explainCodeDetailedButton.addActionListener(
        e -> recipeRunner.runRecipe(new ExplainCodeDetailedPromptProvider()));
    JButton explainCodeHighLevelButton = createWideButton("Explain selected code (high level)");
    explainCodeHighLevelButton.addActionListener(
        e -> recipeRunner.runRecipe(new ExplainCodeHighLevelPromptProvider()));
    JButton generateUnitTestButton = createWideButton("Generate a unit test");
    generateUnitTestButton.addActionListener(
        e -> recipeRunner.runRecipe(new GenerateUnitTestPromptProvider()));
    JButton generateDocstringButton = createWideButton("Generate a docstring");
    generateDocstringButton.addActionListener(
        e -> recipeRunner.runRecipe(new GenerateDocStringPromptProvider()));
    JButton improveVariableNamesButton = createWideButton("Improve variable names");
    improveVariableNamesButton.addActionListener(
        e -> recipeRunner.runRecipe(new ImproveVariableNamesPromptProvider()));
    JButton translateToLanguageButton = createWideButton("Translate to different language");
    translateToLanguageButton.addActionListener(e -> recipeRunner.runTranslateToLanguage());
    JButton gitHistoryButton = createWideButton("Summarize recent code changes");
    gitHistoryButton.addActionListener(e -> recipeRunner.runGitHistory());
    JButton findCodeSmellsButton = createWideButton("Smell code");
    findCodeSmellsButton.addActionListener(
        e -> recipeRunner.runRecipe(new FindCodeSmellsPromptProvider()));
    JButton fixupButton = createWideButton("Fixup code from inline instructions");
    fixupButton.addActionListener(e -> recipeRunner.runFixup());
    JButton contextSearchButton = createWideButton("Codebase context search");
    contextSearchButton.addActionListener(e -> recipeRunner.runContextSearch());
    JButton releaseNotesButton = createWideButton("Generate release notes");
    releaseNotesButton.addActionListener(e -> recipeRunner.runReleaseNotes());
    JButton optimizeCodeButton = createWideButton("Optimize code");
    optimizeCodeButton.addActionListener(
        e -> recipeRunner.runRecipe(new OptimizeCodePromptProvider()));
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
    recipesPanel.add(optimizeCodeButton);

    // Chat panel
    messagesPanel.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, true));
    JBScrollPane chatPanel =
        new JBScrollPane(
            messagesPanel,
            JBScrollPane.VERTICAL_SCROLLBAR_AS_NEEDED,
            JBScrollPane.HORIZONTAL_SCROLLBAR_NEVER);

    // Scroll all the way down after each message
    AdjustmentListener scrollAdjustmentListener =
        e -> {
          if (needScrollingDown) {
            e.getAdjustable().setValue(e.getAdjustable().getMaximum());
            needScrollingDown = false;
          }
        };
    chatPanel.getVerticalScrollBar().addAdjustmentListener(scrollAdjustmentListener);

    // Controls panel
    JPanel controlsPanel = new JPanel();
    controlsPanel.setLayout(new BorderLayout());
    controlsPanel.setBorder(new EmptyBorder(JBUI.insets(14)));
    sendButton = createSendButton(this.project);
    promptInput = createPromptInput(this.project);

    JPanel messagePanel = new JPanel(new BorderLayout());
    messagePanel.add(promptInput, BorderLayout.CENTER);
    messagePanel.setBorder(BorderFactory.createEmptyBorder(0, 0, 10, 0));

    controlsPanel.add(messagePanel, BorderLayout.NORTH);
    controlsPanel.add(sendButton, BorderLayout.EAST);

    // Main content panel
    contentPanel.setLayout(new BorderLayout(0, 0));
    contentPanel.setBorder(BorderFactory.createEmptyBorder(0, 0, 10, 0));
    contentPanel.add(chatPanel, BorderLayout.CENTER);
    contentPanel.add(controlsPanel, BorderLayout.SOUTH);

    // Add welcome message
    addWelcomeMessage();
  }

  @NotNull
  private static JButton createWideButton(@NotNull String text) {
    JButton button = new JButton(text);
    button.setAlignmentX(Component.CENTER_ALIGNMENT);
    button.setMaximumSize(new Dimension(Integer.MAX_VALUE, button.getPreferredSize().height));
    ButtonUI buttonUI = (ButtonUI) DarculaButtonUI.createUI(button);
    button.setUI(buttonUI);
    return button;
  }

  private void addWelcomeMessage() {
    var welcomeText =
        "Hello! I'm Cody. I can write code and answer questions for you. See [Cody documentation](https://docs.sourcegraph.com/cody) for help and tips.";
    addMessageToChat(ChatMessage.createAssistantMessage(welcomeText));
  }

  @NotNull
  private JButton createSendButton(@NotNull Project project) {
    JButton sendButton = new JButton("Send");
    sendButton.putClientProperty(DarculaButtonUI.DEFAULT_STYLE_KEY, Boolean.TRUE);
    ButtonUI buttonUI = (ButtonUI) DarculaButtonUI.createUI(sendButton);
    sendButton.setUI(buttonUI);
    sendButton.addActionListener(e -> sendMessage(project));
    return sendButton;
  }

  @NotNull
  private JBTextArea createPromptInput(@NotNull Project project) {
    JBTextArea promptInput = new RoundedJBTextArea(4, 0, 10);
    BasicTextAreaUI textUI = (BasicTextAreaUI) DarculaTextAreaUI.createUI(promptInput);
    promptInput.setUI(textUI);
    promptInput.setLineWrap(true);
    KeyboardShortcut CTRL_ENTER =
        new KeyboardShortcut(getKeyStroke(VK_ENTER, CTRL_DOWN_MASK), null);
    KeyboardShortcut META_ENTER =
        new KeyboardShortcut(getKeyStroke(VK_ENTER, META_DOWN_MASK), null);
    ShortcutSet DEFAULT_SUBMIT_ACTION_SHORTCUT =
        isMac ? new CustomShortcutSet(CTRL_ENTER, META_ENTER) : new CustomShortcutSet(CTRL_ENTER);
    AnAction sendMessageAction =
        new DumbAwareAction() {
          @Override
          public void actionPerformed(@NotNull AnActionEvent e) {
            sendMessage(project);
          }
        };
    sendMessageAction.registerCustomShortcutSet(DEFAULT_SUBMIT_ACTION_SHORTCUT, promptInput);
    return promptInput;
  }

  public synchronized void addMessageToChat(@NotNull ChatMessage message) {
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              // Bubble panel
              var bubblePanel = new JPanel();
              bubblePanel.setLayout(
                  new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false));
              // Chat bubble
              ChatBubble bubble = new ChatBubble(message);
              bubblePanel.add(bubble, VerticalFlowLayout.TOP);
              messagesPanel.add(bubblePanel);
              messagesPanel.revalidate();
              messagesPanel.repaint();

              // Need this hacky solution to scroll all the way down after each message
              ApplicationManager.getApplication()
                  .invokeLater(
                      () -> {
                        needScrollingDown = true;
                        messagesPanel.revalidate();
                        messagesPanel.repaint();
                      });
            });
  }

  @Override
  public void activateChatTab() {
    this.tabbedPane.setSelectedIndex(CHAT_TAB_INDEX);
  }

  @Override
  public void respondToMessage(@NotNull ChatMessage message, @NotNull String responsePrefix) {
    activateChatTab();
    sendMessage(this.project, message, responsePrefix);
  }

  public synchronized void updateLastMessage(@NotNull ChatMessage message) {
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              if (messagesPanel.getComponentCount() > 0) {
                JPanel lastBubblePanel =
                    (JPanel) messagesPanel.getComponent(messagesPanel.getComponentCount() - 1);
                ChatBubble lastBubble = (ChatBubble) lastBubblePanel.getComponent(0);
                lastBubble.updateText(message);
                messagesPanel.revalidate();
                messagesPanel.repaint();
              }
            });
  }

  private void startMessageProcessing() {
    ApplicationManager.getApplication().invokeLater(() -> sendButton.setEnabled(false));
  }

  @Override
  public void finishMessageProcessing() {
    ApplicationManager.getApplication().invokeLater(() -> sendButton.setEnabled(true));
  }

  @Override
  public void resetConversation() {
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              messagesPanel.removeAll();
              addWelcomeMessage();
              messagesPanel.revalidate();
              messagesPanel.repaint();
            });
  }

  private void sendMessage(@NotNull Project project) {
    String messageText = promptInput.getText();
    sendMessage(
        project,
        ChatMessage.createHumanMessage(messageText, messageText, Collections.emptyList()),
        "");
  }

  private void sendMessage(@NotNull Project project, ChatMessage message, String responsePrefix) {
    if (!sendButton.isEnabled()) return;
    startMessageProcessing();
    // Build message
    boolean isEnterprise =
        ConfigUtil.getInstanceType(project).equals(SettingsComponent.InstanceType.ENTERPRISE);
    String instanceUrl =
        isEnterprise ? ConfigUtil.getEnterpriseUrl(project) : "https://sourcegraph.com/";
    String accessToken =
        isEnterprise
            ? ConfigUtil.getEnterpriseAccessToken(project)
            : ConfigUtil.getDotComAccessToken(project);
    System.out.println("isEnterprise: " + isEnterprise);

    var chat = new Chat("", instanceUrl, accessToken != null ? accessToken : "");
    ArrayList<String> contextFiles =
        EditorContextGetter.getEditorContext(project).getCurrentFileContentAsArrayList();
    ChatMessage humanMessage =
        ChatMessage.createHumanMessage(message.prompt(), message.getDisplayText(), contextFiles);
    addMessageToChat(humanMessage);

    // Get assistant message
    // Note: A separate thread is needed because it's a long-running task. If we did the back-end
    // call
    //       in the main thread and then waited, we wouldn't see the messages streamed back to us.
    new Thread(
            () -> {
              chat.sendMessage(humanMessage, responsePrefix, this); // TODO: Use prefix
            })
        .start();
  }

  public @NotNull JComponent getContentPanel() {
    return tabbedPane;
  }
}
