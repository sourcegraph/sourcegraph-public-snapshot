package com.sourcegraph.cody;

import static com.sourcegraph.cody.chat.ChatUIConstants.TEXT_MARGIN;
import static java.awt.event.InputEvent.ALT_DOWN_MASK;
import static java.awt.event.InputEvent.CTRL_DOWN_MASK;
import static java.awt.event.InputEvent.META_DOWN_MASK;
import static java.awt.event.InputEvent.SHIFT_DOWN_MASK;
import static java.awt.event.KeyEvent.VK_ENTER;
import static javax.swing.KeyStroke.getKeyStroke;

import com.intellij.ide.BrowserUtil;
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
import com.intellij.openapi.options.ShowSettingsUtil;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.VerticalFlowLayout;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.ui.components.JBScrollPane;
import com.intellij.ui.components.JBTabbedPane;
import com.intellij.ui.components.JBTextArea;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.chat.AssistantMessageWithSettingsButton;
import com.sourcegraph.cody.chat.Chat;
import com.sourcegraph.cody.chat.ChatBubble;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.chat.ChatUIConstants;
import com.sourcegraph.cody.chat.ContentWithGradientBorder;
import com.sourcegraph.cody.chat.ContextFilesMessage;
import com.sourcegraph.cody.chat.HumanMessageToMarkdownTextTransformer;
import com.sourcegraph.cody.chat.Interaction;
import com.sourcegraph.cody.chat.Transcript;
import com.sourcegraph.cody.context.ContextGetter;
import com.sourcegraph.cody.context.ContextMessage;
import com.sourcegraph.cody.editor.EditorContext;
import com.sourcegraph.cody.editor.EditorContextGetter;
import com.sourcegraph.cody.localapp.LocalAppManager;
import com.sourcegraph.cody.prompts.Preamble;
import com.sourcegraph.cody.prompts.Prompter;
import com.sourcegraph.cody.recipes.ExplainCodeDetailedAction;
import com.sourcegraph.cody.recipes.ExplainCodeHighLevelAction;
import com.sourcegraph.cody.recipes.FindCodeSmellsAction;
import com.sourcegraph.cody.recipes.GenerateDocStringAction;
import com.sourcegraph.cody.recipes.GenerateUnitTestAction;
import com.sourcegraph.cody.recipes.ImproveVariableNamesAction;
import com.sourcegraph.cody.recipes.OptimizeCodeAction;
import com.sourcegraph.cody.recipes.RecipeRunner;
import com.sourcegraph.cody.recipes.SummarizeRecentChangesRecipe;
import com.sourcegraph.cody.recipes.TranslateToLanguageAction;
import com.sourcegraph.cody.ui.HtmlViewer;
import com.sourcegraph.cody.ui.RoundedJBTextArea;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.SettingsComponent;
import com.sourcegraph.config.SettingsConfigurable;
import com.sourcegraph.telemetry.GraphQlLogger;
import com.sourcegraph.vcs.RepoUtil;
import java.awt.BorderLayout;
import java.awt.CardLayout;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.GridLayout;
import java.awt.event.AdjustmentListener;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import javax.swing.BorderFactory;
import javax.swing.BoxLayout;
import javax.swing.JButton;
import javax.swing.JComponent;
import javax.swing.JEditorPane;
import javax.swing.JPanel;
import javax.swing.border.Border;
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
  private final @NotNull CardLayout allContentLayout = new CardLayout();
  private final @NotNull JPanel allContentPanel = new JPanel(allContentLayout);
  private final @NotNull JBTabbedPane tabbedPane = new JBTabbedPane();
  private final @NotNull JPanel messagesPanel = new JPanel();
  private final @NotNull JBTextArea promptInput;
  private final @NotNull JButton sendButton;
  private final @NotNull Project project;
  private boolean needScrollingDown = true;
  private @NotNull Transcript transcript = new Transcript();
  private boolean isChatVisible = false;

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
    JButton explainCodeDetailedButton = createRecipeButton("Explain selected code (detailed)");
    explainCodeDetailedButton.addActionListener(
        e -> new ExplainCodeDetailedAction().executeRecipeWithPromptProvider(this, project));
    JButton explainCodeHighLevelButton = createRecipeButton("Explain selected code (high level)");
    explainCodeHighLevelButton.addActionListener(
        e -> new ExplainCodeHighLevelAction().executeRecipeWithPromptProvider(this, project));
    JButton generateUnitTestButton = createRecipeButton("Generate a unit test");
    generateUnitTestButton.addActionListener(
        e -> new GenerateUnitTestAction().executeRecipeWithPromptProvider(this, project));
    JButton generateDocstringButton = createRecipeButton("Generate a docstring");
    generateDocstringButton.addActionListener(
        e -> new GenerateDocStringAction().executeRecipeWithPromptProvider(this, project));
    JButton improveVariableNamesButton = createRecipeButton("Improve variable names");
    improveVariableNamesButton.addActionListener(
        e -> new ImproveVariableNamesAction().executeRecipeWithPromptProvider(this, project));
    JButton translateToLanguageButton = createRecipeButton("Translate to different language");
    translateToLanguageButton.addActionListener(
        e -> new TranslateToLanguageAction().executeAction(project));
    JButton gitHistoryButton = createRecipeButton("Summarize recent code changes");
    gitHistoryButton.addActionListener(
        e ->
            new SummarizeRecentChangesRecipe(project, this, recipeRunner).summarizeRecentChanges());
    JButton findCodeSmellsButton = createRecipeButton("Smell code");
    findCodeSmellsButton.addActionListener(
        e -> new FindCodeSmellsAction().executeRecipeWithPromptProvider(this, project));
    // JButton fixupButton = createWideButton("Fixup code from inline instructions");
    // fixupButton.addActionListener(e -> recipeRunner.runFixup());
    // JButton contextSearchButton = createWideButton("Codebase context search");
    // contextSearchButton.addActionListener(e -> recipeRunner.runContextSearch());
    // JButton releaseNotesButton = createWideButton("Generate release notes");
    // releaseNotesButton.addActionListener(e -> recipeRunner.runReleaseNotes());
    JButton optimizeCodeButton = createRecipeButton("Optimize code");
    optimizeCodeButton.addActionListener(
        e -> new OptimizeCodeAction().executeRecipeWithPromptProvider(this, project));
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

    JPanel appNotInstalledPanel = createAppNotInstalledPanel();
    JPanel appNotRunningPanel = createAppNotRunningPanel();
    allContentPanel.add(tabbedPane, "tabbedPane");
    allContentPanel.add(appNotInstalledPanel, "appNotInstalledPanel");
    allContentPanel.add(appNotRunningPanel, "appNotRunningPanel");
    allContentLayout.show(allContentPanel, "appNotInstalledPanel");
    updateVisibilityOfContentPanels();
    // Add welcome message
    addWelcomeMessage();
  }

  private void updateVisibilityOfContentPanels() {
    if (LocalAppManager.isPlatformSupported()
        && ConfigUtil.getInstanceType(project) == SettingsComponent.InstanceType.LOCAL_APP) {
      if (!LocalAppManager.isLocalAppInstalled()) {
        allContentLayout.show(allContentPanel, "appNotInstalledPanel");
        isChatVisible = false;
      } else if (!LocalAppManager.isLocalAppRunning()) {
        allContentLayout.show(allContentPanel, "appNotRunningPanel");
        isChatVisible = false;
      } else {
        allContentLayout.show(allContentPanel, "tabbedPane");
        isChatVisible = true;
      }
    } else {
      allContentLayout.show(allContentPanel, "tabbedPane");
      isChatVisible = true;
    }
  }

  @NotNull
  private JPanel createAppNotInstalledPanel() {
    JPanel appNotInstalledPanel =
        new ContentWithGradientBorder(ChatUIConstants.ASSISTANT_MESSAGE_GRADIENT_WIDTH);
    JEditorPane jEditorPane = HtmlViewer.createHtmlViewer(UIUtil.getPanelBackground());
    jEditorPane.setText(
        "<html><body><h2>Get Started</h2>"
            + "<p>This plugin requires the Cody desktop app to enable context fetching for your private code."
            + " Download and run the Cody desktop app to Configure your local code graph.</p><"
            + "/body></html>");
    appNotInstalledPanel.add(jEditorPane);
    JButton downloadCodyAppButton = createMainButton("Download Cody App");
    downloadCodyAppButton.putClientProperty(DarculaButtonUI.DEFAULT_STYLE_KEY, Boolean.TRUE);
    downloadCodyAppButton.addActionListener(
        e -> {
          BrowserUtil.browse("https://about.sourcegraph.com/app");
          updateVisibilityOfContentPanels();
        });
    Border margin = JBUI.Borders.empty(TEXT_MARGIN);
    jEditorPane.setBorder(margin);
    appNotInstalledPanel.add(downloadCodyAppButton);
    JPanel blankPanel = new JPanel();
    blankPanel.setBorder(margin);
    blankPanel.setOpaque(false);
    appNotInstalledPanel.add(blankPanel);
    JPanel wrapperAppNotInstalledPanel =
        new JPanel(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false));
    wrapperAppNotInstalledPanel.setBorder(margin);
    wrapperAppNotInstalledPanel.add(appNotInstalledPanel);
    JPanel goToSettingsPanel = createPanelWithGoToSettingsButton();
    wrapperAppNotInstalledPanel.add(goToSettingsPanel);
    return wrapperAppNotInstalledPanel;
  }

  @NotNull
  private JPanel createAppNotRunningPanel() {
    JPanel appNotRunningPanel =
        new ContentWithGradientBorder(ChatUIConstants.ASSISTANT_MESSAGE_GRADIENT_WIDTH);
    JEditorPane jEditorPane = HtmlViewer.createHtmlViewer(UIUtil.getPanelBackground());
    jEditorPane.setText(
        "<html><body><h2>Cody App Not Running</h2>"
            + "<p>This plugin requires the Cody desktop app to enable context fetching for your private code.</p><"
            + "/body></html>");
    appNotRunningPanel.add(jEditorPane);
    JButton downloadCodyAppButton = createMainButton("Open Cody App");
    downloadCodyAppButton.putClientProperty(DarculaButtonUI.DEFAULT_STYLE_KEY, Boolean.TRUE);
    downloadCodyAppButton.addActionListener(
        e -> {
          LocalAppManager.runLocalApp();
          updateVisibilityOfContentPanels();
        });
    Border margin = JBUI.Borders.empty(TEXT_MARGIN);
    jEditorPane.setBorder(margin);
    appNotRunningPanel.add(downloadCodyAppButton);
    JPanel blankPanel = new JPanel();
    blankPanel.setBorder(margin);
    blankPanel.setOpaque(false);
    appNotRunningPanel.add(blankPanel);
    JPanel wrapperAppNotRunningPanel =
        new JPanel(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false));
    wrapperAppNotRunningPanel.setBorder(margin);
    wrapperAppNotRunningPanel.add(appNotRunningPanel);
    JPanel goToSettingsPanel = createPanelWithGoToSettingsButton();
    wrapperAppNotRunningPanel.add(goToSettingsPanel);
    return wrapperAppNotRunningPanel;
  }

  private JPanel createPanelWithGoToSettingsButton() {
    JButton goToSettingsButton = new JButton("Sign in with an enterprise account");
    goToSettingsButton.addActionListener(
        e ->
            ShowSettingsUtil.getInstance().showSettingsDialog(project, SettingsConfigurable.class));
    ButtonUI buttonUI = (ButtonUI) DarculaButtonUI.createUI(goToSettingsButton);
    goToSettingsButton.setUI(buttonUI);
    JPanel panelWithSettingsButton = new JPanel(new BorderLayout());
    panelWithSettingsButton.setBorder(JBUI.Borders.empty(TEXT_MARGIN, 0));
    panelWithSettingsButton.add(goToSettingsButton, BorderLayout.CENTER);
    return panelWithSettingsButton;
  }

  @NotNull
  private JButton createRecipeButton(@NotNull String text) {
    JButton button = new JButton(text);
    button.setAlignmentX(Component.CENTER_ALIGNMENT);
    button.setMaximumSize(new Dimension(Integer.MAX_VALUE, button.getPreferredSize().height));
    ButtonUI buttonUI = (ButtonUI) DarculaButtonUI.createUI(button);
    button.setUI(buttonUI);
    return button;
  }

  @NotNull
  private static JButton createMainButton(@NotNull String text) {
    JButton button = new JButton(text);
    button.setAlignmentX(Component.CENTER_ALIGNMENT);
    ButtonUI buttonUI = (ButtonUI) DarculaButtonUI.createUI(button);
    button.setUI(buttonUI);
    return button;
  }

  private void addWelcomeMessage() {
    String accessToken = ConfigUtil.getProjectAccessToken(project);
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
    sendButton.addActionListener(
        e -> {
          GraphQlLogger.logCodyEvent(this.project, "recipe:chat-question", "clicked");
          sendMessage(project);
        });
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

    /* Insert Enter on Shift+Enter, Ctrl+Enter, Alt/Option+Enter, and Meta+Enter */
    KeyboardShortcut SHIFT_ENTER =
        new KeyboardShortcut(getKeyStroke(VK_ENTER, SHIFT_DOWN_MASK), null);
    KeyboardShortcut CTRL_ENTER =
        new KeyboardShortcut(getKeyStroke(VK_ENTER, CTRL_DOWN_MASK), null);
    KeyboardShortcut ALT_OR_OPTION_ENTER =
        new KeyboardShortcut(getKeyStroke(VK_ENTER, ALT_DOWN_MASK), null);
    KeyboardShortcut META_ENTER =
        new KeyboardShortcut(getKeyStroke(VK_ENTER, META_DOWN_MASK), null);
    ShortcutSet INSERT_ENTER_SHORTCUT =
        new CustomShortcutSet(CTRL_ENTER, SHIFT_ENTER, META_ENTER, ALT_OR_OPTION_ENTER);
    AnAction insertEnterAction =
        new DumbAwareAction() {
          @Override
          public void actionPerformed(@NotNull AnActionEvent e) {
            promptInput.insert("\n", promptInput.getCaretPosition());
          }
        };
    insertEnterAction.registerCustomShortcutSet(INSERT_ENTER_SHORTCUT, promptInput);

    /* Submit on enter */
    KeyboardShortcut JUST_ENTER = new KeyboardShortcut(getKeyStroke(VK_ENTER, 0), null);
    ShortcutSet DEFAULT_SUBMIT_ACTION_SHORTCUT = new CustomShortcutSet(JUST_ENTER);
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
              "I'm sorry, something went wrong. Please try again. The error message I got was: \""
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

  @Override
  public void refreshPanelsVisibility() {
    this.updateVisibilityOfContentPanels();
  }

  @Override
  public boolean isChatVisible() {
    return this.isChatVisible;
  }

  private void sendMessage(@NotNull Project project) {
    String messageText =
        new HumanMessageToMarkdownTextTransformer(promptInput.getText()).transform();
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
              String instanceUrl = ConfigUtil.getSourcegraphUrl(project);
              String accessToken = ConfigUtil.getProjectAccessToken(project);

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
                this.displayUsedContext(contextMessages);
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
              GraphQlLogger.logCodyEvent(this.project, "recipe:chat-question", "executed");
            });
  }

  @Override
  public void displayUsedContext(@NotNull List<ContextMessage> contextMessages) {
    // Use context
    if (contextMessages.size() == 0) {
      this.addMessageToChat(
          ChatMessage.createAssistantMessage(
              "I didn't find any context for your ask. I'll try to answer without further context."));
    } else {

      ContextFilesMessage contextFilesMessage = new ContextFilesMessage(contextMessages);
      var messageContentPanel = new JPanel(new BorderLayout());
      messageContentPanel.add(contextFilesMessage);
      this.addComponentToChat(messageContentPanel);
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
        logger.warn(
            "Unable to load context for message: "
                + humanMessage.getText()
                + ", in repo: "
                + repoName,
            e);
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
    VirtualFile fileFromTheRepository =
        currentFile != null
            ? currentFile
            : RepoUtil.getRootFileFromFirstGitRepository(project).orElse(null);
    if (fileFromTheRepository == null) {
      return null;
    }
    try {
      return RepoUtil.getRemoteRepoUrlWithoutScheme(project, fileFromTheRepository);
    } catch (Exception e) {
      return RepoUtil.getSimpleRepositoryName(project, fileFromTheRepository);
    }
  }

  public @NotNull JComponent getContentPanel() {
    return allContentPanel;
  }

  public void focusPromptInput() {
    if (tabbedPane.getSelectedIndex() == CHAT_TAB_INDEX) {
      promptInput.requestFocusInWindow();
      int textLength = promptInput.getDocument().getLength();
      promptInput.setCaretPosition(textLength);
    }
  }

  public JComponent getPreferredFocusableComponent() {
    return promptInput;
  }
}
