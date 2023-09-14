package com.sourcegraph.cody.agent.protocol;

public class Position {
  public int line;
  public int character;

  public Position setLine(int line) {
    this.line = line;
    return this;
  }

  public Position setCharacter(int character) {
    this.character = character;
    return this;
  }
}
