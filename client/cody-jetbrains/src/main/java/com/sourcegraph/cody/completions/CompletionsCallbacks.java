package com.sourcegraph.cody.completions;

// Define a callback interface to handle events
public interface CompletionsCallbacks {
  void onSubscribed();

  void onData(String data);

  void onError(Throwable error);

  void onComplete();
}
