package com.sourcegraph.cody;

import static com.intellij.ui.SimpleTextAttributes.STYLE_PLAIN;
import static java.awt.event.KeyEvent.VK_ENTER;
import static javax.swing.KeyStroke.getKeyStroke;

import com.intellij.icons.AllIcons;
import com.intellij.ide.BrowserUtil;
import com.intellij.ide.ui.laf.darcula.ui.DarculaButtonUI;
import com.intellij.openapi.actionSystem.ActionManager;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.CustomShortcutSet;
import com.intellij.openapi.actionSystem.DefaultActionGroup;
import com.intellij.openapi.actionSystem.KeyboardShortcut;
import com.intellij.openapi.actionSystem.ShortcutSet;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.VerticalFlowLayout;
import com.intellij.ui.ColorUtil;
import com.intellij.ui.DocumentAdapter;
import com.intellij.ui.SimpleTextAttributes;
import com.intellij.ui.components.AnActionLink;
import com.intellij.ui.components.JBPanelWithEmptyText;
import com.intellij.ui.components.JBTabbedPane;
import com.intellij.ui.components.JBTextArea;
import com.intellij.util.IconUtil;
import com.intellij.util.concurrency.annotations.RequiresEdt;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.StatusText;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.CodyAgentServer;
import com.sourcegraph.cody.agent.protocol.RecipeInfo;
import com.sourcegraph.cody.auth.ui.SignInWithSourcegraphAction;
import com.sourcegraph.cody.chat.AssistantMessageWithSettingsButton;
import com.sourcegraph.cody.chat.Chat;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.chat.ChatScrollPane;
import com.sourcegraph.cody.chat.ChatUIConstants;
import com.sourcegraph.cody.chat.ContextFilesMessage;
import com.sourcegraph.cody.chat.MessagePanel;
import com.sourcegraph.cody.chat.Transcript;
import com.sourcegraph.cody.config.AccountType;
import com.sourcegraph.cody.config.CodyAuthenticationManager;
import com.sourcegraph.cody.context.ContextMessage;
import com.sourcegraph.cody.context.EmbeddingStatusView;
import com.sourcegraph.cody.localapp.LocalAppManager;
import com.sourcegraph.cody.ui.AutoGrowingTextArea;
import com.sourcegraph.cody.ui.HtmlViewer;
import com.sourcegraph.cody.ui.UnderlinedActionLink;
import com.sourcegraph.cody.vscode.CancellationToken;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.telemetry.GraphQlLogger;
import java.awt.*;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import javax.swing.BorderFactory;
import javax.swing.BoxLayout;
import javax.swing.JButton;
import javax.swing.JComponent;
import javax.swing.JEditorPane;
import javax.swing.JPanel;
import javax.swing.border.Border;
import javax.swing.border.EmptyBorder;
import javax.swing.event.DocumentEvent;
import javax.swing.plaf.ButtonUI;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.builtInWebServer.BuiltInServerOptions;

public class CodyToolWindowContent implements UpdatableChat {
  public static Logger logger = Logger.getInstance(CodyToolWindowContent.class);
  public static final String SING_IN_WITH_SOURCEGRAPH_PANEL = "singInWithSourcegraphPanel";
  private static final int CHAT_TAB_INDEX = 0;
  private static final int RECIPES_TAB_INDEX = 1;
  private final @NotNull CardLayout allContentLayout = new CardLayout();
  private final @NotNull JPanel allContentPanel = new JPanel(allContentLayout);
  private final @NotNull JBTabbedPane tabbedPane = new JBTabbedPane();
  private final @NotNull JPanel messagesPanel = new JPanel();
  private final @NotNull JBTextArea promptInput;
  private final @NotNull JButton sendButton;
  private final @NotNull Project project;
  private final JButton stopGeneratingButton =
      new JButton("Stop generating", IconUtil.desaturate(AllIcons.Actions.Suspend));
  private final @NotNull JBPanelWithEmptyText recipesPanel;
  public final EmbeddingStatusView embeddingStatusView;
  private @NotNull volatile CancellationToken cancellationToken = new CancellationToken();
  private @NotNull Transcript transcript = new Transcript();
  private boolean isChatVisible = false;

