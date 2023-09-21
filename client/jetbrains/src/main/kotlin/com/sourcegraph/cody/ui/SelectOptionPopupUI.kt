package com.sourcegraph.cody.ui

import com.intellij.ide.actions.BigPopupUI
import com.intellij.openapi.actionSystem.ActionManager
import com.intellij.openapi.actionSystem.CommonShortcuts
import com.intellij.openapi.actionSystem.CustomShortcutSet
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.application.ModalityState
import com.intellij.openapi.application.ReadAction
import com.intellij.openapi.project.DumbAwareAction
import com.intellij.openapi.project.DumbService
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.SystemInfo
import com.intellij.openapi.wm.IdeFocusManager
import com.intellij.ui.CollectionListModel
import com.intellij.ui.DocumentAdapter
import com.intellij.ui.Gray
import com.intellij.ui.JBColor
import com.intellij.ui.SimpleTextAttributes
import com.intellij.ui.components.JBList
import com.intellij.ui.components.JBTextField
import com.intellij.ui.components.panels.NonOpaquePanel
import com.intellij.util.Alarm
import com.intellij.util.BooleanFunction
import com.intellij.util.concurrency.SequentialTaskExecutor
import com.intellij.util.ui.StartupUiUtil
import com.intellij.util.ui.UIUtil
import java.awt.BorderLayout
import java.awt.Dimension
import java.awt.Font
import java.awt.event.FocusAdapter
import java.awt.event.FocusEvent
import java.awt.event.MouseAdapter
import java.awt.event.MouseEvent
import java.util.Locale
import java.util.function.Consumer
import java.util.stream.Collectors
import javax.swing.BorderFactory
import javax.swing.DefaultListCellRenderer
import javax.swing.JLabel
import javax.swing.JPanel
import javax.swing.JTextField
import javax.swing.ListCellRenderer
import javax.swing.event.DocumentEvent
import org.jetbrains.annotations.Nls

