import { AdjustmentDirection, DiffPart, PositionAdjuster } from '@sourcegraph/codeintellify'
import { trimStart } from 'lodash'
import { map } from 'rxjs/operators'
import {
    FileSpec,
    PositionSpec,
    RepoSpec,
    ResolvedRevSpec,
    RevSpec,
    ViewStateSpec,
} from '../../../../../shared/src/util/url'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { querySelectorOrSelf } from '../../shared/util/dom'
import { toAbsoluteBlobURL } from '../../shared/util/url'
import { CodeHost, MountGetter } from '../code_intelligence'
import { CodeView, CodeViewSpec, toCodeViewResolver } from '../code_intelligence/code_views'
import { getSelectionsFromHash, observeSelectionsFromHash } from '../code_intelligence/util/selections'
import { ViewResolver } from '../code_intelligence/views'
import { markdownBodyViewResolver } from './content_views'
import { diffDomFunctions, searchCodeSnippetDOMFunctions, singleFileDOMFunctions } from './dom_functions'
import { getCommandPaletteMount } from './extensions'
import { resolveDiffFileInfo, resolveFileInfo, resolveSnippetFileInfo } from './file_info'
import { commentTextFieldResolver } from './text_fields'
import { setElementTooltip } from './tooltip'
import { getFileContainers, parseURL } from './util'

/**
 * Creates the mount element for the CodeViewToolbar on code views containing
 * a `.file-actions` element.
 */
export function createFileActionsToolbarMount(codeView: HTMLElement): HTMLElement {
    const className = 'sourcegraph-app-annotator'
    const existingMount = codeView.querySelector('.' + className) as HTMLElement
    if (existingMount) {
        return existingMount
    }

    const mountEl = document.createElement('div')
    mountEl.style.display = 'inline-flex'
    mountEl.style.verticalAlign = 'middle'
    mountEl.style.alignItems = 'center'
    mountEl.className = className

    const fileActions = codeView.querySelector('.file-actions')
    if (!fileActions) {
        throw new Error(
            "File actions not found. Make sure you aren't trying to create " +
                "a toolbar mount for a code snippet that shouldn't have one"
        )
    }

    const buttonGroup = fileActions.querySelector('.BtnGroup')
    if (buttonGroup && buttonGroup.parentNode && !codeView.querySelector('.show-file-notes')) {
        // blob view
        buttonGroup.parentNode.insertBefore(mountEl, buttonGroup)
    } else {
        // commit & pull request view
        const note = codeView.querySelector('.show-file-notes')
        if (!note || !note.parentNode) {
            throw new Error('cannot find toolbar mount location')
        }
        note.parentNode.insertBefore(mountEl, note.nextSibling)
    }

    return mountEl
}

const toolbarButtonProps = {
    className: 'btn btn-sm tooltipped tooltipped-s',
}

const diffCodeView: CodeViewSpec = {
    dom: diffDomFunctions,
    getToolbarMount: createFileActionsToolbarMount,
    resolveFileInfo: resolveDiffFileInfo,
    toolbarButtonProps,
}

const diffConversationCodeView: CodeViewSpec = {
    ...diffCodeView,
    getToolbarMount: undefined,
}

const singleFileCodeView: CodeViewSpec = {
    dom: singleFileDOMFunctions,
    getToolbarMount: createFileActionsToolbarMount,
    resolveFileInfo,
    toolbarButtonProps,
    getSelections: getSelectionsFromHash,
    observeSelections: observeSelectionsFromHash,
}

/**
 * Some code snippets get leading white space trimmed. This adjusts based on
 * this. See an example here https://github.com/sourcegraph/browser-extensions/issues/188.
 */
const adjustPositionForSnippet: PositionAdjuster<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec> = ({
    direction,
    codeView,
    position,
}) =>
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

const searchResultCodeViewResolver = toCodeViewResolver('.code-list-item', {
    dom: searchCodeSnippetDOMFunctions,
    adjustPosition: adjustPositionForSnippet,
    resolveFileInfo: resolveSnippetFileInfo,
    toolbarButtonProps,
})

