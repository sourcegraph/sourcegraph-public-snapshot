package com.sourcegraph.cody;

public class InvalidAccessTokenException extends RuntimeException {
  public InvalidAccessTokenException(String message, Throwable cause) {
    super(message, cause);
  }
}
