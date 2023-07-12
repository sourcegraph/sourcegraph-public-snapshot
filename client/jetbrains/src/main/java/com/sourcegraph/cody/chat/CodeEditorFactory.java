package com.sourcegraph.cody.chat;

import com.intellij.ide.highlighter.HighlighterFactory;
import com.intellij.lang.Language;
import com.intellij.openapi.command.WriteCommandAction;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.CaretModel;
import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.EditorFactory;
import com.intellij.openapi.editor.EditorSettings;
import com.intellij.openapi.editor.colors.EditorColorsManager;
import com.intellij.openapi.editor.event.EditorMouseEvent;
import com.intellij.openapi.editor.event.EditorMouseListener;
import com.intellij.openapi.editor.event.EditorMouseMotionListener;
import com.intellij.openapi.editor.ex.EditorEx;
import com.intellij.openapi.editor.highlighter.EditorHighlighter;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.fileTypes.FileType;
import com.intellij.openapi.fileTypes.FileTypeManager;
import com.intellij.openapi.fileTypes.PlainTextFileType;
import com.intellij.openapi.project.Project;
import com.intellij.util.ui.JBInsets;
import com.sourcegraph.cody.ui.TransparentButton;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.Toolkit;
import java.awt.datatransfer.Clipboard;
import java.awt.datatransfer.StringSelection;
import java.awt.event.ComponentAdapter;
import java.awt.event.ComponentEvent;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.awt.event.MouseMotionAdapter;
import java.time.Duration;
import java.util.Optional;
import javax.swing.BorderFactory;
import javax.swing.JComponent;
import javax.swing.JLayeredPane;
import javax.swing.JPanel;
import javax.swing.Timer;
import javax.swing.border.EmptyBorder;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class CodeEditorFactory {

  private static final Logger logger = Logger.getInstance(CodeEditorFactory.class);

  private final @NotNull Project project;
  private final @NotNull JPanel parentPanel;
  private final int rightMargin;
  private final int spaceBetweenButtons = 5;

  public CodeEditorFactory(@NotNull Project project, @NotNull JPanel parentPanel, int rightMargin) {
    this.project = project;
    this.parentPanel = parentPanel;
    this.rightMargin = rightMargin;
  }

  public JComponent createCodeEditor(@NotNull String code, @Nullable String language) {
    Document codeDocument = EditorFactory.getInstance().createDocument(code);
    EditorEx editor = (EditorEx) EditorFactory.getInstance().createViewer(codeDocument);
    setHighlighting(editor, language);
    fillEditorSettings(editor.getSettings());
    editor.setVerticalScrollbarVisible(false);
    editor.getGutterComponentEx().setPaintBackground(false);
    JComponent editorComponent = editor.getComponent();
    Dimension editorPreferredSize = editorComponent.getPreferredSize();
    TransparentButton copyButton = new TransparentButton("Copy");
    copyButton.setToolTipText("Copy text");
    copyButton.setVisible(false);
    // copy text to clipboard
    copyButton.addActionListener(
        e -> {
          String text = editor.getDocument().getText();
          StringSelection stringSelection = new StringSelection(text);
          Clipboard clipboard = Toolkit.getDefaultToolkit().getSystemClipboard();
          clipboard.setContents(stringSelection, null);
          copyButton.setText("Copied");
          Timer timer =
              new Timer((int) Duration.ofSeconds(2).toMillis(), it -> copyButton.setText("Copy"));
          timer.setRepeats(false);
          timer.start();
        });
    TransparentButton insertAtCursorButton = new TransparentButton("Insert at Cursor");
    insertAtCursorButton.setToolTipText("Insert text at current cursor position");
    insertAtCursorButton.setVisible(false);
    insertAtCursorButton.addActionListener(
        e -> {
          FileEditorManager fileEditorManager = FileEditorManager.getInstance(project);
          Editor mainEditor = fileEditorManager.getSelectedTextEditor();
          if (mainEditor != null) {
            CaretModel caretModel = mainEditor.getCaretModel();
            int caretPos = caretModel.getOffset();
            // Paste the text at the caret position
            Document document = mainEditor.getDocument();
            String text = editor.getDocument().getText();
            WriteCommandAction.runWriteCommandAction(
                project,
                () -> {
                  try {
                    document.insertString(caretPos, text);
                  } catch (Exception ex) {
                    logger.warn("Failed to insert text at cursor", ex);
                  }
                });
          }
        });
    Dimension copyButtonPreferredSize = copyButton.getPreferredSize();
    int halfOfButtonHeight = copyButtonPreferredSize.height / 2;
    JLayeredPane layeredEditorPane = new JLayeredPane();
    layeredEditorPane.setOpaque(false);
    // add right margin to show gradient from the parent on the right side
    layeredEditorPane.setBorder(new EmptyBorder(JBInsets.create(new Insets(0, rightMargin, 0, 0))));
    // layered pane should have width of the editor + gradient width and height of the editor + half
    // of the button height
    // to show button on the top border of the editor
    layeredEditorPane.setPreferredSize(
        new Dimension(
            editorPreferredSize.width + rightMargin,
            editorPreferredSize.height + halfOfButtonHeight));
    // add empty space to editor to show button on top border
    editorComponent.setBorder(
        BorderFactory.createEmptyBorder(halfOfButtonHeight, rightMargin, 0, 0));
    editorComponent.setOpaque(false);
    // place the editor in the layered pane
    editorComponent.setBounds(
        rightMargin,
        0,
        parentPanel.getSize().width - rightMargin,
        editorPreferredSize.height + halfOfButtonHeight);
    layeredEditorPane.add(editorComponent, JLayeredPane.DEFAULT_LAYER);
    // place the button in the layered pane on top border of the editor
    copyButton.setBounds(
        editorComponent.getWidth() - copyButtonPreferredSize.width,
        0,
        copyButtonPreferredSize.width,
        copyButtonPreferredSize.height);
    layeredEditorPane.add(copyButton, JLayeredPane.PALETTE_LAYER, 0);

    Dimension insertAtCursorButtonPreferredSize = insertAtCursorButton.getPreferredSize();
    insertAtCursorButton.setBounds(
        editorComponent.getWidth()
            - copyButtonPreferredSize.width
            - insertAtCursorButtonPreferredSize.width
            - spaceBetweenButtons,
        0,
        insertAtCursorButtonPreferredSize.width,
        insertAtCursorButtonPreferredSize.height);
    layeredEditorPane.add(insertAtCursorButton, JLayeredPane.PALETTE_LAYER, 0);
    // resize the editor and move the copy button when the parent panel is resized
    layeredEditorPane.addComponentListener(
        new ComponentAdapter() {
          @Override
          public void componentResized(ComponentEvent e) {
            editorComponent.setBounds(
                rightMargin,
                0,
                parentPanel.getSize().width - rightMargin,
                editorPreferredSize.height + halfOfButtonHeight);
            copyButton.setBounds(
                editorComponent.getWidth() - copyButtonPreferredSize.width,
                0,
                copyButtonPreferredSize.width,
                copyButtonPreferredSize.height);

            insertAtCursorButton.setBounds(
                editorComponent.getWidth()
                    - copyButtonPreferredSize.width
                    - insertAtCursorButtonPreferredSize.width
                    - spaceBetweenButtons,
                0,
                insertAtCursorButtonPreferredSize.width,
                insertAtCursorButtonPreferredSize.height);
          }
        });

    // show and hide the copy button when the mouse is over the editor or over the button
    MouseMotionAdapter buttonsMouseMotionAdapter =
        new MouseMotionAdapter() {
          @Override
          public void mouseMoved(MouseEvent e) {
            copyButton.setVisible(true);
            insertAtCursorButton.setVisible(true);
          }
        };
    MouseAdapter copyButtonMouseAdapter =
        new MouseAdapter() {
          @Override
          public void mouseExited(MouseEvent e) {
            copyButton.setVisible(false);
            insertAtCursorButton.setVisible(false);
          }
        };

    copyButton.addMouseMotionListener(buttonsMouseMotionAdapter);
    copyButton.addMouseListener(copyButtonMouseAdapter);
    insertAtCursorButton.addMouseMotionListener(buttonsMouseMotionAdapter);
    insertAtCursorButton.addMouseListener(copyButtonMouseAdapter);

    EditorMouseMotionListener editorMouseMotionListener =
        new EditorMouseMotionListener() {
          @Override
          public void mouseMoved(@NotNull EditorMouseEvent e) {
            copyButton.setVisible(true);
            insertAtCursorButton.setVisible(true);
          }
        };

    EditorMouseListener editorMouseListener =
        new EditorMouseListener() {
          @Override
          public void mouseExited(@NotNull EditorMouseEvent event) {
            copyButton.setVisible(false);
            insertAtCursorButton.setVisible(false);
          }
        };

    editor.addEditorMouseMotionListener(editorMouseMotionListener);
    editor.addEditorMouseListener(editorMouseListener);
    return layeredEditorPane;
  }

  private static void setHighlighting(@NotNull EditorEx editor, @Nullable String languageName) {
    FileType fileType =
        Language.getRegisteredLanguages().stream()
            .filter(it -> it.getDisplayName().equalsIgnoreCase(languageName))
            .findFirst()
            .flatMap(
                it -> Optional.ofNullable(FileTypeManager.getInstance().findFileTypeByLanguage(it)))
            .orElse(PlainTextFileType.INSTANCE);

    EditorHighlighter editorHighlighter =
        HighlighterFactory.createHighlighter(
            fileType, EditorColorsManager.getInstance().getSchemeForCurrentUITheme(), null);
    editor.setHighlighter(editorHighlighter);
  }

  private static void fillEditorSettings(@NotNull EditorSettings editorSettings) {
    editorSettings.setAdditionalColumnsCount(0);
    editorSettings.setAdditionalLinesCount(0);
    editorSettings.setGutterIconsShown(false);
    editorSettings.setWhitespacesShown(false);
    editorSettings.setLineMarkerAreaShown(false);
    editorSettings.setIndentGuidesShown(false);
    editorSettings.setLineNumbersShown(false);
    editorSettings.setUseSoftWraps(false);
    editorSettings.setCaretRowShown(false);
  }
}
