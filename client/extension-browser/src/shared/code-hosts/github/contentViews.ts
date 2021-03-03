import { ContentView } from '../ui-kit-legacy-shared/contentViews'
import { ViewResolver } from '../ui-kit-legacy-shared/views'

/**
 * Matches all GitHub Markdown body content, including comment bodies, issue/PR descriptions, review
 * comments, and rendered Markdown files.
 */
export const markdownBodyViewResolver: ViewResolver<ContentView> = {
    selector: '.markdown-body',
    resolveView: element => ({ element }),
}
