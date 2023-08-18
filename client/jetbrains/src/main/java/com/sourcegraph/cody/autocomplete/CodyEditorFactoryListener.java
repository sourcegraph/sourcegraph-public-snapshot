package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.command.CommandProcessor;
import com.intellij.openapi.editor.Caret;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.SelectionModel;
import com.intellij.openapi.editor.VisualPosition;
import com.intellij.openapi.editor.event.*;
import com.intellij.openapi.editor.ex.util.EditorUtil;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditor;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.fileEditor.TextEditor;
import com.intellij.openapi.fileEditor.impl.FileEditorManagerImpl;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.CodyAgentClient;
import com.sourcegraph.cody.agent.protocol.Position;
import com.sourcegraph.cody.agent.protocol.Range;
import com.sourcegraph.cody.agent.protocol.TextDocument;
import com.sourcegraph.cody.vscode.InlineAutoCompleteTriggerKind;
import com.sourcegraph.cody.vscode.InlineCompletionTriggerKind;
import com.sourcegraph.config.ConfigUtil;
import java.util.List;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * Determines when to trigger completions and when to clear completions.
 *
 * <p>IntelliJ doesn't have a built-in API to register "inline completion providers" similar to VS
 * Code. Instead, we manually listen to editor events like the caret position, selection changes,
 * and document edits.
 */
public class CodyEditorFactoryListener implements EditorFactoryListener {
  CodySelectionListener selectionListener = new CodySelectionListener();
  CaretListener caretListener = new CodyCaretListener();

  @Override
  public void editorCreated(@NotNull EditorFactoryEvent event) {
    if (!ConfigUtil.isCodyEnabled()) {
      return;
    }
    Editor editor = event.getEditor();
    informAgentAboutEditorChange(editor);
    Project project = editor.getProject();
    if (project == null || project.isDisposed()) {
      return;
    }
    Disposable disposable = Disposer.newDisposable("CodyEditorFactoryListener");
    EditorUtil.disposeWithEditor(editor, disposable);
    editor.getCaretModel().addCaretListener(this.caretListener, disposable);
    editor.getSelectionModel().addSelectionListener(this.selectionListener, disposable);
    editor.getDocument().addDocumentListener(new CodyDocumentListener(editor), disposable);
  }

  private static class CodyCaretListener implements CaretListener {

    @Override
    public void caretPositionChanged(@NotNull CaretEvent e) {
      if (!ConfigUtil.isCodyEnabled()) {
        return;
      }
      informAgentAboutEditorChange(e.getEditor());
      CodyAutoCompleteManager suggestions = CodyAutoCompleteManager.getInstance();
      if (suggestions.isEnabledForEditor(e.getEditor())
          && CodyEditorFactoryListener.isSelectedEditor(e.getEditor())) {
        suggestions.clearAutoCompleteSuggestions(e.getEditor());
        suggestions.triggerAutoComplete(
            e.getEditor(),
            e.getEditor().getCaretModel().getOffset(),
            InlineCompletionTriggerKind.AUTOMATIC);
      }
    }
  }

  private static class CodySelectionListener implements SelectionListener {
    @Override
    public void selectionChanged(@NotNull SelectionEvent e) {
      if (!ConfigUtil.isCodyEnabled()) {
        return;
      }
      informAgentAboutEditorChange(e.getEditor());
      CodyAutoCompleteManager suggestions = CodyAutoCompleteManager.getInstance();
      if (suggestions.isEnabledForEditor(e.getEditor())
          && CodyEditorFactoryListener.isSelectedEditor(e.getEditor())) {
        suggestions.clearAutoCompleteSuggestions(e.getEditor());
      }
    }
  }

  private static class CodyDocumentListener implements BulkAwareDocumentListener {
    private final Editor editor;

    public CodyDocumentListener(@NotNull Editor editor) {
      this.editor = editor;
    }

    public void documentChangedNonBulk(@NotNull DocumentEvent event) {
      if (!CodyEditorFactoryListener.isSelectedEditor(this.editor)) {
        return;
      }
      CodyAutoCompleteManager completions = CodyAutoCompleteManager.getInstance();
      completions.clearAutoCompleteSuggestions(this.editor);
      if (completions.isEnabledForEditor(this.editor)
          && !CommandProcessor.getInstance().isUndoTransparentActionInProgress()) {
        informAgentAboutEditorChange(this.editor);
        int changeOffset = event.getOffset() + event.getNewLength();
        if (this.editor.getCaretModel().getOffset() == changeOffset) {
          InlineAutoCompleteTriggerKind requestType =
              event.getOldLength() != event.getNewLength()
                  ? InlineAutoCompleteTriggerKind.Invoke
                  : InlineAutoCompleteTriggerKind.Automatic;
          completions.triggerAutoComplete(
              this.editor, changeOffset, InlineCompletionTriggerKind.AUTOMATIC);
        }
      }
    }
  }

  /**
   * Returns true if this editor is currently open and focused by the user. Returns true if this
   * editor is in a separate tab or not focused/selected by the user.
   */
  private static boolean isSelectedEditor(Editor editor) {
    if (editor == null) {
      return false;
    }
    Project project = editor.getProject();
    if (project == null || project.isDisposed()) {
      return false;
    }
    FileEditorManager editorManager = FileEditorManager.getInstance(project);
    if (editorManager == null) {
      return false;
    }
    if (editorManager instanceof FileEditorManagerImpl) {
      Editor current = ((FileEditorManagerImpl) editorManager).getSelectedTextEditor(true);
      return current != null && current.equals(editor);
    }
    FileEditor current = editorManager.getSelectedEditor();
    return current instanceof TextEditor && editor.equals(((TextEditor) current).getEditor());
  }

  @Nullable
  private static Range getSelection(Editor editor) {
    SelectionModel selectionModel = editor.getSelectionModel();
    VisualPosition selectionStartPosition = selectionModel.getSelectionStartPosition();
    VisualPosition selectionEndPosition = selectionModel.getSelectionEndPosition();
    if (selectionStartPosition != null && selectionEndPosition != null) {
      return new Range()
          .setStart(
              new Position()
                  .setLine(selectionStartPosition.line)
                  .setCharacter(selectionStartPosition.column))
          .setEnd(
              new Position()
                  .setLine(selectionEndPosition.line)
                  .setCharacter(selectionEndPosition.column));
    }
    List<Caret> carets = editor.getCaretModel().getAllCarets();
    if (!carets.isEmpty()) {
      Caret caret = carets.get(0);
      Position position =
          new Position()
              .setLine(caret.getLogicalPosition().line)
              .setCharacter(caret.getLogicalPosition().column);
      // A single-offset caret is a selection where end == start.
      return new Range().setStart(position).setEnd(position);
    }
    return null;
  }

  // Sends a textDocument/didChange notification to the agent server.
  public static void informAgentAboutEditorChange(@Nullable Editor editor) {
    if (editor == null) {
      return;
    }
    if (editor.getProject() == null) {
      return;
    }
    if (!CodyAgent.isConnected(editor.getProject())) {
      return;
    }
    CodyAgentClient client = CodyAgent.getClient(editor.getProject());
    if (client.server == null) {
      return;
    }
    VirtualFile file = FileDocumentManager.getInstance().getFile(editor.getDocument());
    if (file == null) {
      return;
    }
    TextDocument document =
        new TextDocument()
            .setFilePath(file.getPath())
            .setContent(editor.getDocument().getText())
            .setSelection(getSelection(editor));
    client.server.textDocumentDidChange(document);

    if (client.codebase == null) {
      return;
    }
    client.codebase.handlePotentialCodebaseChange(editor.getProject(), file);
  }
}
