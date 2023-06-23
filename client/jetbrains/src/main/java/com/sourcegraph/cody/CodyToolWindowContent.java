package com.sourcegraph.cody;

import static com.intellij.openapi.util.SystemInfoRt.isMac;
import static java.awt.event.InputEvent.CTRL_DOWN_MASK;
import static java.awt.event.InputEvent.META_DOWN_MASK;
import static java.awt.event.KeyEvent.VK_ENTER;
import static javax.swing.KeyStroke.getKeyStroke;

import com.intellij.ide.ui.laf.darcula.ui.DarculaButtonUI;
import com.intellij.ide.ui.laf.darcula.ui.DarculaTextAreaUI;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.CustomShortcutSet;
import com.intellij.openapi.actionSystem.KeyboardShortcut;
import com.intellij.openapi.actionSystem.ShortcutSet;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.VerticalFlowLayout;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.ui.components.JBScrollPane;
import com.intellij.ui.components.JBTabbedPane;
import com.intellij.ui.components.JBTextArea;
import com.intellij.util.ui.*;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.chat.*;
import com.sourcegraph.cody.context.ContextFile;
import com.sourcegraph.cody.context.ContextGetter;
import com.sourcegraph.cody.context.ContextMessage;
import com.sourcegraph.cody.editor.EditorContext;
import com.sourcegraph.cody.editor.EditorContextGetter;
import com.sourcegraph.cody.prompts.Preamble;
import com.sourcegraph.cody.prompts.Prompter;
import com.sourcegraph.cody.prompts.SupportedLanguages;
import com.sourcegraph.cody.recipes.ExplainCodeDetailedPromptProvider;
import com.sourcegraph.cody.recipes.ExplainCodeHighLevelPromptProvider;
import com.sourcegraph.cody.recipes.FindCodeSmellsPromptProvider;
import com.sourcegraph.cody.recipes.GenerateDocStringPromptProvider;
import com.sourcegraph.cody.recipes.GenerateUnitTestPromptProvider;
import com.sourcegraph.cody.recipes.ImproveVariableNamesPromptProvider;
import com.sourcegraph.cody.recipes.Language;
import com.sourcegraph.cody.recipes.OptimizeCodePromptProvider;
import com.sourcegraph.cody.recipes.PromptProvider;
import com.sourcegraph.cody.recipes.RecipeRunner;
import com.sourcegraph.cody.recipes.SummarizeRecentChangesRecipe;
import com.sourcegraph.cody.recipes.TranslateToLanguagePromptProvider;
import com.sourcegraph.cody.ui.RoundedJBTextArea;
import com.sourcegraph.cody.ui.SelectOptionManager;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.SettingsComponent;
import com.sourcegraph.vcs.RepoUtil;
import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.GridLayout;
import java.awt.event.AdjustmentListener;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Objects;
import java.util.function.Consumer;
import java.util.stream.Collectors;
import javax.swing.BorderFactory;
import javax.swing.BoxLayout;
import javax.swing.JButton;
import javax.swing.JComponent;
import javax.swing.JPanel;
import javax.swing.border.EmptyBorder;
import javax.swing.plaf.ButtonUI;
import javax.swing.plaf.basic.BasicTextAreaUI;
import org.apache.commons.lang3.StringUtils;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

