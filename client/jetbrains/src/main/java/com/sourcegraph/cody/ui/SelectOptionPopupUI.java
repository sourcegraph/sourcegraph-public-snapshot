package com.sourcegraph.cody.ui;

import static com.intellij.openapi.wm.IdeFocusManager.getGlobalInstance;

import com.intellij.ide.actions.BigPopupUI;
import com.intellij.openapi.actionSystem.*;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.application.ModalityState;
import com.intellij.openapi.application.ReadAction;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.DumbService;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.NlsContexts;
import com.intellij.openapi.util.SystemInfo;
import com.intellij.ui.*;
import com.intellij.ui.components.JBList;
import com.intellij.ui.components.JBTextField;
import com.intellij.ui.components.panels.NonOpaquePanel;
import com.intellij.util.Alarm;
import com.intellij.util.BooleanFunction;
import com.intellij.util.concurrency.SequentialTaskExecutor;
import com.intellij.util.ui.StartupUiUtil;
import com.intellij.util.ui.StatusText;
import com.intellij.util.ui.UIUtil;
import java.awt.*;
import java.awt.event.*;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.function.Consumer;
import java.util.stream.Collectors;
import javax.swing.*;
import javax.swing.event.DocumentEvent;
import org.jetbrains.annotations.Nls;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class SelectOptionPopupUI extends BigPopupUI {
  public static final int SEARCH_FIELD_COLUMNS = 25;

  private final List<String> myOptions;
  private final Consumer<String> runOnSelect;
  private boolean myIsItemSelected;
  private final Alarm myListRenderingAlarm = new Alarm(Alarm.ThreadToUse.SWING_THREAD);

  private final ExecutorService myExecutorService =
      SequentialTaskExecutor.createSequentialApplicationPoolExecutor(
          "Option selection list building");

  public SelectOptionPopupUI(
      @Nullable Project project, List<String> options, Consumer<String> runOnSelect) {
    super(project);
    this.myOptions = options;
    this.runOnSelect = runOnSelect;
    init();
    initSearchActions();
    initResultsList();
    initSearchField();
    initMySearchField();
  }

  private void onMouseClicked(@NotNull MouseEvent event) {
    int clickCount = event.getClickCount();
    if (clickCount > 1 && clickCount % 2 == 0) {
      event.consume();
      final int i = myResultsList.locationToIndex(event.getPoint());
      if (i != -1) {
        getGlobalInstance()
            .doWhenFocusSettlesDown(() -> getGlobalInstance().requestFocus(getField(), true));
        ApplicationManager.getApplication()
            .invokeLater(
                () -> {
                  myResultsList.setSelectedIndex(i);
                  executeCommand();
                });
      }
    }
  }

  private JTextField getField() {
    return mySearchField;
  }

  private void executeCommand() {
    String selectedValue = (String) myResultsList.getSelectedValue();
    if (selectedValue == null) return;
    runOnSelect.accept(selectedValue);
    mySearchField.setText("");
    searchFinishedHandler.run();
  }

  private void initSearchActions() {
    myResultsList.addMouseListener(
        new MouseAdapter() {
          @Override
          public void mouseClicked(MouseEvent e) {
            onMouseClicked(e);
          }
        });

    AnAction escape = ActionManager.getInstance().getAction("EditorEscape");
    DumbAwareAction.create(__ -> searchFinishedHandler.run())
        .registerCustomShortcutSet(
            escape == null ? CommonShortcuts.ESCAPE : escape.getShortcutSet(), this);

    DumbAwareAction.create(e -> executeCommand())
        .registerCustomShortcutSet(
            CustomShortcutSet.fromString(
                "ENTER", "shift ENTER", "alt ENTER", "alt shift ENTER", "meta ENTER"),
            mySearchField,
            this);
  }

  public void initResultsList() {
    myResultsList.addListSelectionListener(
        e -> {
          Object selectedValue = myResultsList.getSelectedValue();
          if (selectedValue == null) return;
          myIsItemSelected = true;
        });
  }

  private void initSearchField() {
    mySearchField
        .getDocument()
        .addDocumentListener(
            new DocumentAdapter() {
              @Override
              protected void textChanged(@NotNull DocumentEvent e) {

                if (mySearchField.hasFocus()) {
                  ApplicationManager.getApplication().invokeLater(() -> myIsItemSelected = false);

                  if (!myIsItemSelected) {
                    clearSelection();

                    // invoke later here allows to get correct pattern from mySearchField
                    ApplicationManager.getApplication().invokeLater(() -> rebuildList());
                  }

                  adjustMainListEmptyText(mySearchField);

                  adjustEmptyText(mySearchField, field -> true, "", "Select your option");
                }
              }
            });

    mySearchField.addFocusListener(
        new FocusAdapter() {
          @Override
          public void focusGained(FocusEvent e) {
            rebuildList();
          }

          @Override
          public void focusLost(FocusEvent e) {
            searchFinishedHandler.run();
          }
        });
  }

  public void initMySearchField() {
    adjustMainListEmptyText(mySearchField);

    initSearchField();

    mySearchField.setColumns(SEARCH_FIELD_COLUMNS);
  }

  private static void adjustMainListEmptyText(@NotNull JBTextField editor) {
    adjustEmptyText(editor, field -> field.getText().isEmpty(), "Select option", "");
  }

  public static void adjustEmptyText(
      @NotNull JBTextField textEditor,
      @NotNull BooleanFunction<? super JBTextField> function,
      @NotNull @NlsContexts.StatusText String leftText,
      @NotNull @NlsContexts.StatusText String rightText) {

    textEditor.putClientProperty("StatusVisibleFunction", function);
    StatusText statusText = textEditor.getEmptyText();
    statusText.setShowAboveCenter(false);
    statusText.setText(leftText, SimpleTextAttributes.GRAY_ATTRIBUTES);
    statusText.appendText(false, 0, rightText, SimpleTextAttributes.GRAY_ATTRIBUTES, null);
    statusText.setFont(UIUtil.getLabelFont(UIUtil.FontSize.SMALL));
  }

  private void rebuildList() {
    ApplicationManager.getApplication().assertIsDispatchThread();

    myListRenderingAlarm.cancelAllRequests();
    myResultsList.getEmptyText().setText("No options found");

    if (DumbService.getInstance(myProject).isDumb()) {
      myResultsList.setEmptyText("Select option is not supported while indexes are updating");
      return;
    }

    ReadAction.nonBlocking(
            () ->
                new CollectionListModel<Object>(
                    myOptions.stream()
                        .filter(it -> it.toLowerCase().startsWith(getSearchPattern().toLowerCase()))
                        .collect(Collectors.toList())))
        .coalesceBy(this)
        .finishOnUiThread(
            ModalityState.defaultModalityState(),
            model ->
                myListRenderingAlarm.addRequest(
                    () -> {
                      addListDataListener(model);
                      myResultsList.setModel(model);
                      model.allContentsChanged();
                    },
                    150))
        .submit(myExecutorService);
  }

  private void clearSelection() {
    myResultsList.getSelectionModel().clearSelection();
  }

  @Override
  public @NotNull JBList<Object> createList() {
    CollectionListModel<Object> myListModel = new CollectionListModel<>(myOptions);
    addListDataListener(myListModel);
    return new JBList<>(myListModel);
  }

  @Override
  protected @NotNull ListCellRenderer<Object> createCellRenderer() {
    return new DefaultListCellRenderer();
  }

  @Override
  protected @NotNull JPanel createTopLeftPanel() {
    JLabel myTextFieldTitle = new JLabel("Select option");
    JPanel topPanel = new NonOpaquePanel(new BorderLayout());
    Color foregroundColor =
        StartupUiUtil.isUnderDarcula()
            ? UIUtil.isUnderWin10LookAndFeel() ? JBColor.WHITE : new JBColor(Gray._240, Gray._200)
            : UIUtil.getLabelForeground();

    myTextFieldTitle.setForeground(foregroundColor);
    myTextFieldTitle.setBorder(BorderFactory.createEmptyBorder(3, 5, 5, 0));
    if (SystemInfo.isMac) {
      myTextFieldTitle.setFont(
          myTextFieldTitle
              .getFont()
              .deriveFont(Font.BOLD, myTextFieldTitle.getFont().getSize() - 1f));
    } else {
      myTextFieldTitle.setFont(myTextFieldTitle.getFont().deriveFont(Font.BOLD));
    }

    topPanel.add(myTextFieldTitle);

    return topPanel;
  }

  @Override
  protected @NotNull JPanel createSettingsPanel() {
    return new JPanel();
  }

  @Override
  protected @NotNull @Nls String getAccessibleName() {
    return "Select option";
  }

  @Override
  public void dispose() {}

  @Override
  public Dimension getMinimumSize() {
    return super.getExpandedSize();
  }
}
