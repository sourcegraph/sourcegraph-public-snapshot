// This module takes the place of vs/editor/editor.main and vs/editor/browser/editor.all.
// We import a subset of the modules those files import, to reduce the bundle size in
// the browser (and because we don't need all of the functionality included by default).
import "vs/editor/browser/widget/codeEditorWidget";
import "vs/editor/contrib/clipboard/browser/clipboard";
import "vs/editor/contrib/contextmenu/browser/contextmenu";
import "vs/editor/contrib/find/browser/find";
import "vs/editor/contrib/goToDeclaration/browser/goToDeclaration"; // TODO(sqs!vscode): disable ctrl-hover
import "vs/editor/contrib/hover/browser/hover";
import "vs/editor/contrib/links/browser/links";
import "vs/editor/contrib/referenceSearch/browser/referenceSearch";
import "vs/editor/contrib/toggleWordWrap/common/toggleWordWrap";
import "vs/editor/contrib/wordHighlighter/browser/wordHighlighter.css";
import "vs/editor/contrib/wordHighlighter/common/wordHighlighter";

import "vs/editor/common/languages.common";

import "sourcegraph/editor/modes";

// HACK: vscode's markdown parser exports itself in a nonstandard
// way. This workaround avoids errors like "Uncaught TypeError:
// marked_1.marked.Renderer is not a constructor".
import * as marked from "vs/base/common/marked/marked";
marked.marked.Renderer = (marked as any).marked.marked.Renderer;
Object.assign(marked, marked.marked); // make it callable
// END HACK

import {createMonacoEditorAPI} from "vs/editor/browser/standalone/standaloneEditor";
import {createMonacoLanguagesAPI} from "vs/editor/browser/standalone/standaloneLanguages";
import {DefaultConfig} from "vs/editor/common/config/defaultConfig";
import {createMonacoBaseAPI} from "vs/editor/common/standalone/standaloneBase";

// Set defaults for standalone editor
DefaultConfig.editor.wrappingIndent = "none";
DefaultConfig.editor.folding = false;

export const monaco = Object.assign({}, createMonacoBaseAPI());
monaco.editor = createMonacoEditorAPI();
monaco.languages = createMonacoLanguagesAPI();
