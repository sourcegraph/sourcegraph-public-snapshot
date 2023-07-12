package com.sourcegraph.cody;

import org.jetbrains.annotations.Nullable;

public class UpdatableChatHolderService {
  private @Nullable UpdatableChat updatableChat;

  public @Nullable UpdatableChat getUpdatableChat() {
    return updatableChat;
  }

  public void setUpdatableChat(@Nullable UpdatableChat updatableChat) {
    this.updatableChat = updatableChat;
  }
}
