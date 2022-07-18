package com.sourcegraph.config;

import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
    @Nullable
    public String oldUrl;

    @Nullable
    public String oldAccessToken;

    @Nullable
    public String newUrl;

    @Nullable
    public String newAccessToken;

    public PluginSettingChangeContext(@Nullable String oldUrl, @Nullable String oldAccessToken, @Nullable String newUrl, @Nullable String newAccessToken) {
        this.oldUrl = oldUrl;
        this.oldAccessToken = oldAccessToken;
        this.newUrl = newUrl;
        this.newAccessToken = newAccessToken;
    }
}