class CodyToolWindowContent implements UpdatableChat {
  public static Logger logger = Logger.getInstance(CodyToolWindowContent.class);
  private static final int CHAT_TAB_INDEX = 0;
  private static final int RECIPES_TAB_INDEX = 1;
  private final @NotNull JBTabbedPane tabbedPane = new JBTabbedPane();
  private final @NotNull JPanel messagesPanel = new JPanel();
  private final @NotNull JBTextArea promptInput;
  private final @NotNull JButton sendButton;
  private final @NotNull Project project;
  private boolean needScrollingDown = true;
  private @NotNull Transcript transcript = new Transcript();

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
        e ->
            executeRecipeWithPromptProvider(recipeRunner, new ExplainCodeDetailedPromptProvider()));
    JButton explainCodeHighLevelButton = createWideButton("Explain selected code (high level)");
    explainCodeHighLevelButton.addActionListener(
        e ->
            executeRecipeWithPromptProvider(
                recipeRunner, new ExplainCodeHighLevelPromptProvider()));
    JButton generateUnitTestButton = createWideButton("Generate a unit test");
    generateUnitTestButton.addActionListener(
        e -> executeRecipeWithPromptProvider(recipeRunner, new GenerateUnitTestPromptProvider()));
    JButton generateDocstringButton = createWideButton("Generate a docstring");
    generateDocstringButton.addActionListener(
        e -> executeRecipeWithPromptProvider(recipeRunner, new GenerateDocStringPromptProvider()));
    JButton improveVariableNamesButton = createWideButton("Improve variable names");
    improveVariableNamesButton.addActionListener(
        e ->
            executeRecipeWithPromptProvider(
                recipeRunner, new ImproveVariableNamesPromptProvider()));
    JButton translateToLanguageButton = createWideButton("Translate to different language");
    translateToLanguageButton.addActionListener(
        e ->
            runIfCodeSelected(
                (editorSelection) -> {
                  SelectOptionManager selectOptionManager =
                      SelectOptionManager.getInstance(project);
                  selectOptionManager.show(
                      project,
                      SupportedLanguages.LANGUAGE_NAMES,
                      (selectedLanguage) ->
                          recipeRunner.runRecipe(
                              new TranslateToLanguagePromptProvider(new Language(selectedLanguage)),
                              editorSelection));
                }));
    JButton gitHistoryButton = createWideButton("Summarize recent code changes");
    gitHistoryButton.addActionListener(
        e ->
            new SummarizeRecentChangesRecipe(project, this, recipeRunner).summarizeRecentChanges());
    JButton findCodeSmellsButton = createWideButton("Smell code");
    findCodeSmellsButton.addActionListener(
        e -> executeRecipeWithPromptProvider(recipeRunner, new FindCodeSmellsPromptProvider()));
    // JButton fixupButton = createWideButton("Fixup code from inline instructions");
    // fixupButton.addActionListener(e -> recipeRunner.runFixup());
    // JButton contextSearchButton = createWideButton("Codebase context search");
    // contextSearchButton.addActionListener(e -> recipeRunner.runContextSearch());
    // JButton releaseNotesButton = createWideButton("Generate release notes");
    // releaseNotesButton.addActionListener(e -> recipeRunner.runReleaseNotes());
    JButton optimizeCodeButton = createWideButton("Optimize code");
    optimizeCodeButton.addActionListener(
        e -> executeRecipeWithPromptProvider(recipeRunner, new OptimizeCodePromptProvider()));
    recipesPanel.add(explainCodeDetailedButton);
    recipesPanel.add(explainCodeHighLevelButton);
    recipesPanel.add(generateUnitTestButton);
    recipesPanel.add(generateDocstringButton);
    recipesPanel.add(improveVariableNamesButton);
    recipesPanel.add(translateToLanguageButton);
    recipesPanel.add(gitHistoryButton);
    recipesPanel.add(findCodeSmellsButton);
    //    recipesPanel.add(fixupButton);
    //    recipesPanel.add(contextSearchButton);
    //    recipesPanel.add(releaseNotesButton);
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

    JBScrollPane promptInputWithScroll =
        new JBScrollPane(
            promptInput,
            JBScrollPane.VERTICAL_SCROLLBAR_AS_NEEDED,
            JBScrollPane.HORIZONTAL_SCROLLBAR_NEVER);
    messagePanel.add(promptInputWithScroll, BorderLayout.CENTER);
    messagePanel.setBorder(BorderFactory.createEmptyBorder(0, 0, 10, 0));

    controlsPanel.add(messagePanel, BorderLayout.NORTH);
    controlsPanel.add(sendButton, BorderLayout.EAST);

    // Main content panel
    contentPanel.setLayout(new BorderLayout(0, 0));
    contentPanel.setBorder(BorderFactory.createEmptyBorder(0, 0, 10, 0));
    contentPanel.add(chatPanel, BorderLayout.CENTER);
    contentPanel.add(controlsPanel, BorderLayout.SOUTH);
    tabbedPane.addChangeListener(e -> this.focusPromptInput());
    // Add welcome message
    addWelcomeMessage();
  }

  private void executeRecipeWithPromptProvider(
      RecipeRunner recipeRunner, PromptProvider promptProvider) {
    runIfCodeSelected((editorSelection) -> recipeRunner.runRecipe(promptProvider, editorSelection));
  }

  private void runIfCodeSelected(@NotNull Consumer<String> runIfCodeSelected) {
    EditorContext editorContext = EditorContextGetter.getEditorContext(project);
    String editorSelection = editorContext.getSelection();
    if (editorSelection == null) {
      this.activateChatTab();
      this.addMessageToChat(
          ChatMessage.createAssistantMessage(
              "No code selected. Please select some code and try again."));
      return;
    }
    runIfCodeSelected.accept(editorSelection);
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
    boolean isEnterprise =
        ConfigUtil.getInstanceType(project).equals(SettingsComponent.InstanceType.ENTERPRISE);
    String accessToken =
        isEnterprise
            ? ConfigUtil.getEnterpriseAccessToken(project)
            : ConfigUtil.getDotComAccessToken(project);
    String welcomeText =
        "Hello! I'm Cody. I can write code and answer questions for you. See [Cody documentation](https://docs.sourcegraph.com/cody) for help and tips.";
    addMessageToChat(ChatMessage.createAssistantMessage(welcomeText));
    if (StringUtils.isEmpty(accessToken)) {
      String noAccessTokenText =
          "<p>It looks like you don't have Sourcegraph Access Token configured.</p>"
              + "<p>See our <a href=\"https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token\">user docs</a> how to create one and configure it in the settings to use Cody.</p>";
      AssistantMessageWithSettingsButton assistantMessageWithSettingsButton =
          new AssistantMessageWithSettingsButton(project, noAccessTokenText);
      var messageContentPanel = new JPanel(new BorderLayout());
      messageContentPanel.add(assistantMessageWithSettingsButton);
      ApplicationManager.getApplication()
          .invokeLater(() -> addComponentToChat(messageContentPanel));
    }
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
    promptInput.setFont(UIUtil.getLabelFont());
    promptInput.setLineWrap(true);
    promptInput.setWrapStyleWord(true);
    promptInput.requestFocusInWindow();
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
              ChatBubble bubble = new ChatBubble(message);
              addComponentToChat(bubble);
            });
  }

  private void addComponentToChat(JPanel message) {

    var bubblePanel = new JPanel();
    bubblePanel.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false));
    // Chat message
    bubblePanel.add(message, VerticalFlowLayout.TOP);
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

  @Override
  public void respondToErrorFromServer(@NotNull String errorMessage) {
    if (errorMessage.equals("Connection refused")) {
      this.addMessageToChat(
          ChatMessage.createAssistantMessage(
              "I'm sorry, I can't connect to the server. Please make sure that the server is running and try again."));
    } else if (errorMessage.startsWith("Got error response 401")) {
      String invalidAccessTokenText =
          "<p>It looks like your Sourcegraph Access Token is invalid or not configured.</p>"
              + "<p>See our <a href=\"https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token\">user docs</a> how to create one and configure it in the settings to use Cody.</p>";
      AssistantMessageWithSettingsButton assistantMessageWithSettingsButton =
          new AssistantMessageWithSettingsButton(project, invalidAccessTokenText);
      var messageContentPanel = new JPanel(new BorderLayout());
      messageContentPanel.add(assistantMessageWithSettingsButton);
      ApplicationManager.getApplication()
          .invokeLater(() -> this.addComponentToChat(messageContentPanel));
    } else {
      this.addMessageToChat(
          ChatMessage.createAssistantMessage(
              "I'm sorry, something wet wrong. Please try again. The error message I got was: \""
                  + errorMessage
                  + "\"."));
    }
  }

  public synchronized void updateLastMessage(@NotNull ChatMessage message) {
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              transcript.addAssistantResponse(message);
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
    transcript = new Transcript();
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              sendButton.setEnabled(true);
              messagesPanel.removeAll();
              addWelcomeMessage();
              messagesPanel.revalidate();
              messagesPanel.repaint();
            });
  }

  private void sendMessage(@NotNull Project project) {
    String messageText = promptInput.getText();
    promptInput.setText("");
    sendMessage(project, ChatMessage.createHumanMessage(messageText, messageText), "");
  }

  private void sendMessage(@NotNull Project project, ChatMessage message, String responsePrefix) {
    if (!sendButton.isEnabled()) {
      return;
    }
    startMessageProcessing();
    // Build message

    EditorContext editorContext = EditorContextGetter.getEditorContext(project);

    String truncatedPrompt =
        TruncationUtils.truncateText(message.prompt(), TruncationUtils.MAX_HUMAN_INPUT_TOKENS);
    ChatMessage humanMessage =
        ChatMessage.createHumanMessage(truncatedPrompt, message.getDisplayText());
    addMessageToChat(humanMessage);
    VirtualFile currentFile = getCurrentFile(project);
    // This cannot run on EDT (Event Dispatch Thread) because it may block for a long time.
    // Also, if we did the back-end call in the main thread and then waited, we wouldn't see the
    // messages streamed back to us.
    ApplicationManager.getApplication()
        .executeOnPooledThread(
            () -> {
              boolean isEnterprise =
                  ConfigUtil.getInstanceType(project)
                      .equals(SettingsComponent.InstanceType.ENTERPRISE);
              String instanceUrl =
                  isEnterprise ? ConfigUtil.getEnterpriseUrl(project) : "https://sourcegraph.com/";
              String accessToken =
                  isEnterprise
                      ? ConfigUtil.getEnterpriseAccessToken(project)
                      : ConfigUtil.getDotComAccessToken(project);

              String repoName = getRepoName(project, currentFile);
              String accessTokenOrEmpty = accessToken != null ? accessToken : "";
              Chat chat = new Chat(instanceUrl, accessTokenOrEmpty);
              if (CodyAgent.isConnected(project)) {
                try {
                  chat.sendMessageViaAgent(
                      CodyAgent.getClient(project),
                      CodyAgent.getInitializedServer(project),
                      humanMessage,
                      this);
                } catch (Exception e) {
                  logger.error("Error sending message '" + humanMessage + "' to chat", e);
                }
              } else {
                List<ContextMessage> contextMessages =
                    getContextFromEmbeddings(
                        project, humanMessage, instanceUrl, repoName, accessTokenOrEmpty);
                displayUsedFiles(contextMessages);
                List<ContextMessage> editorContextMessages =
                    getEditorContextMessages(editorContext);
                contextMessages.addAll(editorContextMessages);
                List<ContextMessage> selectionContextMessages =
                    getSelectionContextMessages(editorContext);
                contextMessages.addAll(selectionContextMessages);
                // Add human message
                transcript.addInteraction(new Interaction(humanMessage, contextMessages));

                List<Message> prompt =
                    transcript.getPromptForLastInteraction(
                        Preamble.getPreamble(repoName),
                        TruncationUtils.MAX_AVAILABLE_PROMPT_LENGTH);

                try {
                  chat.sendMessageWithoutAgent(prompt, responsePrefix, this);
                } catch (Exception e) {
                  logger.error("Error sending message '" + humanMessage + "' to chat", e);
                }
              }
            });
  }

  private void displayUsedFiles(List<ContextMessage> contextMessages) {
    // Use context
    if (contextMessages.size() == 0) {
      this.addMessageToChat(
          ChatMessage.createAssistantMessage(
              "I didn't find any context for your ask. I'll try to answer without further context."));
    } else {
      // Collect file names and display them to user
      List<String> contextFileNames =
          contextMessages.stream()
              .map(ContextMessage::getFile)
              .filter(Objects::nonNull)
              .map(ContextFile::getFileName)
              .collect(Collectors.toList());

      StringBuilder contextMessageText =
          new StringBuilder(
              "I found some context for your ask. I'll try to answer with the context of these "
                  + contextFileNames.size()
                  + " files:\n");
      contextFileNames.forEach(fileName -> contextMessageText.append(fileName).append("\n"));
      this.addMessageToChat(ChatMessage.createAssistantMessage(contextMessageText.toString()));
    }
  }

  @NotNull
  private List<ContextMessage> getContextFromEmbeddings(
      @NotNull Project project,
      ChatMessage humanMessage,
      String instanceUrl,
      String repoName,
      String accessTokenOrEmpty) {
    List<ContextMessage> contextMessages = new ArrayList<>();
    if (repoName != null) {
      try {
        contextMessages =
            new ContextGetter(
                    repoName,
                    instanceUrl,
                    accessTokenOrEmpty,
                    ConfigUtil.getCustomRequestHeaders(project))
                .getContextMessages(humanMessage.getText(), 8, 2, true);
      } catch (IOException e) {
        this.addMessageToChat(
            ChatMessage.createAssistantMessage(
                "I didn't get a correct response. This is what I encountered while trying to get some context for your ask: \""
                    + e.getMessage()
                    + "\". I'll try to answer without further context."));
      }
    }
    return contextMessages;
  }

  private List<ContextMessage> getEditorContextMessages(EditorContext editorContext) {
    if (editorContext.getCurrentFileName() != null
        && editorContext.getCurrentFileContent() != null) {
      String truncatedCurrentFileContent =
          TruncationUtils.truncateText(
              editorContext.getCurrentFileContent(), TruncationUtils.MAX_CURRENT_FILE_TOKENS);
      String currentFilePrompt =
          Prompter.getCurrentEditorCodePrompt(
              editorContext.getCurrentFileName(), truncatedCurrentFileContent);
      ContextMessage currentFileHumanMessage = ContextMessage.createHumanMessage(currentFilePrompt);
      ContextMessage defaultAssistantMessage = ContextMessage.createDefaultAssistantMessage();
      return List.of(currentFileHumanMessage, defaultAssistantMessage);
    }
    return Collections.emptyList();
  }

  public List<ContextMessage> getSelectionContextMessages(EditorContext editorContext) {
    if (editorContext.getCurrentFileName() != null && editorContext.getSelection() != null) {
      String selectedText = editorContext.getSelection();
      String truncatedSelectedText =
          TruncationUtils.truncateText(selectedText, TruncationUtils.MAX_CURRENT_FILE_TOKENS);
      String selectedTextPrompt =
          Prompter.getCurrentEditorSelectedCode(
              editorContext.getCurrentFileName(), truncatedSelectedText);
      ContextMessage selectedTextHumanMessage =
          ContextMessage.createHumanMessage(selectedTextPrompt);
      ContextMessage defaultAssistantMessage = ContextMessage.createDefaultAssistantMessage();
      return List.of(selectedTextHumanMessage, defaultAssistantMessage);
    }
    return Collections.emptyList();
  }

  @Nullable
  private static VirtualFile getCurrentFile(@NotNull Project project) {
    VirtualFile currentFile = null;
    Editor editor = FileEditorManager.getInstance(project).getSelectedTextEditor();
    if (editor != null) {
      Document currentDocument = editor.getDocument();
      currentFile = FileDocumentManager.getInstance().getFile(currentDocument);
    }
    return currentFile;
  }

  @Nullable
  private static String getRepoName(@NotNull Project project, @Nullable VirtualFile currentFile) {
    if (currentFile == null) {
      return null;
    }
    try {
      return RepoUtil.getRemoteRepoUrlWithoutScheme(project, currentFile);
    } catch (Exception e) {
      return null;
    }
  }

  public @NotNull JComponent getContentPanel() {
    return tabbedPane;
  }

  public void focusPromptInput() {
    if (tabbedPane.getSelectedIndex() == CHAT_TAB_INDEX) {
      promptInput.requestFocusInWindow();
      int textLength = promptInput.getDocument().getLength();
      promptInput.setCaretPosition(textLength);
    }
  }
}
