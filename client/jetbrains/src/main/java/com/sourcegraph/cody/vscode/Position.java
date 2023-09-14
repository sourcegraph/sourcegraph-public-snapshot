package com.sourcegraph.cody.vscode;

import java.util.Objects;

public class Position {
  public final int line;
  public final int character;

  public Position(int line, int character) {
    this.line = line;
    this.character = character;
  }

  public Position withLine(int newLine) {
    return new Position(newLine, this.character);
  }

  public Position withCharacter(int newCharacter) {
    return new Position(this.line, newCharacter);
  }

  @Override
  public String toString() {
    return "Position{" + "line=" + line + ", character=" + character + '}';
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (!(o instanceof Position)) return false;
    Position position = (Position) o;
    return line == position.line && character == position.character;
  }

  @Override
  public int hashCode() {
    return Objects.hash(line, character);
  }
}
