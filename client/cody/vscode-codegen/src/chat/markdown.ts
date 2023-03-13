import { marked } from 'marked'
import sanitize from 'sanitize-html'

import { highlightCode, registerHighlightContributions } from './highlight'

const sanitizeOptions = {
	...sanitize.defaults,
	allowedAttributes: {
		...sanitize.defaults.allowedAttributes,
		a: [
			...sanitize.defaults.allowedAttributes.a,
			'title',
			'class',
			{ name: 'rel', values: ['noopener', 'noreferrer'] },
		],
		// Allow highlight.js styles, e.g.
		// <span class="hljs-keyword">
		// <code class="language-javascript">
		span: ['class'],
		code: ['class'],
	},
}

export function renderMarkdown(text: string): string {
	registerHighlightContributions()
	return sanitize(marked.parse(text, { gfm: true, highlight: highlightCode, breaks: true }), sanitizeOptions)
}
