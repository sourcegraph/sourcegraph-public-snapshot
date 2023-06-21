package com.sourcegraph.cody.agent;

import com.sourcegraph.cody.agent.protocol.TextDocument;
import java.util.HashMap;
import java.util.Map;

public class CodyAgentDocuments {
  private final CodyAgentServer underlying;
  private String focusedPath = null;
  private Map<String, TextDocument> documents = new HashMap<>();

  public CodyAgentDocuments(CodyAgentServer underlying) {
    this.underlying = underlying;
  }

  private void handleDocument(TextDocument document) {
    TextDocument old = this.documents.get(document.filePath);
    if (old == null) {
      this.documents.put(document.filePath, document);
      return;
    }
    if (document.content == null) {
      document.content = old.content;
    }
    if (document.selection == null) {
      document.selection = old.selection;
    }
    this.documents.put(document.filePath, document);
  }

  public void didOpen(TextDocument document) {
    this.documents.put(document.filePath, document);
    underlying.textDocumentDidOpen(document);
  }

  public void didFocus(TextDocument document) {
    this.documents.put(document.filePath, document);
  }

  public void didChange(TextDocument document) {}

  public void didClose(TextDocument document) {}
}
