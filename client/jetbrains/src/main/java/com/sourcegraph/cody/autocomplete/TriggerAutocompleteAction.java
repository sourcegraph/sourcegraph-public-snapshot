package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.actionSystem.DataContext;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.Caret;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.actionSystem.EditorAction;
import com.intellij.openapi.editor.actionSystem.EditorActionHandler;
import com.sourcegraph.cody.vscode.InlineCompletionTriggerKind;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class TriggerAutocompleteAction extends EditorAction {
  public TriggerAutocompleteAction() {
    super(new TriggerAutocompleteActionHandler());
  }

  private static class TriggerAutocompleteActionHandler extends EditorActionHandler {
    Logger logger = Logger.getInstance(TriggerAutocompleteActionHandler.class);

    @Override
    protected boolean isEnabledForCaret(
        @NotNull Editor editor, @NotNull Caret caret, DataContext dataContext) {
      return CodyAutoCompleteManager.isEditorInstanceSupported(editor);
    }

    @Override
    protected void doExecute(
        @NotNull Editor editor, @Nullable Caret caret, DataContext dataContext) {

      int offset =
          caret == null ? editor.getCaretModel().getCurrentCaret().getOffset() : caret.getOffset();
      CodyAutoCompleteManager.getInstance()
          .triggerAutoComplete(editor, offset, InlineCompletionTriggerKind.INVOKE);
    }
  }
}
