package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.cody.vscode.TextDocument;
import java.util.HashMap;
import java.util.Map;
import org.apache.commons.lang.StringUtils;
import org.jetbrains.annotations.NotNull;

public class UnstableCodegenLanguageUtil {
  private static final Logger logger = Logger.getInstance(UnstableCodegenLanguageUtil.class);
  private static final Map<String, String> fileExtensionToModelLanguageId =
      new HashMap<>() {
        {
          put(".js", "javascript");
          put(".jsx", "javascript");
          put(".java", "java");
          put(".py", "python");
          put(".cpp", "cpp");
          put(".apex", "apex");
          put(".ts", "typescript");
          put(".kt", "kotlin");
          put(".cls", "apex");
          put(".scss", "css");
          put(".css", "css");
          put(".html", "html");
          put(".cs", "c-sharp");
          put(".php", "php");
          put(".sql", "sql");
          put(".rs", "rust");
          put(".rlib", "rust");
          put(".vue", "vue");
          put(".rb", "ruby");
          put(".swift", "swift");
          put(".dart", "dart");
          put(".sh", "shell");
          put(".lua", "lua");
          put(".go", "go");
        }
      };

  @NotNull
  public static String getModelLanguageId(@NotNull TextDocument textDocument) {
    String fileExtension = "." + StringUtils.substringAfterLast(textDocument.fileName(), ".");
    String languageIdBasedOnExtension =
        fileExtensionToModelLanguageId.getOrDefault(fileExtension, "");
    String languageIdBasedOnIntelliJ =
        textDocument.getLanguageId().map(String::toLowerCase).orElse("");
    boolean intelliJLanguageIdIsSupported =
        fileExtensionToModelLanguageId.containsValue(languageIdBasedOnIntelliJ);
    if (!languageIdBasedOnExtension.isEmpty()
        && !languageIdBasedOnIntelliJ.isEmpty()
        && !languageIdBasedOnExtension.equals(languageIdBasedOnIntelliJ)) {
      logger.warn( // logging the mismatch to make debugging easier
          "Cody: AutoComplete: Detected mismatch between the code language detected by IntelliJ vs based on extension. "
              + "IntelliJ detected: "
              + languageIdBasedOnIntelliJ
              + " and extension-based detection: "
              + languageIdBasedOnExtension);
      if (intelliJLanguageIdIsSupported) {
        logger.warn(
            "Cody: IntelliJ detected language is supported by `unstable-codegen`, so it takes priority.");
      }
    }
    String fileExtensionFallback = StringUtils.stripStart(fileExtension, ".");
    if (intelliJLanguageIdIsSupported) {
      // if the languageId returned from IntelliJ is supported, that takes priority
      return languageIdBasedOnIntelliJ;
    } else if (!languageIdBasedOnExtension.isEmpty()) {
      // otherwise, if a supported language based on extension was found, we pick that
      return languageIdBasedOnExtension;
    } else if (!languageIdBasedOnIntelliJ.isEmpty()) {
      // otherwise, we pick whatever we got from IntelliJ anyway, if it's not empty
      return languageIdBasedOnIntelliJ;
    } else if (!fileExtensionFallback.isEmpty()) {
      // otherwise, we try to return the file extension (with no leading dot)
      return fileExtensionFallback;
    } else {
      // if the file doesn't even have an extension, we've run out of options
      return "no-known-extension-detected";
    }
  }
}
