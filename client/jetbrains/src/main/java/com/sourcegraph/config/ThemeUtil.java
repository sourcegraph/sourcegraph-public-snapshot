package com.sourcegraph.config;

import com.google.gson.JsonObject;
import com.intellij.lang.java.JavaLanguage;
import com.intellij.openapi.editor.colors.*;
import com.intellij.util.ui.UIUtil;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.swing.*;
import javax.swing.plaf.ColorUIResource;
import java.awt.*;
import java.util.*;

public class ThemeUtil {
    private static final Logger logger = LoggerFactory.getLogger(ThemeUtil.class);

    @NotNull
    public static JsonObject getCurrentThemeAsJson() {
        JsonObject intelliJTheme = new JsonObject();
        UIDefaults defaults = UIManager.getDefaults();
        Enumeration<Object> keysEnumeration = defaults.keys();
        ArrayList<Object> keysList = Collections.list(keysEnumeration);
        for (Object key : keysList) {
            try {
                Object value = UIManager.get(key);
                if (value instanceof ColorUIResource) {
                    intelliJTheme.addProperty(key.toString(), getHexString(UIManager.getColor(key)));
                }
            } catch (Exception e) {
                logger.error(e.getMessage());
            }
        }

        JsonObject syntaxTheme = new JsonObject();
        try {
            EditorColorsScheme colorScheme = EditorColorsManager.getInstance().getGlobalScheme();

            EditorColorPalette palette = EditorColorPaletteFactory.getInstance().getPalette(colorScheme, JavaLanguage.INSTANCE);
            for (Map.Entry<Color, Collection<TextAttributesKey>> entry : palette.withForegroundColors().getEntries()) {
                Color color = entry.getKey();
                for (TextAttributesKey key : entry.getValue()) {
                    recursivelyAddToAllAttributeKeys(syntaxTheme, getHexString(color), key);
                }
            }
        } catch (Exception e) {
            logger.error(e.getMessage());
        }

        JsonObject theme = new JsonObject();
        theme.addProperty("isDarkTheme", isDarkTheme());
        theme.add("intelliJTheme", intelliJTheme);
        theme.add("syntaxTheme", syntaxTheme);
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
            String colorString = Integer.toHexString(color.getRGB());
            if (colorString.length() > 2) {
                return "#" + colorString.substring(2);
            } else {
                return "#000000";
            }
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

    private static void recursivelyAddToAllAttributeKeys(JsonObject object, String value, TextAttributesKey key) {
        if (key == null) {
            return;
        }
        object.addProperty(key.getExternalName(), value);
        recursivelyAddToAllAttributeKeys(object, value, key.getFallbackAttributeKey());
    }
}
