package com.sourcegraph.cody.vscode;

import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.util.TextRange;
import com.intellij.openapi.vfs.VirtualFile;
import java.net.URI;
import java.nio.file.Paths;

/** Implementation of vscode.TextDocument backed by IntelliJ's Editor. */
public class IntelliJTextDocument implements TextDocument {
  public final Editor editor;
  public VirtualFile file;

  public IntelliJTextDocument(Editor editor) {
    this.editor = editor;
    this.file = FileDocumentManager.getInstance().getFile(editor.getDocument());
  }

  @Override
  public URI uri() {
    return Paths.get(file.getPath()).toUri();
  }

  @Override
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
}
