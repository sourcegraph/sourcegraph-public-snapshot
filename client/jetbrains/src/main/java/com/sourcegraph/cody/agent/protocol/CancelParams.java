package com.sourcegraph.cody.agent.protocol;

public class CancelParams {
  public String id;

  public CancelParams() {}

  public CancelParams(String id) {
    this.id = id;
  }
}
