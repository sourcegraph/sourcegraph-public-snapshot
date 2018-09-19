import { AdjustmentDirection, PositionAdjuster } from '@sourcegraph/codeintellify'
import { trimStart } from 'lodash'
import { map } from 'rxjs/operators'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { CodeHost, CodeView, CodeViewResolver, CodeViewWithOutSelector } from '../code_intelligence'
import { diffDomFunctions, searchCodeSnippetDOMFunctions, singleFileDOMFunctions } from './dom_functions'
import { resolveDiffFileInfo, resolveFileInfo, resolveSnippetFileInfo } from './file_info'
import { createCodeViewToolbarMount, parseURL } from './util'

const toolbarButtonProps = {
    className: 'btn btn-sm tooltipped tooltipped-n',
    style: { marginRight: '5px', textDecoration: 'none', color: 'inherit' },
}

const diffCodeView: CodeViewWithOutSelector = {
    dom: diffDomFunctions,
    getToolbarMount: createCodeViewToolbarMount,
    resolveFileInfo: resolveDiffFileInfo,
    toolbarButtonProps,
}

const singleFileCodeView: CodeViewWithOutSelector = {
    dom: singleFileDOMFunctions,
    getToolbarMount: createCodeViewToolbarMount,
    resolveFileInfo,
    toolbarButtonProps,
}

/**
 * Some code snippets get leading white space trimmed. This adjusts based on
 * this. See an example here https://github.com/sourcegraph/browser-extensions/issues/188.
 */
const adjustPositionForSnippet: PositionAdjuster = ({ direction, codeView, position }) =>
    fetchBlobContentLines(position).pipe(
        map(lines => {
            const codeElement = singleFileDOMFunctions.getCodeElementFromLineNumber(
                codeView,
                position.line,
                position.part
            )
            if (!codeElement) {
                throw new Error('(adjustPosition) could not find code element for line provided')
            }

            const actualLine = lines[position.line - 1]
            const documentLine = codeElement.textContent || ''

            const actualLeadingWhiteSpace = actualLine.length - trimStart(actualLine).length
            const documentLeadingWhiteSpace = documentLine.length - trimStart(documentLine).length

            const modifier = direction === AdjustmentDirection.ActualToCodeView ? -1 : 1
            const delta = Math.abs(actualLeadingWhiteSpace - documentLeadingWhiteSpace) * modifier

            return {
                line: position.line,
                character: position.character + delta,
            }
        })
    )

const searchResultCodeView: CodeView = {
    selector: '.code-list-item',
    dom: searchCodeSnippetDOMFunctions,
    adjustPosition: adjustPositionForSnippet,
    resolveFileInfo: resolveSnippetFileInfo,
    toolbarButtonProps,
}

const commentSnippetCodeView: CodeView = {
    selector: '.js-comment-body',
    dom: singleFileDOMFunctions,
    resolveFileInfo: resolveSnippetFileInfo,
    adjustPosition: adjustPositionForSnippet,
    toolbarButtonProps,
}

const resolveCodeView = (elem: HTMLElement): CodeViewWithOutSelector => {
    const files = document.getElementsByClassName('file')
    const { filePath } = parseURL()
    const isSingleCodeFile = files.length === 1 && filePath && document.getElementsByClassName('diff-view').length === 0

    return isSingleCodeFile ? singleFileCodeView : diffCodeView
}

const codeViewResolver: CodeViewResolver = {
    selector: '.file',
    resolveCodeView,
}

function checkIsGithub(): boolean {
    const href = window.location.href

    const isGithub = /^https?:\/\/(www.)?github.com/.test(href)
    const ogSiteName = document.head.querySelector(`meta[property='og:site_name']`) as HTMLMetaElement
    const isGitHubEnterprise = ogSiteName ? ogSiteName.content === 'GitHub Enterprise' : false

    return isGithub || isGitHubEnterprise
}

export const githubCodeHost: CodeHost = {
    name: 'github',
    codeViews: [searchResultCodeView, commentSnippetCodeView],
    codeViewResolver,
    check: checkIsGithub,
}
