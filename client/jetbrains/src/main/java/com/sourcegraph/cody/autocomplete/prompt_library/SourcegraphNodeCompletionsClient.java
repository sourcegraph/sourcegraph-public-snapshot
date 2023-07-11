package com.sourcegraph.cody.autocomplete.prompt_library;

import com.google.gson.Gson;
import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.cody.api.CompletionsCallbacks;
import com.sourcegraph.cody.api.CompletionsInput;
import com.sourcegraph.cody.api.CompletionsService;
import com.sourcegraph.cody.vscode.CancellationToken;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CompletableFuture;

public class SourcegraphNodeCompletionsClient {
  private static final Logger logger = Logger.getInstance(SourcegraphNodeCompletionsClient.class);
  public final CompletionsService completionsService;
  private final CancellationToken token;

  public SourcegraphNodeCompletionsClient(
      CompletionsService completionsService, CancellationToken token) {
    this.completionsService = completionsService;
    this.token = token;
  }

  public CompletableFuture<CompletionResponse> complete(CompletionParameters params) {
    CodeCompletionCallbacks callbacks = new CodeCompletionCallbacks(token);
    //    logger.info(
    //        "QUERY: " +
    // params.messages.stream().map(Message::prompt).collect(Collectors.joining("")));
    completionsService.streamCompletion(
        new CompletionsInput(
            params.messages,
            params.temperature,
            params.stopSequences,
            params.maxTokensToSample,
            params.topK,
            params.topP),
        callbacks,
        CompletionsService.Endpoint.Code);
    return callbacks.promise;
  }

  private static class CodeCompletionCallbacks implements CompletionsCallbacks {
    private final CancellationToken token;
    CompletableFuture<CompletionResponse> promise = new CompletableFuture<>();
    List<String> chunks = new ArrayList<>();

    private CodeCompletionCallbacks(CancellationToken token) {
      this.token = token;
    }

    @Override
    public void onSubscribed() {
      // Do nothing
    }

    @Override
    public void onData(String data) {
      chunks.add(data);
    }

    @Override
    public void onError(Throwable error) {
      promise.completeExceptionally(error);
      logger.error(error);
    }

    @Override
    public void onComplete() {
      String json = String.join("", chunks);
      CompletionResponse completionResponse = new Gson().fromJson(json, CompletionResponse.class);
      promise.complete(completionResponse);
    }

    @Override
    public void onCancelled() {
      this.token.abort();
    }
  }
}
