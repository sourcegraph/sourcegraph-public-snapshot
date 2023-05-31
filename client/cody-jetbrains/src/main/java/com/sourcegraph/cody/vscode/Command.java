package com.sourcegraph.cody.vscode;

public class Command {
  public final String title;
  public final String command;
  public final String tooltip;

  public Command(String title, String command, String tooltip) {
    this.title = title;
    this.command = command;
    this.tooltip = tooltip;
  }
}
