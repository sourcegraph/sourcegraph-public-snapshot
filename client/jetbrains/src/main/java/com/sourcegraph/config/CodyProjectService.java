package com.sourcegraph.config;

import com.intellij.openapi.components.PersistentStateComponent;
import com.intellij.openapi.components.State;
import com.intellij.openapi.components.Storage;
import com.intellij.openapi.project.Project;
import com.sourcegraph.find.Search;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

@State(
    name = "Config",
    storages = {@Storage("sourcegraph.xml")})
public class CodyProjectService implements PersistentStateComponent<CodyProjectService> {

  @NotNull
  static CodyProjectService getInstance(@NotNull Project project) {
    return project.getService(CodyProjectService.class);
  }

  @Nullable public String instanceType;
  @Nullable public String url;
  @Nullable public String dotComAccessToken;
  @Nullable public String enterpriseAccessToken;
  @Nullable public String customRequestHeaders;
  @Nullable public String defaultBranch;
  @Nullable public String remoteUrlReplacements;
  @Nullable public String lastSearchQuery;
  public boolean lastSearchCaseSensitive;
  @Nullable public String lastSearchPatternType;
  @Nullable public String lastSearchContextSpec;

  @Nullable
  public String getInstanceType() {
    return instanceType;
  }

  @Nullable
  public String getSourcegraphUrl() {
    return url;
  }

  @Nullable
  public String getDotComAccessToken() {
    return dotComAccessToken;
  }

  @Nullable
  public String getCustomRequestHeaders() {
    return customRequestHeaders;
  }

  @Nullable
  public String getDefaultBranchName() {
    return defaultBranch;
  }

  @Nullable
  public String getRemoteUrlReplacements() {
    return remoteUrlReplacements;
  }

  @Nullable
  public Search getLastSearch() {
    if (lastSearchQuery == null) {
      return null;
    } else {
      return new Search(
          lastSearchQuery, lastSearchCaseSensitive, lastSearchPatternType, lastSearchContextSpec);
    }
  }

  @Nullable
  @Override
  public CodyProjectService getState() {
    return this;
  }

  @Override
  public void loadState(@NotNull CodyProjectService settings) {
    this.instanceType = settings.instanceType;
    this.url = settings.url;
    this.dotComAccessToken = settings.dotComAccessToken;
    this.enterpriseAccessToken = settings.enterpriseAccessToken;
    this.customRequestHeaders = settings.customRequestHeaders;
    this.defaultBranch = settings.defaultBranch;
    this.remoteUrlReplacements = settings.remoteUrlReplacements;
    this.lastSearchQuery = settings.lastSearchQuery != null ? settings.lastSearchQuery : "";
    this.lastSearchCaseSensitive = settings.lastSearchCaseSensitive;
    this.lastSearchPatternType =
        settings.lastSearchPatternType != null ? settings.lastSearchPatternType : "literal";
    this.lastSearchContextSpec =
        settings.lastSearchContextSpec != null ? settings.lastSearchContextSpec : "global";
  }

  @Nullable
  public String getEnterpriseAccessToken() {
    return enterpriseAccessToken;
  }

  public boolean areChatPredictionsEnabled() {
    // TODO: implement
    return false;
  }

  public String getCodebase() {
    // TODO: implement
    return null;
  }
}
