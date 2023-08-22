package com.sourcegraph.cody;

import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.ex.FocusChangeListener;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.CodyAgentClient;
import com.sourcegraph.cody.agent.protocol.TextDocument;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;

public class CodyAgentFocusListener implements FocusChangeListener {
  @Override
  public void focusGained(@NotNull Editor editor) {
    if (!ConfigUtil.isCodyEnabled()) {
      return;
    }
    if (editor.getProject() == null) {
      return;
    }
    VirtualFile file = FileDocumentManager.getInstance().getFile(editor.getDocument());
    if (file == null) {
      return;
    }
    if (!CodyAgent.isConnected(editor.getProject())) {
      return;
    }
    CodyAgentClient client = CodyAgent.getClient(editor.getProject());
    if (client.server == null) {
      return;
    }
    client.server.textDocumentDidFocus(new TextDocument().setFilePath(file.getPath()));

    if (client.codebase == null) {
      return;
    }
    client.codebase.handlePotentialCodebaseChange(editor.getProject(), file);
  }
}
