package com.sourcegraph.cody.vscode;

import com.sourcegraph.cody.completions.CompletionDocumentContext;
import java.net.URI;
import org.jetbrains.annotations.Nullable;

public interface TextDocument {
  URI uri();

  String fileName();

  int offsetAt(Position position);

  String getText();

  String getText(Range range);

  Position positionAt(int offset);

  CompletionDocumentContext getCompletionContext(int offset);

  @Nullable
  String getLanguageId();
}
