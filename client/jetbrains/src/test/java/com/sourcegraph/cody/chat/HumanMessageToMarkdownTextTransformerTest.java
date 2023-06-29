package com.sourcegraph.cody.chat;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.*;

import org.commonmark.node.HtmlInline;
import org.commonmark.node.Node;
import org.commonmark.node.Paragraph;
import org.commonmark.node.Text;
import org.commonmark.parser.Parser;
import org.junit.jupiter.api.Test;

public class HumanMessageToMarkdownTextTransformerTest {
  @Test
  public void shouldTransformNewLinesToMarkdownNewLines() {
    // given
    HumanMessageToMarkdownTextTransformer transformer =
        new HumanMessageToMarkdownTextTransformer("foo\nbar");

    // when
    String result = transformer.transform();

    // then
    Node rootMarkdownNode = readMarkdown(result);
    Paragraph firstParagraph = (Paragraph) rootMarkdownNode.getFirstChild();
    Text textInFirstParagraph = (Text) firstParagraph.getFirstChild();
    assertThat(textInFirstParagraph.getLiteral()).isEqualTo("foo");
    Paragraph secondParagraph = (Paragraph) rootMarkdownNode.getLastChild();
    Text textInSecondParagraph = (Text) secondParagraph.getLastChild();
    assertThat(textInSecondParagraph.getLiteral()).isEqualTo("bar");
  }

  @Test
  public void shouldNotSplitTextToNewLineInMarkdownWhenThereIsNoNewLine() {
    // given
    HumanMessageToMarkdownTextTransformer transformer =
        new HumanMessageToMarkdownTextTransformer("foo bar");

    // when
    String result = transformer.transform();

    // then
    Node rootMarkdownNode = readMarkdown(result);
    Paragraph firstParagraph = (Paragraph) rootMarkdownNode.getFirstChild();
    Text textInFirstParagraph = (Text) firstParagraph.getFirstChild();
    assertThat(textInFirstParagraph.getLiteral()).isEqualTo("foo bar");
    Paragraph secondParagraph = (Paragraph) rootMarkdownNode.getLastChild();
    // there are no new nodes, first and last child is the same node
    assertThat(firstParagraph == secondParagraph).isTrue();
  }

  @Test
  public void shouldAddDoubledMarkdownLines() {
    // given
    HumanMessageToMarkdownTextTransformer transformer =
        new HumanMessageToMarkdownTextTransformer("foo\n\nbar");

    // when
    String result = transformer.transform();

    // then
    Node rootMarkdownNode = readMarkdown(result);
    Paragraph firstParagraph = (Paragraph) rootMarkdownNode.getFirstChild();
    Text textInFirstParagraph = (Text) firstParagraph.getFirstChild();
    assertThat(textInFirstParagraph.getLiteral()).isEqualTo("foo");
    HtmlInline firstNewLine = (HtmlInline) textInFirstParagraph.getNext();
    assertThat(firstNewLine.getLiteral()).isEqualTo("<br />");
    HtmlInline secondNewLine = (HtmlInline) firstNewLine.getNext();
    assertThat(secondNewLine.getLiteral()).isEqualTo("<br />");
    Text textInLastParagraph = (Text) secondNewLine.getNext();
    assertThat(textInLastParagraph.getLiteral()).isEqualTo("bar");
  }

  private Node readMarkdown(String result) {
    return Parser.builder().build().parse(result);
  }
}