  public CodyToolWindowContent(@NotNull Project project) {
    this.project = project;
    // Tabs
    @NotNull JPanel contentPanel = new JPanel();
    tabbedPane.insertTab("Chat", null, contentPanel, null, CHAT_TAB_INDEX);
    recipesPanel = new JBPanelWithEmptyText(new GridLayout(0, 1));
    recipesPanel.setLayout(new BoxLayout(recipesPanel, BoxLayout.Y_AXIS));
    tabbedPane.insertTab("Commands", null, recipesPanel, null, RECIPES_TAB_INDEX);

    // Initiate filling recipes panel in the background
    refreshRecipes();

    // Chat panel
    messagesPanel.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, true));
    ChatScrollPane chatPanel = new ChatScrollPane(messagesPanel);

    // Controls panel
    JPanel controlsPanel = new JPanel();
    controlsPanel.setLayout(new BorderLayout());
    controlsPanel.setBorder(new EmptyBorder(JBUI.insets(0, 14, 14, 14)));
    JPanel promptPanel = new JPanel(new BorderLayout());
    sendButton = createSendButton(this.project);
    AutoGrowingTextArea autoGrowingTextArea = new AutoGrowingTextArea(3, 9, promptPanel);
    promptInput = autoGrowingTextArea.getTextArea();
    /* Submit on enter */
    KeyboardShortcut JUST_ENTER = new KeyboardShortcut(getKeyStroke(VK_ENTER, 0), null);
    ShortcutSet DEFAULT_SUBMIT_ACTION_SHORTCUT = new CustomShortcutSet(JUST_ENTER);
    AnAction sendMessageAction =
        new DumbAwareAction() {
          @Override
          public void actionPerformed(@NotNull AnActionEvent e) {
            sendChatMessage(project);
          }
        };
    sendMessageAction.registerCustomShortcutSet(DEFAULT_SUBMIT_ACTION_SHORTCUT, promptInput);

    // Enable/disable the send button based on whether promptInput is empty
    promptInput
        .getDocument()
        .addDocumentListener(
            new DocumentAdapter() {
              @Override
              protected void textChanged(@NotNull DocumentEvent e) {
                sendButton.setEnabled(!promptInput.getText().isEmpty());
              }
            });

    promptPanel.add(autoGrowingTextArea.getScrollPane(), BorderLayout.CENTER);
    promptPanel.setBorder(BorderFactory.createEmptyBorder(0, 0, 10, 0));

    JPanel stopGeneratingButtonPanel = new JPanel(new FlowLayout(FlowLayout.CENTER, 0, 5));
    stopGeneratingButtonPanel.setPreferredSize(
        new Dimension(Short.MAX_VALUE, stopGeneratingButton.getPreferredSize().height + 10));
    stopGeneratingButton.addActionListener(
        e -> {
          cancellationToken.abort();
          stopGeneratingButton.setVisible(false);
          sendButton.setEnabled(true);
        });
    stopGeneratingButton.setVisible(false);
    stopGeneratingButtonPanel.add(stopGeneratingButton);
    stopGeneratingButtonPanel.setOpaque(false);
    controlsPanel.add(promptPanel, BorderLayout.NORTH);
    controlsPanel.add(sendButton, BorderLayout.EAST);
    JPanel lowerPanel = new JPanel(new BorderLayout());
    Color borderColor = ColorUtil.brighter(UIUtil.getPanelBackground(), 3);
    Border topBorder = BorderFactory.createMatteBorder(1, 0, 0, 0, borderColor);
    lowerPanel.setBorder(topBorder);
    lowerPanel.setLayout(new BoxLayout(lowerPanel, BoxLayout.Y_AXIS));
    lowerPanel.add(stopGeneratingButtonPanel);
    lowerPanel.add(controlsPanel);

    embeddingStatusView = new EmbeddingStatusView(project);
    embeddingStatusView.setBorder(topBorder);
    lowerPanel.add(embeddingStatusView);

    // Main content panel
    contentPanel.setLayout(new BorderLayout(0, 0));
    contentPanel.setBorder(BorderFactory.createEmptyBorder(0, 0, 10, 0));

    contentPanel.add(chatPanel, BorderLayout.CENTER);
    contentPanel.add(lowerPanel, BorderLayout.SOUTH);

    tabbedPane.addChangeListener(e -> this.focusPromptInput());

    JPanel singInWithSourcegraphPanel = createSignInWithSourcegraphPanel();
    JPanel appNotRunningPanel = createAppNotRunningPanel();
    allContentPanel.add(tabbedPane, "tabbedPane");
    allContentPanel.add(singInWithSourcegraphPanel, SING_IN_WITH_SOURCEGRAPH_PANEL);
    allContentPanel.add(appNotRunningPanel, "appNotRunningPanel");
    allContentLayout.show(allContentPanel, SING_IN_WITH_SOURCEGRAPH_PANEL);
    updateVisibilityOfContentPanels();
    // Add welcome message
    addWelcomeMessage();
  }

  @RequiresEdt
  public void refreshRecipes() {
    recipesPanel.removeAll();
    recipesPanel.getEmptyText().setText("Loading recipes...");
    recipesPanel.revalidate();
    recipesPanel.repaint();

    CodyAgentServer server = CodyAgent.getServer(project);
    if (server == null) {
      setRecipesPanelError();
      return;
    }

    ApplicationManager.getApplication()
        .executeOnPooledThread( // Non-blocking data fetch
            () -> {
              try {
                server
                    .recipesList()
                    .thenAccept(
                        (List<RecipeInfo> recipes) ->
                            ApplicationManager.getApplication()
                                .invokeLater(
                                    () -> updateUIWithRecipeList(recipes))); // Update on EDT
              } catch (Exception e) {
                logger.warn("Error fetching recipes from agent", e);
                // Update on EDT
                ApplicationManager.getApplication().invokeLater(this::setRecipesPanelError);
              }
            });
  }

  @RequiresEdt
  private void setRecipesPanelError() {
    StatusText emptyText = recipesPanel.getEmptyText();

    emptyText.setText(
        "Error fetching recipes. Check your connection. If the problem persists, please contact support.");
    emptyText.appendLine(
        "Retry",
        new SimpleTextAttributes(STYLE_PLAIN, JBUI.CurrentTheme.Link.Foreground.ENABLED),
        __ -> refreshRecipes());
  }

  @RequiresEdt
  private void updateUIWithRecipeList(@NotNull List<RecipeInfo> recipes) {
    // we don't want to display recipes with ID "chat-question" and "code-question"
    var excludedRecipeIds = List.of("chat-question", "code-question", "translate-to-language");
    var recipesToDisplay =
        recipes.stream()
            .filter(recipe -> !excludedRecipeIds.contains(recipe.id))
            .collect(Collectors.toList());

    fillRecipesPanel(recipesToDisplay);
    fillContextMenu(recipesToDisplay);
  }

  @RequiresEdt
  private void fillRecipesPanel(@NotNull List<RecipeInfo> recipes) {
    recipesPanel.removeAll();

    // Loop on recipes and add a button for each item
    for (RecipeInfo recipe : recipes) {
      JButton recipeButton = createRecipeButton(recipe.title);
      recipeButton.addActionListener(
          e -> {
            GraphQlLogger.logCodyEvent(this.project, "recipe:" + recipe.id, "clicked");
            sendMessage(this.project, recipe.title, recipe.id);
          });
      recipesPanel.add(recipeButton);
    }
  }

  private void fillContextMenu(@NotNull List<RecipeInfo> recipes) {
    ActionManager actionManager = ActionManager.getInstance();
    DefaultActionGroup group = (DefaultActionGroup) actionManager.getAction("CodyEditorActions");

    // Loop on recipes and create an action for each new item
    for (RecipeInfo recipe : recipes) {
      String actionId = "cody.recipe." + recipe.id;
      var existingAction = actionManager.getAction(actionId);
      if (existingAction != null) {
        continue;
      }
      var action =
          new DumbAwareAction(recipe.title) {
            @Override
            public void actionPerformed(@NotNull AnActionEvent e) {
              GraphQlLogger.logCodyEvent(project, "recipe:" + recipe.id, "clicked");
              sendMessage(project, recipe.title, recipe.id);
            }
          };
      actionManager.registerAction(actionId, action);
      group.addAction(action);
    }
  }

  public static @NotNull CodyToolWindowContent getInstance(@NotNull Project project) {
    return project.getService(CodyToolWindowContent.class);
  }

  @RequiresEdt
  private void updateVisibilityOfContentPanels() {
    CodyAuthenticationManager codyAuthenticationManager = CodyAuthenticationManager.getInstance();
    if (codyAuthenticationManager.getAccounts().isEmpty()) {
      allContentLayout.show(allContentPanel, SING_IN_WITH_SOURCEGRAPH_PANEL);
      isChatVisible = false;
      return;
    }
    if (LocalAppManager.isPlatformSupported()
        && codyAuthenticationManager.getActiveAccountType(project) == AccountType.LOCAL_APP) {
      if (!LocalAppManager.isLocalAppRunning()) {
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

  @RequiresEdt
  private @NotNull JPanel createSignInWithSourcegraphPanel() {
    JEditorPane jEditorPane = HtmlViewer.createHtmlViewer(UIUtil.getPanelBackground());
    jEditorPane.setText(
        "<html><body><h2>Welcome to Cody</h2>"
            + "<p>Understand and write code faster with an AI assistant</p>"
            + "</body></html>");
    JButton signInWithSourcegraphButton = createMainButton("Sign in for free with Sourcegraph.com");
    signInWithSourcegraphButton.putClientProperty(DarculaButtonUI.DEFAULT_STYLE_KEY, Boolean.TRUE);
    signInWithSourcegraphButton.addActionListener(
        e -> {
          int port =
              ApplicationManager.getApplication()
                  .getService(BuiltInServerOptions.class)
                  .getEffectiveBuiltInServerPort();
          BrowserUtil.browse(
              ConfigUtil.DOTCOM_URL
                  + "user/settings/tokens/new/callback"
                  + "?requestFrom=JETBRAINS"
                  + "&port="
                  + port);
          updateVisibilityOfContentPanels();
        });
    CodyOnboardingPanel codyOnboardingPanel =
        new CodyOnboardingPanel(jEditorPane, signInWithSourcegraphButton);
    JPanel signInWithAnEnterpriseInstance = createPanelWithSignInWithAnEnterpriseInstance();
    codyOnboardingPanel.add(signInWithAnEnterpriseInstance);
    return codyOnboardingPanel;
  }

  private JPanel createPanelWithSignInWithAnEnterpriseInstance() {
    AnActionLink signInWithAnEnterpriseInstance =
        new UnderlinedActionLink(
            "Sign in with an Enterprise Instance", new SignInWithSourcegraphAction(""));
    signInWithAnEnterpriseInstance.setAlignmentX(Component.CENTER_ALIGNMENT);
    JPanel panelWithSettingsLink = new JPanel(new BorderLayout());
    panelWithSettingsLink.setBorder(JBUI.Borders.empty(20, 0));
    JPanel linkPanel = new JPanel(new GridBagLayout());
    linkPanel.add(signInWithAnEnterpriseInstance);
    panelWithSettingsLink.add(linkPanel, BorderLayout.PAGE_START);
    return panelWithSettingsLink;
  }

  @RequiresEdt
  private @NotNull JPanel createAppNotRunningPanel() {

    JEditorPane jEditorPane = HtmlViewer.createHtmlViewer(UIUtil.getPanelBackground());
    jEditorPane.setText(
        "<html><body><h2>Cody App Not Running</h2>"
            + "<p>This plugin requires the Cody desktop app to enable context fetching for your private code.</p><"
            + "/body></html>");
    JButton runCodyAppButton = createMainButton("Open Cody App");
    runCodyAppButton.putClientProperty(DarculaButtonUI.DEFAULT_STYLE_KEY, Boolean.TRUE);
    runCodyAppButton.addActionListener(
        e -> {
          LocalAppManager.runLocalApp();
          updateVisibilityOfContentPanels();
        });
    return new CodyOnboardingPanel(jEditorPane, runCodyAppButton);
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
    button.setMaximumSize(new Dimension(Short.MAX_VALUE, button.getPreferredSize().height));
    button.setAlignmentX(Component.CENTER_ALIGNMENT);
    ButtonUI buttonUI = (ButtonUI) DarculaButtonUI.createUI(button);
    button.setUI(buttonUI);
    return button;
  }

  private void addWelcomeMessage() {
    String welcomeText =
        "Hello! I'm Cody. I can write code and answer questions for you. See [Cody documentation](https://docs.sourcegraph.com/cody) for help and tips.";
    addMessageToChat(ChatMessage.createAssistantMessage(welcomeText));
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
          sendChatMessage(project);
        });
    return sendButton;
  }

  public synchronized void addMessageToChat(@NotNull ChatMessage message) {
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              // Bubble panel
              MessagePanel messagePanel =
                  new MessagePanel(
                      message,
                      project,
                      messagesPanel,
                      ChatUIConstants.ASSISTANT_MESSAGE_GRADIENT_WIDTH);
              addComponentToChat(messagePanel);
            });
  }

  public void addComponentToChat(@NotNull JPanel messageContent) {

    var wrapperPanel = new JPanel();
    wrapperPanel.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false));
    // Chat message
    wrapperPanel.add(messageContent, VerticalFlowLayout.TOP);
    messagesPanel.add(wrapperPanel);
    messagesPanel.revalidate();
    messagesPanel.repaint();
  }

  @Override
  public void activateChatTab() {
    this.tabbedPane.setSelectedIndex(CHAT_TAB_INDEX);
  }

  @Override
  public void respondToErrorFromServer(@NotNull String errorMessage) {
    if (errorMessage.equals("Connection refused")) {
      this.addMessageToChat(
          ChatMessage.createAssistantMessage(
              "I'm sorry, I can't connect to the server. Please make sure that the server is running and try again."));
    } else if (errorMessage.startsWith("Got error response 401")) {
      addMessageWithInformationAboutInvalidAccessToken();
    } else {
      this.addMessageToChat(
          ChatMessage.createAssistantMessage(
              "I'm sorry, something went wrong. Please try again. The error message I got was: \""
                  + errorMessage
                  + "\"."));
    }
  }

  private void addMessageWithInformationAboutInvalidAccessToken() {
    String invalidAccessTokenText =
        "<p>It looks like your Sourcegraph Access Token is invalid or not configured.</p>"
            + "<p>See our <a href=\"https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token\">user docs</a> how to create one and configure it in the settings to use Cody.</p>";
    AssistantMessageWithSettingsButton assistantMessageWithSettingsButton =
        new AssistantMessageWithSettingsButton(invalidAccessTokenText);
    var messageContentPanel = new JPanel(new BorderLayout());
    messageContentPanel.add(assistantMessageWithSettingsButton);
    ApplicationManager.getApplication()
        .invokeLater(() -> this.addComponentToChat(messageContentPanel));
  }

  public synchronized void updateLastMessage(@NotNull ChatMessage message) {
    ApplicationManager.getApplication()
        .invokeLater(
            () ->
                Optional.of(messagesPanel)
                    .filter(mp -> mp.getComponentCount() > 0)
                    .map(mp -> mp.getComponent(mp.getComponentCount() - 1))
                    .filter(component -> component instanceof JPanel)
                    .map(component -> (JPanel) component)
                    .map(lastWrapperPanel -> lastWrapperPanel.getComponent(0))
                    .filter(component -> component instanceof MessagePanel)
                    .map(component -> (MessagePanel) component)
                    .ifPresent(
                        lastMessage -> {
                          transcript.addAssistantResponse(message);
                          lastMessage.updateContentWith(message);
                        }));
  }

  private void startMessageProcessing() {
    cancellationToken.abort();
    cancellationToken = new CancellationToken();
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              stopGeneratingButton.setVisible(true);
              sendButton.setEnabled(false);
            });
  }

  @Override
  public void finishMessageProcessing() {
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              stopGeneratingButton.setVisible(false);
              sendButton.setEnabled(true);
            });
  }

  @Override
  public void resetConversation() {
    transcript = new Transcript();
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              cancellationToken.abort();
              stopGeneratingButton.setVisible(false);
              sendButton.setEnabled(true);
              messagesPanel.removeAll();
              addWelcomeMessage();
              messagesPanel.revalidate();
              messagesPanel.repaint();
              CodyAgent.getInitializedServer(project).thenAccept(CodyAgentServer::transcriptReset);
            });
  }

  @Override
  @RequiresEdt
  public void refreshPanelsVisibility() {
    this.updateVisibilityOfContentPanels();
  }

  @Override
  public boolean isChatVisible() {
    return this.isChatVisible;
  }

  @RequiresEdt
  private void sendChatMessage(@NotNull Project project) {
    String text = promptInput.getText();
    sendMessage(project, text, "chat-question");
    promptInput.setText("");
  }

  @RequiresEdt
  private void sendMessage(
      @NotNull Project project, @NotNull String message, @NotNull String recipeId) {
    if (!sendButton.isEnabled()) {
      return;
    }

    startMessageProcessing();

    ChatMessage humanMessage = ChatMessage.createHumanMessage(message, message);
    addMessageToChat(humanMessage);
    activateChatTab();

    // This cannot run on EDT (Event Dispatch Thread) because it may block for a long time.
    // Also, if we did the back-end call in the main thread and then waited, we wouldn't see the
    // messages streamed back to us.
    ApplicationManager.getApplication()
        .executeOnPooledThread(
            () -> {
              Chat chat = new Chat();
              if (CodyAgent.isConnected(project)) {
                try {
                  chat.sendMessageViaAgent(
                      CodyAgent.getClient(project),
                      CodyAgent.getInitializedServer(project),
                      humanMessage,
                      recipeId,
                      this,
                      cancellationToken);
                } catch (Exception e) {
                  logger.warn("Error sending message '" + humanMessage + "' to chat", e);
                }
              } else {
                logger.warn("Agent is disabled, can't use chat.");
                this.addMessageToChat(
                    ChatMessage.createAssistantMessage(
                        "Cody is not able to reply at the moment. "
                            + "This is a bug, please report an issue to the sourcegraph/sourcegraph "
                            + "repository and try to include relevant context from idea.log if possible."));
                this.finishMessageProcessing();
              }
              GraphQlLogger.logCodyEvent(this.project, "recipe:chat-question", "executed");
            });
  }

  @Override
  public void displayUsedContext(@NotNull List<ContextMessage> contextMessages) {
    if (contextMessages.isEmpty()) {
      // Do nothing when there are no context files. It's normal that some answers have no context
      // files.
      return;
    }

    ContextFilesMessage contextFilesMessage = new ContextFilesMessage(contextMessages);
    var messageContentPanel = new JPanel(new BorderLayout());
    messageContentPanel.add(contextFilesMessage);
    this.addComponentToChat(messageContentPanel);
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
