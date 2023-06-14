package com.sourcegraph.agent.protocol;

import java.util.List;

public class InteractionJSON {
  public InteractionMessage humanMessage;
  public InteractionMessage assistantMessage;
  public List<ContextMessage> context;
  public String timestamp;
}
