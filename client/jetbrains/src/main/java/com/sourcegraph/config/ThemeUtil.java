package com.sourcegraph.config;

import com.google.gson.JsonObject;
import com.intellij.util.ui.UIUtil;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import javax.swing.*;
import java.awt.*;

public class ThemeUtil {
    @NotNull
    public static JsonObject getCurrentThemeAsJson() {
        // Find the name of properties here: https://plugins.jetbrains.com/docs/intellij/themes-metadata.html#key-naming-scheme
        JsonObject theme = new JsonObject();
        theme.addProperty("isDarkTheme", isDarkTheme());
        theme.addProperty("backgroundColor", getHexString(UIUtil.getPanelBackground()));
        theme.addProperty("buttonArc", UIManager.get("Button.arc").toString());
        theme.addProperty("buttonColor", getHexString(UIManager.getColor("Button.default.background")));
        theme.addProperty("color", getHexString(UIUtil.getLabelForeground()));
        theme.addProperty("font", UIUtil.getLabelFont().getFontName());
        theme.addProperty("fontSize", UIUtil.getLabelFont().getSize());
        theme.addProperty("labelBackground", getHexString(UIManager.getColor("Label.background")));
        return theme;
    }

    @NotNull
    public static String getPanelBackgroundColorHexString() {
        return getHexString(UIUtil.getPanelBackground());
    }

    public static boolean isDarkTheme() {
        return getBrightnessFromColor(UIUtil.getPanelBackground()) < 128;
    }

    @Nullable
    private static String getHexString(@Nullable Color color) {
        if (color != null) {
            return "#" + Integer.toHexString(color.getRGB()).substring(2);
        } else {
            return null;
        }
    }

    /**
     * Calculates the brightness between 0 (dark) and 255 (bright) from the given color.
     * Source: <a href="https://alienryderflex.com/hsp.html">https://alienryderflex.com/hsp.html</a>
     */
    private static int getBrightnessFromColor(@NotNull Color color) {
        return (int) Math.sqrt(color.getRed() * color.getRed() * .299 + color.getGreen() * color.getGreen() * .587 + color.getBlue() * color.getBlue() * .114);
    }
}
