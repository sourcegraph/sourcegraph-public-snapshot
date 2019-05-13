import { querySelectorAllOrSelf } from '../../shared/util/dom'
import { ViewResolver } from '../code_intelligence/views'

/**
 * Matches all GitHub Markdown body content, including comment bodies, issue/PR descriptions, review
 * comments, and rendered Markdown files.
 */
export const markdownBodyViewResolver: ViewResolver = container =>
    [...querySelectorAllOrSelf<HTMLElement>(container, '.markdown-body')].map(element => ({ element }))
