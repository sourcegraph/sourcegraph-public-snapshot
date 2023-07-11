package com.sourcegraph.cody.agent;

import com.sourcegraph.cody.agent.protocol.*;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import org.eclipse.lsp4j.jsonrpc.services.JsonNotification;
import org.eclipse.lsp4j.jsonrpc.services.JsonRequest;

/**
 * Interface for the server-part of the Cody agent protocol. The implementation of this interface is
 * written in TypeScript in the file "cody/agent/src/agent.ts". The Eclipse LSP4J bindings create a
 * Java implementation of this interface by using a JVM-reflection feature called "Proxy", which
 * works similar to JavaScript Proxy.
 */
public interface CodyAgentServer {

  // Requests
  @JsonRequest("initialize")
  CompletableFuture<ServerInfo> initialize(ClientInfo clientInfo);

  @JsonRequest("shutdown")
  CompletableFuture<Void> shutdown();

  @JsonRequest("recipes/list")
  CompletableFuture<List<RecipeInfo>> recipesList();

  @JsonRequest("recipes/execute")
  CompletableFuture<Void> recipesExecute(ExecuteRecipeParams params);

  // Notifications
  @JsonNotification("initialized")
  void initialized();

  @JsonNotification("exit")
  void exit();

  @JsonNotification("connectionConfiguration/didChange")
  void configurationDidChange(ConnectionConfiguration document);

  @JsonNotification("textDocument/didFocus")
  void textDocumentDidFocus(TextDocument document);

  @JsonNotification("textDocument/didOpen")
  void textDocumentDidOpen(TextDocument document);

  @JsonNotification("textDocument/didChange")
  void textDocumentDidChange(TextDocument document);

  @JsonNotification("textDocument/didClose")
  void textDocumentDidClose(TextDocument document);
}
