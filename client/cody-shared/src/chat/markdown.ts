import { registerHighlightContributions, renderMarkdown as renderMarkdownCommon } from '@sourcegraph/common'

export function renderMarkdown(markdown: string): string {
    registerHighlightContributions()

    // TODO(sqs): add Cody-specific Markdown rendering if needed
    return renderMarkdownCommon(markdown)
}
