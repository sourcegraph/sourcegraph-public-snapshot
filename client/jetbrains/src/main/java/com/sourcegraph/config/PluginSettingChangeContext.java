package com.sourcegraph.config;

import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
    @Nullable
    final public String oldUrl;

    @Nullable
    final public String oldAccessToken;

    @Nullable
    final public String newUrl;

    @Nullable
    final public String newAccessToken;

    public PluginSettingChangeContext(@Nullable String oldUrl, @Nullable String oldAccessToken, @Nullable String newUrl, @Nullable String newAccessToken) {
        this.oldUrl = oldUrl;
        this.oldAccessToken = oldAccessToken;
        this.newUrl = newUrl;
        this.newAccessToken = newAccessToken;
    }
}
