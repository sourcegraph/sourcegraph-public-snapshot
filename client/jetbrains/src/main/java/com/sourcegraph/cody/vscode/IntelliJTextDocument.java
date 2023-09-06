package com.sourcegraph.cody.vscode;

import com.intellij.lang.Language;
import com.intellij.lang.LanguageUtil;
import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.TextRange;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.cody.autocomplete.AutocompleteDocumentContext;
import java.net.URI;
import java.nio.file.Paths;
import java.util.Optional;
import org.jetbrains.annotations.NotNull;

/** Implementation of vscode.TextDocument backed by IntelliJ's Editor. */
public class IntelliJTextDocument implements TextDocument {
  public final Editor editor;
  public VirtualFile file;
  public Language language;

  public IntelliJTextDocument(Editor editor, Project project) {
    this.editor = editor;
    Document document = editor.getDocument();
    this.file = FileDocumentManager.getInstance().getFile(document);
    this.language = LanguageUtil.getLanguageForPsi(project, file);
  }

  @Override
  public URI uri() {
    return Paths.get(file.getPath()).toUri();
  }

  @Override
  @NotNull
  public String fileName() {
    return file.getName();
  }

  @Override
  public int offsetAt(Position position) {
    return this.editor.getDocument().getLineStartOffset(position.line) + position.character;
  }

  @Override
  public String getText() {
    return this.editor.getDocument().getText();
  }

  @Override
  public String getText(Range range) {
    return this.editor
        .getDocument()
        .getText(TextRange.create(offsetAt(range.start), offsetAt(range.end)));
  }

  @Override
  public Position positionAt(int offset) {
    int line = this.editor.getDocument().getLineNumber(offset);
    int lineStartOffset = offsetAt(new Position(line, 0));
    return new Position(line, offset - lineStartOffset);
  }

  @Override
  public AutocompleteDocumentContext getAutocompleteContext(int offset) {
    Document document = this.editor.getDocument();
    int line = document.getLineNumber(offset);
    int lineEndOffset = document.getLineEndOffset(line);
    String sameLineSuffix = document.getText(TextRange.create(offset, lineEndOffset));
    int lineStartOffset = document.getLineStartOffset(line);
    String sameLinePrefix = document.getText(TextRange.create(lineStartOffset, offset));
    return new AutocompleteDocumentContext(sameLinePrefix, sameLineSuffix);
  }

  @Override
  public @NotNull Optional<String> getLanguageId() {
    return Optional.ofNullable(this.language).map(Language::getID);
  }
}
