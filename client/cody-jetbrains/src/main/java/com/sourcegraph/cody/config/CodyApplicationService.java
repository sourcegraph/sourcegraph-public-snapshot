package com.sourcegraph.cody.config;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.components.PersistentStateComponent;
import com.intellij.openapi.components.State;
import com.intellij.openapi.components.Storage;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

@State(
    name = "ApplicationConfig",
    storages = {@Storage("cody.xml")})
public class CodyApplicationService
    implements CodyService, PersistentStateComponent<CodyApplicationService> {
  @Nullable private String instanceType;
  @Nullable private String dotcomAccessToken;
  @Nullable private String enterpriseUrl;
  @Nullable private String enterpriseAccessToken;
  @Nullable private String customRequestHeaders;
  @Nullable private String codebase;
  @Nullable private String anonymousUserId;
  private boolean isInstallEventLogged;
  private boolean areChatPredictionsEnabled;
  @Nullable private Boolean authenticationFailedLastTime;

  @NotNull
  public static CodyApplicationService getInstance() {
    return ApplicationManager.getApplication().getService(CodyApplicationService.class);
  }

  @Nullable
  public String getInstanceType() {
    return instanceType;
  }

  public void setInstanceType(@Nullable String instanceType) {
    this.instanceType = instanceType;
  }

  @Nullable
  public String getDotcomAccessToken() {
    return dotcomAccessToken;
  }

  public void setDotcomAccessToken(@Nullable String dotcomAccessToken) {
    this.dotcomAccessToken = dotcomAccessToken;
  }

  @Nullable
  public String getEnterpriseUrl() {
    return enterpriseUrl;
  }

  public void setEnterpriseUrl(@Nullable String enterpriseUrl) {
    this.enterpriseUrl = enterpriseUrl;
  }

  @Nullable
  public String getEnterpriseAccessToken() {
    return enterpriseAccessToken;
  }

  public void setEnterpriseAccessToken(@Nullable String enterpriseAccessToken) {
    this.enterpriseAccessToken = enterpriseAccessToken;
  }

  @Nullable
  public String getCustomRequestHeaders() {
    return customRequestHeaders;
  }

  public void setCustomRequestHeaders(@Nullable String customRequestHeaders) {
    this.customRequestHeaders = customRequestHeaders;
  }

  @Nullable
  public String getCodebase() {
    return codebase;
  }

  public void setCodebase(@Nullable String codebase) {
    this.codebase = codebase;
  }

  @Nullable
  public String getAnonymousUserId() {
    return anonymousUserId;
  }

  public void setAnonymousUserId(@Nullable String anonymousUserId) {
    this.anonymousUserId = anonymousUserId;
  }

  public boolean isInstallEventLogged() {
    return isInstallEventLogged;
  }

  public void setInstallEventLogged(boolean installEventLogged) {
    isInstallEventLogged = installEventLogged;
  }

  public Boolean areChatPredictionsEnabled() {
    return areChatPredictionsEnabled;
  }

  public void setChatPredictionsEnabled(Boolean areChatPredictionsEnabled) {
    this.areChatPredictionsEnabled = areChatPredictionsEnabled;
  }

  @Nullable
  public Boolean getAuthenticationFailedLastTime() {
    return authenticationFailedLastTime;
  }

  public void setAuthenticationFailedLastTime(@Nullable Boolean authenticationFailedLastTime) {
    this.authenticationFailedLastTime = authenticationFailedLastTime;
  }

  @Nullable
  @Override
  public CodyApplicationService getState() {
    return this;
  }

  @Override
  public void loadState(@NotNull CodyApplicationService settings) {
    this.instanceType = settings.instanceType;
    this.enterpriseUrl = settings.enterpriseUrl;
    this.enterpriseAccessToken = settings.enterpriseAccessToken;
    this.customRequestHeaders = settings.customRequestHeaders;
    this.codebase = settings.codebase;
    this.anonymousUserId = settings.anonymousUserId;
    this.areChatPredictionsEnabled = settings.areChatPredictionsEnabled;
    this.authenticationFailedLastTime = settings.authenticationFailedLastTime;
  }
}
