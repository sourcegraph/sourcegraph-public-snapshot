package com.sourcegraph.agent.protocol;

import java.util.List;

public class TranscriptJSON {
  public String id;
  public List<InteractionJSON> interactions;
  public String lastInteractionTimestamp;
}
