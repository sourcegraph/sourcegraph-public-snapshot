package com.sourcegraph.cody.vscode;

import com.google.common.collect.Iterables;
import com.sourcegraph.cody.autocomplete.AutocompleteDocumentContext;
import java.net.URI;
import java.util.Optional;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class TestTextDocument implements TextDocument {
  @NotNull private final URI uri;
  @NotNull private final String fileName;
  @NotNull private final String text;

  @Nullable private final String languageId;

  public TestTextDocument(
      @NotNull URI uri,
      @NotNull String fileName,
      @NotNull String text,
      @Nullable String languageId) {
    this.uri = uri;
    this.fileName = fileName;
    this.text = text;
    this.languageId = languageId;
  }

  @Override
  public URI uri() {
    return this.uri;
  }

  @Override
  public @NotNull String fileName() {
    return this.fileName;
  }

  @Override
  public int offsetAt(Position position) {
    return 0;
  }

  @Override
  public @NotNull String getText() {
    return this.text;
  }

  @Override
  public String getText(Range range) {
    return null;
  }

  @Override
  public Position positionAt(int offset) {
    return new Position(0, 0);
  }

  @Override
  public AutocompleteDocumentContext getAutocompleteContext(int offset) {
    String sameLinePrefix =
        Iterables.getLast(this.text.substring(0, offset).lines().collect(Collectors.toList()));
    String sameLineSuffix =
        Iterables.getFirst(this.text.substring(offset).lines().collect(Collectors.toList()), "");
    assert sameLineSuffix != null;
    return new AutocompleteDocumentContext(sameLinePrefix, sameLineSuffix);
  }

  @Override
  public @NotNull Optional<String> getLanguageId() {
    return Optional.ofNullable(this.languageId);
  }
}
