package com.sourcegraph.cody.vscode;

import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;

public class CancellationToken {
  private final CompletableFuture<Boolean> cancelled = new CompletableFuture<>();

  public void onCancellationRequested(Runnable callback) {
    this.cancelled.thenAccept(
        (cancelled) -> {
          if (cancelled) {
            try {
              callback.run();
            } catch (Exception ignored) {
              // Do nothing about exceptions in cancelation callbacks
            }
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

  public void dispose() {
    this.cancelled.complete(false);
  }

  public void abort() {
    this.cancelled.complete(true);
  }
}
