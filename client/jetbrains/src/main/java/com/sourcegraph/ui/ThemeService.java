package com.sourcegraph.ui;

import com.intellij.util.ui.UIUtil;

import java.awt.*;

public class ThemeService {
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
