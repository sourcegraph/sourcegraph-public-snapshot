package com.sourcegraph.cody.config;

import com.intellij.openapi.components.PersistentStateComponent;
import com.intellij.openapi.components.State;
import com.intellij.openapi.components.Storage;
import com.intellij.openapi.project.Project;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

@State(
    name = "Config",
    storages = {@Storage("cody.xml")})
public class CodyProjectService implements CodyService, PersistentStateComponent<CodyProjectService> {
    @Nullable
    private String instanceType;
    @Nullable
    private String dotcomAccessToken;
    @Nullable
    private String enterpriseUrl;
    @Nullable
    private String enterpriseAccessToken;
    @Nullable
    private String customRequestHeaders;
    @Nullable
    private String defaultBranch;
    @Nullable
    private String codebase;
    @Nullable
    private Boolean areChatPredictionsEnabled;

    @NotNull
    public static CodyProjectService getInstance(@NotNull Project project) {
        return project.getService(CodyProjectService.class);
    }

    public void setInstanceType(@Nullable String instanceType) {
        this.instanceType = instanceType;
    }

    @Nullable
    public String getInstanceType() {
        return instanceType;
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
    @Override
    public CodyProjectService getState() {
        return this;
    }

    @Nullable
    public Boolean areChatPredictionsEnabled() {
        return areChatPredictionsEnabled;
    }

    public void setChatPredictionsEnabled(@Nullable Boolean areChatPredictionsEnabled) {
        this.areChatPredictionsEnabled = areChatPredictionsEnabled;
    }

    @Override
    public void loadState(@NotNull CodyProjectService settings) {
        this.instanceType = settings.instanceType;
        this.dotcomAccessToken = settings.dotcomAccessToken;
        this.enterpriseUrl = settings.enterpriseUrl;
        this.enterpriseAccessToken = settings.enterpriseAccessToken;
        this.customRequestHeaders = settings.customRequestHeaders;
        this.defaultBranch = settings.defaultBranch;
        this.codebase = settings.codebase;
        this.areChatPredictionsEnabled = settings.areChatPredictionsEnabled;
    }
}
