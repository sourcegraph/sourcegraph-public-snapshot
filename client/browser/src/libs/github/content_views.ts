import { ContentView } from '../code_intelligence/content_views'
import { ViewResolver } from '../code_intelligence/views'

/**
 * Matches all GitHub Markdown body content, including comment bodies, issue/PR descriptions, review
 * comments, and rendered Markdown files.
 */
export const markdownBodyViewResolver: ViewResolver<ContentView> = {
    selector: '.markdown-body',
    resolveView: element => ({ element }),
}
