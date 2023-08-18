package com.sourcegraph.cody.agent;

import java.io.IOException;

public class CodyAgentException extends Exception {
  public CodyAgentException(String message) {
    super(message);
  }

  public CodyAgentException(String message, IOException e) {
    super(message, e);
  }

  public static CodyAgentException NOT_INITIALIZED =
      new CodyAgentException("Cody agent is not initialized");

  @Override
  public Throwable fillInStackTrace() {
    // don't fill in stack trace
    return this;
  }
}
