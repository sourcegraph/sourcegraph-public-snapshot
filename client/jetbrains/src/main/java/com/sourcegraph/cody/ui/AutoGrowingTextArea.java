package com.sourcegraph.cody.ui;

import static java.awt.event.InputEvent.ALT_DOWN_MASK;
import static java.awt.event.InputEvent.CTRL_DOWN_MASK;
import static java.awt.event.InputEvent.META_DOWN_MASK;
import static java.awt.event.InputEvent.SHIFT_DOWN_MASK;
import static java.awt.event.KeyEvent.VK_ENTER;
import static javax.swing.KeyStroke.getKeyStroke;

import com.intellij.ide.ui.laf.darcula.ui.DarculaTextAreaUI;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.CustomShortcutSet;
import com.intellij.openapi.actionSystem.KeyboardShortcut;
import com.intellij.openapi.actionSystem.ShortcutSet;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.ui.components.JBScrollPane;
import com.intellij.ui.components.JBTextArea;
import com.intellij.util.ui.UIUtil;
import java.awt.Dimension;
import java.awt.FontMetrics;
import javax.swing.JPanel;
import javax.swing.ScrollPaneConstants;
import javax.swing.plaf.basic.BasicTextAreaUI;
import javax.swing.text.AttributeSet;
import javax.swing.text.BadLocationException;
import javax.swing.text.Document;
import javax.swing.text.PlainDocument;
import org.jetbrains.annotations.NotNull;

public class AutoGrowingTextArea {

  private final @NotNull JBTextArea textArea;
  private final @NotNull JBScrollPane scrollPane;
  private final Dimension initialPreferredSize;
  private final int minRows;
  private final int autoGrowUpToRow;

  public AutoGrowingTextArea(int minRows, int maxRows, JPanel outerPanel) {
    this.minRows = minRows;
    this.autoGrowUpToRow = maxRows + 1;
    textArea = createTextArea();
    scrollPane = new JBScrollPane(textArea);
    initialPreferredSize = scrollPane.getPreferredSize();
    Document document =
        new PlainDocument() {
          @Override
          public void insertString(int offs, String str, AttributeSet a)
              throws BadLocationException {
            super.insertString(offs, str, a);
            updateTextAreaSize();
            outerPanel.revalidate();
          }

          @Override
          public void remove(int offs, int len) throws BadLocationException {
            super.remove(offs, len);
            updateTextAreaSize();
            outerPanel.revalidate();
          }
        };

    textArea.setDocument(document);
  }

  @NotNull
  private JBTextArea createTextArea() {
    JBTextArea promptInput = new RoundedJBTextArea(minRows, 10);
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

    return promptInput;
  }

  private void updateTextAreaSize() {
    // Get the preferred size of the JTextArea based on its content
    Dimension preferredSize = textArea.getPreferredSize();
    // Limit the number of rows to maxRows
    FontMetrics fontMetrics = textArea.getFontMetrics(textArea.getFont());
    int maxTextAreaHeight = fontMetrics.getHeight() * autoGrowUpToRow;
    int preferredHeight = Math.min(preferredSize.height, maxTextAreaHeight);
    preferredHeight = Math.max(preferredHeight, initialPreferredSize.height);

    // Set the preferred size of the JScrollPane to accommodate the JTextArea
    Dimension scrollPaneSize = scrollPane.getSize();
    scrollPaneSize.height = preferredHeight;
    scrollPane.setPreferredSize(scrollPaneSize);

    boolean shouldShowScrollbar = preferredSize.height > maxTextAreaHeight;
    scrollPane.setVerticalScrollBarPolicy(
        shouldShowScrollbar
            ? ScrollPaneConstants.VERTICAL_SCROLLBAR_ALWAYS
            : ScrollPaneConstants.VERTICAL_SCROLLBAR_NEVER);
  }

  public @NotNull JBTextArea getTextArea() {
    return textArea;
  }

  public @NotNull JBScrollPane getScrollPane() {
    return scrollPane;
  }
}
