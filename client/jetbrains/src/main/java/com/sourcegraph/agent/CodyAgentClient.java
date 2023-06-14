package com.sourcegraph.agent;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.agent.protocol.*;
import com.sourcegraph.cody.editor.EditorContext;
import com.sourcegraph.cody.editor.EditorContextGetter;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.function.Consumer;
import java.util.function.Supplier;
import org.eclipse.lsp4j.jsonrpc.services.JsonNotification;
import org.eclipse.lsp4j.jsonrpc.services.JsonRequest;
import org.jetbrains.annotations.Nullable;

/** Implementation of the client part of the Cody agent protocol. */
public class CodyAgentClient {

  @Nullable public CodyAgentServer server;
  @Nullable public Consumer<ChatMessage> onChatUpdateMessageInProgress;
  @Nullable public Editor editor;

  /**
   * Helper to run client request/notification handlers on the IntelliJ event thread. Use this
   * helper for handlers that require access to the IntelliJ editor, for example to read the text
   * contents of the open editor.
   */
  private <T> CompletableFuture<T> onEventThread(Supplier<T> handler) {
    CompletableFuture<T> result = new CompletableFuture<>();
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              try {
                result.complete(handler.get());
              } catch (Exception e) {
                result.completeExceptionally(e);
              }
            });
    return result;
  }

  // ========
  // Requests
  // ========

  @JsonRequest("editor/quickPick")
  public CompletableFuture<List<String>> editorQuickPick(List<String> params) {
    // TODO
    return CompletableFuture.completedFuture(null);
  }

  @JsonRequest("editor/prompt")
  public CompletableFuture<String> editorPrompt(String params) {
    // TODO
    return CompletableFuture.completedFuture(null);
  }

  @JsonRequest("editor/active")
  public CompletableFuture<ActiveTextEditor> editorActive() {
    return this.onEventThread(
        () -> {
          if (editor == null) {
            return null;
          }
          VirtualFile file = FileDocumentManager.getInstance().getFile(editor.getDocument());
          if (file == null) {
            return null;
          }
          return new ActiveTextEditor()
              .setContent(editor.getDocument().getText())
              .setFilePath(file.getPath());
        });
  }

  @JsonRequest("editor/selection")
  public CompletableFuture<ActiveTextEditorSelection> editorSelection() {
    return onEventThread(
        () -> {
          if (editor == null) {
            return null;
          }
          Project project = editor.getProject();
          if (project == null) {
            return null;
          }
          EditorContext context = EditorContextGetter.getEditorContext(project);
          return new ActiveTextEditorSelection()
              .setFileName(context.getCurrentFileName())
              .setPrecedingText(context.getPrecedingText())
              .setSelectedText(context.getSelection())
              .setFollowingText(context.getFollowingText());
        });
  }

  @JsonRequest("editor/selectionOrEntireFile")
  public CompletableFuture<ActiveTextEditorSelection> editorActiveOrEntireFile() {
    // TODO
    return CompletableFuture.completedFuture(null);
  }

  @JsonRequest("editor/visibleContent")
  public CompletableFuture<ActiveTextEditorVisibleContent> editorVisibleContent() {
    return onEventThread(
        () -> {
          if (editor == null) {
            return null;
          }
          VirtualFile file = FileDocumentManager.getInstance().getFile(editor.getDocument());
          if (file == null) {
            return null;
          }
          return new ActiveTextEditorVisibleContent()
              .setContent(editor.getDocument().getText())
              .setFileName(file.getPath());
        });
  }

  @JsonRequest("intent/isCodebaseContextRequired")
  public CompletableFuture<Boolean> intentIsCodebaseContextRequired(String params) {
    // TODO
    return CompletableFuture.completedFuture(false);
  }

  @JsonRequest("intent/isEditorContextRequired")
  public CompletableFuture<Boolean> intentIsEditorContextRequired(String params) {
    // TODO
    return CompletableFuture.completedFuture(false);
  }

  @JsonRequest("editor/replaceSelection")
  public CompletableFuture<ReplaceSelectionResult> editorReplaceSelection(
      ReplaceSelectionParams params) {
    // TODO
    return CompletableFuture.completedFuture(null);
  }

  // =============
  // Notifications
  // =============

  @JsonNotification("editor/warning")
  public void editorWarning(String params) {
    // TODO
  }

  @JsonNotification("chat/updateMessageInProgress")
  public void chatUpdateMessageInProgress(ChatMessage params) {
    if (onChatUpdateMessageInProgress != null && params != null) {
      ApplicationManager.getApplication()
          .invokeLater(() -> onChatUpdateMessageInProgress.accept(params));
    }
  }

  @JsonNotification("chat/updateTranscript")
  public void chatUpdateTranscript(TranscriptJSON params) {
    // TODO
  }
}