const commentSnippetCodeViewResolver = toCodeViewResolver('.js-comment-body', {
    dom: singleFileDOMFunctions,
    resolveFileInfo: resolveSnippetFileInfo,
    adjustPosition: adjustPositionForSnippet,
    toolbarButtonProps,
})

export const createFileLineContainerToolbarMount: NonNullable<CodeView['getToolbarMount']> = (
    repositoryContent: HTMLElement
): HTMLElement => {
    const className = 'sourcegraph-app-annotator'
    const existingMount = repositoryContent.querySelector(`.${className}`) as HTMLElement
    if (existingMount) {
        return existingMount
    }
    const mountEl = document.createElement('div')
    mountEl.style.display = 'inline-flex'
    mountEl.style.verticalAlign = 'middle'
    mountEl.style.alignItems = 'center'
    mountEl.className = className
    const rawURLLink = repositoryContent.querySelector('#raw-url')
    const buttonGroup = rawURLLink && rawURLLink.closest('.BtnGroup')
    if (!buttonGroup || !buttonGroup.parentNode) {
        throw new Error('File actions not found')
    }
    buttonGroup.parentNode.insertBefore(mountEl, buttonGroup)
    return mountEl
}

/**
 * The modern single file blob view.
 *
 */
export const fileLineContainerResolver: ViewResolver<CodeView> = {
    selector: '.js-file-line-container',
    resolveView: (fileLineContainer: HTMLElement): CodeView | null => {
        const { filePath } = parseURL()
        if (!filePath) {
            // this is not a single-file code view
            return null
        }
        const repositoryContent = fileLineContainer.closest('.repository-content')
        if (!repositoryContent) {
            throw new Error('Could not find repository content element')
        }
        return {
            element: repositoryContent as HTMLElement,
            ...singleFileCodeView,
            getToolbarMount: createFileLineContainerToolbarMount,
        }
    },
}

const genericCodeViewResolver: ViewResolver<CodeView> = {
    selector: '.file',
    resolveView: (elem: HTMLElement): CodeView | null => {
        if (elem.querySelector('article.markdown-body')) {
            // This code view is rendered markdown, we shouldn't add code intelligence
            return null
        }

        // This is a suggested change on a GitHub PR
        if (elem.closest('.js-suggested-changes-blob')) {
            return null
        }

        const files = document.getElementsByClassName('file')
        const { filePath } = parseURL()
        const isSingleCodeFile =
            files.length === 1 && filePath && document.getElementsByClassName('diff-view').length === 0

        if (isSingleCodeFile) {
            return { element: elem, ...singleFileCodeView }
        }

        if (elem.closest('.discussion-item-body')) {
            return { element: elem, ...diffConversationCodeView }
        }

        return { element: elem, ...diffCodeView }
    },
}

/**
 * Returns true if the current page is GitHub Enterprise.
 */
export function checkIsGitHubEnterprise(): boolean {
    const ogSiteName = document.head.querySelector<HTMLMetaElement>('meta[property="og:site_name"]')
    return (
        !!ogSiteName &&
        // GitHub Enterprise v2.14.11 has "GitHub" as og:site_name
        (ogSiteName.content === 'GitHub Enterprise' || ogSiteName.content === 'GitHub') &&
        document.body.classList.contains('enterprise')
    )
}

/**
 * Returns true if the current page is github.com.
 */
export const checkIsGitHubDotCom = (): boolean => /^https?:\/\/(www.)?github.com/.test(window.location.href)

/**
 * Returns true if the current page is either github.com or GitHub Enterprise.
 */
export const checkIsGitHub = (): boolean => checkIsGitHubDotCom() || checkIsGitHubEnterprise()

const OPEN_ON_SOURCEGRAPH_ID = 'open-on-sourcegraph'

export const createOpenOnSourcegraphIfNotExists: MountGetter = (container: HTMLElement): HTMLElement | null => {
    const pageheadActions = querySelectorOrSelf(container, '.pagehead-actions')
    // If ran on page that isn't under a repository namespace.
    if (!pageheadActions || pageheadActions.children.length === 0) {
        return null
    }
    // Check for existing
    let mount = pageheadActions.querySelector<HTMLElement>('#' + OPEN_ON_SOURCEGRAPH_ID)
    if (mount) {
        return mount
    }
    // Create new
    mount = document.createElement('li')
    mount.id = OPEN_ON_SOURCEGRAPH_ID
    pageheadActions.insertAdjacentElement('afterbegin', mount)
    return mount
}

