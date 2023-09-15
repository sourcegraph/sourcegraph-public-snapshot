package com.sourcegraph.cody;

import com.intellij.openapi.editor.Document;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.fileEditor.FileEditorManagerListener;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.CodyAgentClient;
import com.sourcegraph.cody.agent.CodyAgentServer;
import com.sourcegraph.cody.agent.protocol.TextDocument;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;

public class CodyFileEditorListener implements FileEditorManagerListener {
  @Override
  public void fileOpened(@NotNull FileEditorManager source, @NotNull VirtualFile file) {
    if (!ConfigUtil.isCodyEnabled()) {
      return;
    }
    Document document = FileDocumentManager.getInstance().getDocument(file);
    if (document == null) {
      return;
    }
    if (!CodyAgent.isConnected(source.getProject())) {
      return;
    }
    CodyAgentClient client = CodyAgent.getClient(source.getProject());
    if (client.server == null) {
      return;
    }
    client.server.textDocumentDidOpen(
        new TextDocument().setFilePath(file.getPath()).setContent(document.getText()));

    if (client.codebase == null) {
      return;
    }
    client.codebase.onFileOpened(source.getProject(), file);
  }

  @Override
  public void fileClosed(@NotNull FileEditorManager source, @NotNull VirtualFile file) {
    if (!ConfigUtil.isCodyEnabled()) {
      return;
    }
    CodyAgentServer server = CodyAgent.getServer(source.getProject());
    if (server == null) {
      return;
    }
    server.textDocumentDidClose(new TextDocument().setFilePath(file.getPath()));
  }
}