class SelectOptionPopupUI(
    project: Project?,
    private val myOptions: List<String>,
    private val runOnSelect: Consumer<String>,
) : BigPopupUI(project) {
  private var myIsItemSelected = false
  private val myListRenderingAlarm = Alarm(Alarm.ThreadToUse.SWING_THREAD)
  private val myExecutorService =
      SequentialTaskExecutor.createSequentialApplicationPoolExecutor(
          "Option selection list building",
      )

  init {
    init()
    initSearchActions()
    initResultsList()
    initSearchField()
    initMySearchField()
  }

  private fun onMouseClicked(event: MouseEvent) {
    val clickCount = event.clickCount
    if (clickCount > 1 && clickCount % 2 == 0) {
      event.consume()
      val i = myResultsList.locationToIndex(event.point)
      if (i != -1) {
        IdeFocusManager.getGlobalInstance().doWhenFocusSettlesDown {
          IdeFocusManager.getGlobalInstance().requestFocus(field, true)
        }
        ApplicationManager.getApplication().invokeLater {
          myResultsList.selectedIndex = i
          executeCommand()
        }
      }
    }
  }

  private val field: JTextField
    get() = mySearchField

  private fun executeCommand() {
    val selectedValue = (myResultsList.selectedValue ?: return) as String
    runOnSelect.accept(selectedValue)
    mySearchField.text = ""
    searchFinishedHandler.run()
  }

  private fun initSearchActions() {
    myResultsList.addMouseListener(
        object : MouseAdapter() {
          override fun mouseClicked(e: MouseEvent) {
            onMouseClicked(e)
          }
        },
    )
    val escape = ActionManager.getInstance().getAction("EditorEscape")
    DumbAwareAction.create { _ -> searchFinishedHandler.run() }
        .registerCustomShortcutSet(
            escape?.shortcutSet ?: CommonShortcuts.ESCAPE,
            this,
        )
    DumbAwareAction.create { _ -> executeCommand() }
        .registerCustomShortcutSet(
            CustomShortcutSet.fromString(
                "ENTER",
                "shift ENTER",
                "alt ENTER",
                "alt shift ENTER",
                "meta ENTER",
            ),
            mySearchField,
            this,
        )
  }

  private fun initResultsList() {
    myResultsList.addListSelectionListener { _ ->
      myResultsList.selectedValue ?: return@addListSelectionListener
      myIsItemSelected = true
    }
  }

  private fun initSearchField() {
    mySearchField.document.addDocumentListener(
        object : DocumentAdapter() {
          override fun textChanged(e: DocumentEvent) {
            if (mySearchField.hasFocus()) {
              ApplicationManager.getApplication().invokeLater { myIsItemSelected = false }
              if (!myIsItemSelected) {
                clearSelection()

                // invoke later here allows to get correct pattern from mySearchField
                ApplicationManager.getApplication().invokeLater { rebuildList() }
              }
              adjustMainListEmptyText(mySearchField)
              adjustEmptyText(mySearchField, { _ -> true }, "", "Select your option")
            }
          }
        },
    )
    mySearchField.addFocusListener(
        object : FocusAdapter() {
          override fun focusGained(e: FocusEvent) {
            rebuildList()
          }

          override fun focusLost(e: FocusEvent) {
            searchFinishedHandler.run()
          }
        },
    )
  }

  private fun initMySearchField() {
    adjustMainListEmptyText(mySearchField)
    initSearchField()
    mySearchField.columns = SEARCH_FIELD_COLUMNS
  }

  private fun rebuildList() {
    ApplicationManager.getApplication().assertIsDispatchThread()
    myListRenderingAlarm.cancelAllRequests()
    myResultsList.emptyText.setText("No options found")
    if (DumbService.getInstance(myProject!!).isDumb) {
      myResultsList.setEmptyText("Select option is not supported while indexes are updating")
      return
    }
    ReadAction.nonBlocking<CollectionListModel<Any>> {
          CollectionListModel(
              myOptions
                  .stream()
                  .filter { it: String ->
                    it.lowercase(Locale.getDefault())
                        .startsWith(searchPattern.lowercase(Locale.getDefault()))
                  }
                  .collect(Collectors.toList()),
          )
        }
        .coalesceBy(this)
        .finishOnUiThread(
            ModalityState.defaultModalityState(),
        ) { model: CollectionListModel<Any> ->
          myListRenderingAlarm.addRequest(
              {
                addListDataListener(model)
                myResultsList.setModel(model)
                model.allContentsChanged()
              },
              150,
          )
        }
        .submit(myExecutorService)
  }

  private fun clearSelection() {
    myResultsList.selectionModel.clearSelection()
  }

  override fun createList(): JBList<Any> {
    val myListModel = CollectionListModel<Any>(myOptions)
    addListDataListener(myListModel)
    return JBList(myListModel)
  }

  override fun createCellRenderer(): ListCellRenderer<Any> {
    return DefaultListCellRenderer()
  }

  @Deprecated("Deprecated in Java")
  override fun createTopLeftPanel(): JPanel {
    val myTextFieldTitle = JLabel("Select option")
    val topPanel: JPanel = NonOpaquePanel(BorderLayout())
    val foregroundColor =
        if (StartupUiUtil.isUnderDarcula()) {
          if (UIUtil.isUnderWin10LookAndFeel()) {
            JBColor.WHITE
          } else {
            JBColor(Gray._240, Gray._200)
          }
        } else {
          UIUtil.getLabelForeground()
        }
    myTextFieldTitle.foreground = foregroundColor
    myTextFieldTitle.border = BorderFactory.createEmptyBorder(3, 5, 5, 0)
    if (SystemInfo.isMac) {
      myTextFieldTitle.font =
          myTextFieldTitle.font.deriveFont(Font.BOLD, myTextFieldTitle.font.size - 1f)
    } else {
      myTextFieldTitle.font = myTextFieldTitle.font.deriveFont(Font.BOLD)
    }
    topPanel.add(myTextFieldTitle)
    return topPanel
  }

  @Deprecated("Deprecated in Java")
  override fun createSettingsPanel(): JPanel {
    return JPanel()
  }

  @Nls
  override fun getAccessibleName(): String {
    return "Select option"
  }

  override fun dispose() {}

  override fun getMinimumSize(): Dimension {
    return super.getExpandedSize()
  }

  companion object {
    const val SEARCH_FIELD_COLUMNS = 25

    private fun adjustMainListEmptyText(editor: JBTextField) {
      adjustEmptyText(editor, { field: JBTextField -> field.text.isEmpty() }, "Select option", "")
    }

    fun adjustEmptyText(
        textEditor: JBTextField,
        function: BooleanFunction<in JBTextField>,
        leftText: String,
        rightText: String,
    ) {
      textEditor.putClientProperty("StatusVisibleFunction", function)
      val statusText = textEditor.emptyText
      statusText.setShowAboveCenter(false)
      statusText.setText(leftText, SimpleTextAttributes.GRAY_ATTRIBUTES)
      statusText.appendText(false, 0, rightText, SimpleTextAttributes.GRAY_ATTRIBUTES, null)
      statusText.setFont(UIUtil.getLabelFont(UIUtil.FontSize.SMALL))
    }
  }
}
