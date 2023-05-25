package com.sourcegraph.cody.prompts;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.util.HashMap;
import java.util.Map;

public class LanguageUtils {

    private static final @NotNull Map<String, String> EXTENSION_TO_LANGUAGE;

    static {
        EXTENSION_TO_LANGUAGE = new HashMap<>();
        EXTENSION_TO_LANGUAGE.put("py", "Python");
        EXTENSION_TO_LANGUAGE.put("rb", "Ruby");
        EXTENSION_TO_LANGUAGE.put("md", "Markdown");
        EXTENSION_TO_LANGUAGE.put("php", "PHP");
        EXTENSION_TO_LANGUAGE.put("js", "Javascript");
        EXTENSION_TO_LANGUAGE.put("ts", "Typescript");
        EXTENSION_TO_LANGUAGE.put("jsx", "JSX");
        EXTENSION_TO_LANGUAGE.put("tsx", "TSX");
    }

    public static @NotNull String getNormalizedLanguageName(@Nullable String extension) {
        if (extension == null || extension.isEmpty()) {
            return "";
        }

        String language = EXTENSION_TO_LANGUAGE.get(extension);
        if (language == null) {
            return extension.substring(0, 1).toUpperCase() + extension.substring(1);
        }

        return language;
    }

    public static boolean isMarkdownFile(@NotNull String filePath) {
        String extension = getExtension(filePath);
        return extension.equals("md") || extension.equals("markdown");
    }

    public static @NotNull String getExtension(@NotNull String filePath) {
        int lastDotIndex = filePath.lastIndexOf('.');
        if (lastDotIndex == -1) {
            return "";
        }
        return filePath.substring(lastDotIndex + 1);
    }
}
