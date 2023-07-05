package com.sourcegraph.cody.vscode;

import com.sourcegraph.cody.autocomplete.AutoCompleteDocumentContext;
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

  AutoCompleteDocumentContext getAutoCompleteContext(int offset);

  @NotNull
  Optional<String> getLanguageId();
}
