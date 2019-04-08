import { AdjustmentDirection, DiffPart, PositionAdjuster } from '@sourcegraph/codeintellify'
import { trimStart } from 'lodash'
import { NEVER } from 'rxjs'
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
import { CodeViewSpec, DiffViewSpec, DiffViewSpecResolver, MountGetter } from '../code_intelligence'
import { observeDiffViewVisibleRanges as observeDiffViewCollapsed, setDiffViewCollapsed } from './diff_views'
import { diffDomFunctions, searchCodeSnippetDOMFunctions, singleFileDOMFunctions } from './dom_functions'
import { getCommandPaletteMount, getGlobalDebugMount } from './extensions'
import { resolveDiffFileInfo, resolveFileInfo, resolveSnippetFileInfo } from './file_info'
import { getFileContainers, parseURL } from './util'

/**
 * Creates the mount element for the CodeViewToolbar.
 */
export function createCodeViewToolbarMount(codeView: HTMLElement): HTMLElement {
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

const searchResultCodeView: CodeViewSpec = {
    selector: '.code-list-item',
    dom: searchCodeSnippetDOMFunctions,
    adjustPosition: adjustPositionForSnippet,
    resolveFileInfo: resolveSnippetFileInfo,
    toolbarButtonProps,
    isDiff: false,
}

const commentSnippetCodeView: CodeViewSpec = {
    selector: '.js-comment-body',
    dom: singleFileDOMFunctions,
    resolveFileInfo: resolveSnippetFileInfo,
    adjustPosition: adjustPositionForSnippet,
    toolbarButtonProps,
    isDiff: false,
}

/**
 * The modern single file blob view.
 *
 * @todo This code view does not follow the code view contract because
 * the selector returns just the code table, not including the toolbar.
 * This requires `getToolbarMount()` to look at the parent elements, which makes it not possible
 * unit test like other toolbar mount getters.
 * Change this after https://github.com/sourcegraph/sourcegraph/issues/3271 is fixed.
 */
export const fileLineContainerCodeView = {
    selector: '.js-file-line-container',
    dom: singleFileDOMFunctions,
    getToolbarMount: createCodeViewToolbarMount,
    resolveFileInfo,
    toolbarButtonProps,
    isDiff: false,
}

const diffViewSpecResolver: DiffViewSpecResolver = {
    // TODO!(sqs): ensure this doesnt match issues with snippets
    selector: '.file.has-inline-notes, .file[data-file-deleted]',
    resolveDiffViewSpec: (elem: HTMLElement): DiffViewSpec | null => {
        const isPRTimelineComment = !!elem.closest('.discussion-item-body')
        const hasDiffHeader = !isPRTimelineComment
        return {
            dom: diffDomFunctions,
            getToolbarMount: hasDiffHeader ? createCodeViewToolbarMount : undefined,
            resolveDiffInfo: resolveDiffFileInfo,
            toolbarButtonProps,
            collapsedChanges: hasDiffHeader ? observeDiffViewCollapsed(elem) : NEVER,
            setCollapsed: ranges => setDiffViewCollapsed(elem, ranges),
        }
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

const getOverlayMount: MountGetter = (container: HTMLElement): HTMLElement | null => {
    const jsRepoPjaxContainer = querySelectorOrSelf(container, '#js-repo-pjax-container')
    if (!jsRepoPjaxContainer) {
        return null
    }
    let mount = jsRepoPjaxContainer.querySelector<HTMLElement>('.hover-overlay-mount')
    if (mount) {
        return mount
    }
    mount = document.createElement('div')
    mount.className = 'hover-overlay-mount'
    jsRepoPjaxContainer.appendChild(mount)
    return mount
}

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

export const githubCodeHost = {
    name: 'github',
    codeViewSpecs: [searchResultCodeView, commentSnippetCodeView, fileLineContainerCodeView],
    diffViewSpecResolver: [diffViewSpecResolver],
    getContext: parseURL,
    getViewContextOnSourcegraphMount: createOpenOnSourcegraphIfNotExists,
    viewOnSourcegraphButtonClassProps: {
        className: 'btn btn-sm tooltipped tooltipped-s',
        iconClassName: 'action-item__icon--github v-align-text-bottom',
    },
    check: checkIsGitHub,
    getOverlayMount,
    getCommandPaletteMount,
    getGlobalDebugMount,
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
    hoverOverlayClassProps: {
        actionItemClassName: 'btn btn-secondary',
        actionItemPressedClassName: 'active',
        closeButtonClassName: 'btn',
    },
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
