package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.actionSystem.DataContext;
import com.intellij.openapi.command.WriteCommandAction;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.editor.actionSystem.EditorAction;
import com.intellij.openapi.editor.actionSystem.EditorActionHandler;
import com.intellij.openapi.project.Project;
import com.sourcegraph.telemetry.GraphQlLogger;
import java.util.List;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * The action that gets triggered when the user accepts a Cody completion.
 *
 * <p>The action works by reading the Inlay at the caret position and inserting the completion text
 * into the editor.
 */
public class AcceptCodyAutoCompleteAction extends EditorAction {
  public AcceptCodyAutoCompleteAction() {
    super(new AcceptCompletionActionHandler());
  }

  private static class AcceptCompletionActionHandler extends EditorActionHandler {

    @Override
    protected boolean isEnabledForCaret(
        @NotNull Editor editor, @NotNull Caret caret, DataContext dataContext) {
      // Returns false to fall back to normal TAB character if there is no suggestion at the caret.
      return CodyAutoCompleteManager.isEditorInstanceSupported(editor)
          && getCodyCompletionAtCaret(editor, caret) != null;
    }

    private @Nullable CodyAutoCompleteElementRenderer getCodyCompletionAtCaret(
        @NotNull Editor editor, @Nullable Caret caret) {
      if (caret == null) {
        return null;
      }
      // can't use editor.getInlayModel().getInlineElementAt(caret.getVisualPosition()) here, as it
      // requires a write EDT thread;
      // we work around it by just looking at a range containing a single point
      List<Inlay<?>> inlays =
          editor.getInlayModel().getInlineElementsInRange(caret.getOffset(), caret.getOffset());
      return (CodyAutoCompleteElementRenderer)
          inlays.stream()
              .filter(i -> i.getRenderer() instanceof CodyAutoCompleteElementRenderer)
              .map(Inlay::getRenderer)
              .findFirst()
              .orElse(null);
    }

    @Override
    protected void doExecute(
        @NotNull Editor editor, @Nullable Caret maybeCaret, DataContext dataContext) {

      if (maybeCaret == null) {
        List<Caret> carets = editor.getCaretModel().getAllCarets();
        if (carets.size() != 1) {
          // Only accept completion if there's a single caret.
          return;
        }
        maybeCaret = carets.get(0);
      }
      if (maybeCaret == null) {
        return;
      }
      final Caret caret = maybeCaret;

      CodyAutoCompleteElementRenderer completion = getCodyCompletionAtCaret(editor, caret);
      if (completion == null) {
        return;
      }

      /* Log the event */
      Project project = editor.getProject();
      if (project != null) {
        GraphQlLogger.logCodyEvent(project, "completion", "accepted");
      }

      WriteCommandAction.runWriteCommandAction(
          editor.getProject(),
          "Accept Cody Completion",
          "Cody", // TODO: what groupID should we use here?
          () -> {
            editor
                .getDocument()
                .replaceString(caret.getOffset(), caret.getOffset(), completion.text);
            editor.getCaretModel().moveToOffset(caret.getOffset() + completion.text.length());
            editor.getScrollingModel().scrollToCaret(ScrollType.MAKE_VISIBLE);
          });
    }
  }
}
