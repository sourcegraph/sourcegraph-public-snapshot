package com.sourcegraph.cody;

public class UpdatableChatHolderService {
  private UpdatableChat updatableChat;

  public UpdatableChat getUpdatableChat() {
    return updatableChat;
  }

  public void setUpdatableChat(UpdatableChat updatableChat) {
    this.updatableChat = updatableChat;
  }
}