export const githubCodeHost: CodeHost = {
    name: 'github',
    codeViewResolvers: [
        genericCodeViewResolver,
        fileLineContainerResolver,
        searchResultCodeViewResolver,
        commentSnippetCodeViewResolver,
    ],
    contentViewResolvers: [markdownBodyViewResolver],
    textFieldResolvers: [commentTextFieldResolver],
    getContext: parseURL,
    getViewContextOnSourcegraphMount: createOpenOnSourcegraphIfNotExists,
    viewOnSourcegraphButtonClassProps: {
        className: 'btn btn-sm tooltipped tooltipped-s',
        iconClassName: 'action-item__icon--github v-align-text-bottom',
    },
    check: checkIsGitHub,
    getCommandPaletteMount,
    commandPaletteClassProps: {
        popoverClassName: 'Box',
        formClassName: 'p-1',
        inputClassName: 'form-control input-sm header-search-input jump-to-field',
        listClassName: 'p-0 m-0 js-navigation-container jump-to-suggestions-results-container',
        selectedListItemClassName: 'navigation-focus',
        listItemClassName:
            'd-flex flex-justify-start flex-items-center p-0 f5 navigation-item js-navigation-item js-jump-to-scoped-search',
        actionItemClassName:
            'command-palette-action-item--github no-underline d-flex flex-auto flex-items-center jump-to-suggestions-path p-2',
        noResultsClassName: 'd-flex flex-auto flex-items-center jump-to-suggestions-path p-2',
    },
    codeViewToolbarClassProps: {
        className: 'code-view-toolbar--github',
        listItemClass: 'code-view-toolbar__item--github BtnGroup',
        actionItemClass: 'btn btn-sm tooltipped tooltipped-s BtnGroup-item action-item--github',
        actionItemPressedClass: 'selected',
        actionItemIconClass: 'action-item__icon--github v-align-text-bottom',
    },
    completionWidgetClassProps: {
        widgetContainerClassName: 'suggester-container',
        widgetClassName: 'suggester',
        listClassName: 'suggestions',
        selectedListItemClassName: 'navigation-focus',
        listItemClassName: 'text-normal',
    },
    hoverOverlayClassProps: {
        actionItemClassName: 'btn btn-secondary',
        actionItemPressedClassName: 'active',
        closeButtonClassName: 'btn',
    },
    setElementTooltip,
    linkPreviewContentClass: 'text-small text-gray p-1 mx-1 border rounded-1 bg-gray text-gray-dark',
    urlToFile: (
        location: RepoSpec & RevSpec & FileSpec & Partial<PositionSpec> & Partial<ViewStateSpec> & { part?: DiffPart }
    ) => {
        if (location.viewState) {
            // A view state means that a panel must be shown, and panels are currently only supported on
            // Sourcegraph (not code hosts).
            return toAbsoluteBlobURL(location)
        }

        const rev = location.rev || 'HEAD'
        // If we're provided options, we can make the j2d URL more specific.
        const { repoName } = parseURL()

        const sameRepo = repoName === location.repoName
        // Stay on same page in PR if possible.
        if (sameRepo && location.part) {
            const containers = getFileContainers()
            for (const container of containers) {
                const header = container.querySelector('.file-header') as HTMLElement
                const anchorPath = header.dataset.path
                if (anchorPath === location.filePath) {
                    const anchorUrl = header.dataset.anchor
                    const url = `${window.location.origin}${window.location.pathname}#${anchorUrl}${
                        location.part === 'base' ? 'L' : 'R'
                    }${location.position ? location.position.line : ''}`

                    return url
                }
            }
        }

        const fragment = location.position
            ? `#L${location.position.line}${location.position.character ? ':' + location.position.character : ''}`
            : ''
        return `https://${location.repoName}/blob/${rev}/${location.filePath}${fragment}`
    },
}
