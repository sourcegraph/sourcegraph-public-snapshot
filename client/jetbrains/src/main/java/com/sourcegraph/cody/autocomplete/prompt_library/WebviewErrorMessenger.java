package com.sourcegraph.cody.autocomplete.prompt_library;

import com.intellij.openapi.diagnostic.Logger;

public class WebviewErrorMessenger {
  private static final Logger logger = Logger.getInstance(WebviewErrorMessenger.class);

  public void show(String message) {
    logger.warn("WebviewErrorMessenger: " + message);
  }
}
