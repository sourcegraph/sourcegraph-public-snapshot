// This module takes the place of vs/editor/editor.main and vs/editor/browser/editor.all.
// We import a subset of the modules those files import, to reduce the bundle size in
// the browser (and because we don't need all of the functionality included by default).
import "vs/editor/browser/widget/codeEditorWidget";
import "vs/editor/browser/widget/diffEditorWidget";
import "vs/editor/contrib/clipboard/browser/clipboard";
import "vs/editor/contrib/contextmenu/browser/contextmenu";
import "vs/editor/contrib/find/browser/find";
import "vs/editor/contrib/goToDeclaration/browser/goToDeclaration";
import "vs/editor/contrib/hover/browser/hover";
import "vs/editor/contrib/links/browser/links";
import "vs/editor/contrib/referenceSearch/browser/referenceSearch";
import "vs/editor/contrib/wordHighlighter/browser/wordHighlighter.css";
import "vs/editor/contrib/wordHighlighter/common/wordHighlighter";

import "monaco-languages/out/monaco.contribution";
import "monaco-typescript/out/monaco.contribution";

import { DefaultConfig } from "vs/editor/common/config/defaultConfig";

// Set defaults for standalone editor
DefaultConfig.editor.wrappingIndent = "none";
DefaultConfig.editor.folding = false;
