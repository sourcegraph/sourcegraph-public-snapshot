package com.sourcegraph.cody.chat

import org.commonmark.node.Document
import org.commonmark.node.FencedCodeBlock
import org.commonmark.node.IndentedCodeBlock
import org.commonmark.node.Node

fun Node.isCodeBlock(): Boolean {
  return this is FencedCodeBlock || this is IndentedCodeBlock
}

fun Node.findNodeAfterLastCodeBlock(): Node {
  val lastNodeAfterCode =
      generateSequence(lastChild) { it.previous }.takeWhile { !it.isCodeBlock() }.lastOrNull()
  return lastNodeAfterCode.buildNewDocumentFrom()
}

fun Node?.buildNewDocumentFrom(): Document {
  var nodeAfterCode = this
  val document = Document()
  while (nodeAfterCode != null) {
    val nextNode = nodeAfterCode.next
    document.appendChild(nodeAfterCode)
    nodeAfterCode = nextNode
  }
  return document
}

fun Node.extractCodeAndLanguage() =
    when (this) {
      is FencedCodeBlock -> Pair(literal, info)
      is IndentedCodeBlock -> Pair(literal, "")
      else -> Pair("", "")
    }
