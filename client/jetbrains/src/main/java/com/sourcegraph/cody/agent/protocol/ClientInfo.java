package com.sourcegraph.cody.agent.protocol;

import com.sourcegraph.cody.agent.ExtensionConfiguration;
import org.jetbrains.annotations.Nullable;

public class ClientInfo {

  public String name;
  public String version;
  public String workspaceRootPath;
  @Nullable public ExtensionConfiguration extensionConfiguration;

  public ClientInfo setName(String name) {
    this.name = name;
    return this;
  }

  public ClientInfo setVersion(String version) {
    this.version = version;
    return this;
  }

  public ClientInfo setWorkspaceRootPath(String workspaceRootPath) {
    this.workspaceRootPath = workspaceRootPath;
    return this;
  }

  public ClientInfo setExtensionConfiguration(ExtensionConfiguration extensionConfiguration) {
    this.extensionConfiguration = extensionConfiguration;
    return this;
  }
}
