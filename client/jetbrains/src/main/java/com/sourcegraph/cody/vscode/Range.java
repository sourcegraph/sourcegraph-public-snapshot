package com.sourcegraph.cody.vscode;

public class Range {
  public final Position start;
  public final Position end;

  public Range(Position start, Position end) {
    this.start = start;
    this.end = end;
  }

  @Override
  public String toString() {
    return "Range{" + "start=" + start + ", end=" + end + '}';
  }
}
