package com.sourcegraph.ui;

import com.google.gson.JsonObject;
import com.intellij.util.ui.UIUtil;

import javax.swing.*;
import java.awt.*;

public class ThemeService {
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

    public static String getPanelBackgroundColorHexString() {
        return getHexString(UIUtil.getPanelBackground());
    }

    public static boolean isDarkTheme() {
        return getBrightnessFromColor(UIUtil.getPanelBackground()) < 128;
    }

    private static String getHexString(Color color) {
        return "#" + Integer.toHexString(color.getRGB()).substring(2);
    }

    /**
     * Calculates the brightness between 0 (dark) and 255 (bright) from the given color.
     * Source: <a href="https://alienryderflex.com/hsp.html">https://alienryderflex.com/hsp.html</a>
     */
    private static int getBrightnessFromColor(Color color) {
        return (int) Math.sqrt(color.getRed() * color.getRed() * .299 + color.getGreen() * color.getGreen() * .587 + color.getBlue() * color.getBlue() * .114);
    }
}
