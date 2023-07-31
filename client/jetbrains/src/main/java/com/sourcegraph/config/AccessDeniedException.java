package com.sourcegraph.config;

public class AccessDeniedException extends Exception {
  public AccessDeniedException(String message) {
    super(message);
  }
}
