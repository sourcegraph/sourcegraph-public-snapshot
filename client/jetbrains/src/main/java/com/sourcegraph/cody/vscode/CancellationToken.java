package com.sourcegraph.cody.vscode;

import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;

public class CancellationToken {
  public final CompletableFuture<Boolean> cancelled = new CompletableFuture<>();

  public void onCancellationRequested(Runnable callback) {
    this.cancelled.thenAccept(
        (cancelled) -> {
          if (cancelled) {
            callback.run();
          }
        });
  }

  public boolean isCancelled() {
    try {
      return this.cancelled.isDone() && this.cancelled.get();
    } catch (ExecutionException | InterruptedException ignored) {
      return true;
    }
  }

  public void abort() {
    this.cancelled.complete(true);
  }
}
