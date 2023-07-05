package com.sourcegraph.cody.vscode;

import com.sourcegraph.cody.completions.CompletionDocumentContext;
import java.net.URI;
import java.util.Optional;
import org.jetbrains.annotations.NotNull;

public interface TextDocument {
  URI uri();

  @NotNull
  String fileName();

  int offsetAt(Position position);

  String getText();

  String getText(Range range);

  Position positionAt(int offset);

  CompletionDocumentContext getCompletionContext(int offset);

  @NotNull
  Optional<String> getLanguageId();
}
