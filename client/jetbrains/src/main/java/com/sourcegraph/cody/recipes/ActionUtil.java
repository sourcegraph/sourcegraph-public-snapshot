package com.sourcegraph.cody.recipes;

import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.editor.EditorContext;
import com.sourcegraph.cody.editor.EditorContextGetter;
import java.util.function.Consumer;
import org.jetbrains.annotations.NotNull;

public class ActionUtil {

  public static void runIfCodeSelected(
      @NotNull UpdatableChat updatableChat,
      @NotNull Project project,
      @NotNull Consumer<String> runIfCodeSelected) {
    EditorContext editorContext = EditorContextGetter.getEditorContext(project);
    String editorSelection = editorContext.getSelection();
    if (editorSelection == null) {
      updatableChat.activateChatTab();
      updatableChat.addMessageToChat(
          ChatMessage.createAssistantMessage(
              "No code selected. Please select some code and try again."));
      return;
    }
    runIfCodeSelected.accept(editorSelection);
  }
}
