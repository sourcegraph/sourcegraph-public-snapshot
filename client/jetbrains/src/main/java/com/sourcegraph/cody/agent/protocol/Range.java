package com.sourcegraph.cody.agent.protocol;

public class Range {
  public Position start;
  public Position end;

  public Range setStart(Position start) {
    this.start = start;
    return this;
  }

  public Range setEnd(Position end) {
    this.end = end;
    return this;
  }
}
