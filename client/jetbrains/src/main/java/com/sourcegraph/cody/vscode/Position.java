package com.sourcegraph.cody.vscode;

public class Position {
  public final int line;
  public final int character;

  public Position(int line, int character) {
    this.line = line;
    this.character = character;
  }

  @Override
  public String toString() {
    return "Position{" + "line=" + line + ", character=" + character + '}';
  }
}
