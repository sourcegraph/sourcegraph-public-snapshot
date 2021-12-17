import * as Sentry from '@sentry/browser'
import { trimStart } from 'lodash'
import { defer, of } from 'rxjs'
import { map } from 'rxjs/operators'
import { Omit } from 'utility-types'

import { NotificationType } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { AdjustmentDirection, PositionAdjuster } from '@sourcegraph/shared/src/codeintellify'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { observeSystemIsLightTheme } from '@sourcegraph/shared/src/theme'
import {
    FileSpec,
    RepoSpec,
    ResolvedRevisionSpec,
    RevisionSpec,
    toAbsoluteBlobURL,
} from '@sourcegraph/shared/src/util/url'

import LogoSVG from '../../../../assets/img/sourcegraph-mark.svg'
import { background } from '../../../browser-extension/web-extension-api/runtime'
import { fetchBlobContentLines } from '../../repo/backend'
import { querySelectorAllOrSelf, querySelectorOrSelf } from '../../util/dom'
import { CodeHost, MountGetter } from '../shared/codeHost'
import { CodeView, toCodeViewResolver } from '../shared/codeViews'
import { createNotificationClassNameGetter } from '../shared/getNotificationClassName'
import { NativeTooltip } from '../shared/nativeTooltips'
import { getSelectionsFromHash, observeSelectionsFromHash } from '../shared/util/selections'
import { ViewResolver } from '../shared/views'

import { markdownBodyViewResolver } from './contentViews'
import { diffDomFunctions, searchCodeSnippetDOMFunctions, singleFileDOMFunctions } from './domFunctions'
import { getCommandPaletteMount } from './extensions'
import { resolveDiffFileInfo, resolveFileInfo, resolveSnippetFileInfo } from './fileInfo'
import { setElementTooltip } from './tooltip'
import { getFileContainers, parseURL } from './util'

/**
 * Creates the mount element for the CodeViewToolbar on code views containing
 * a `.file-actions` element, for instance:
 * - A diff code view on a PR's files page, or a commit page
 * - An older GHE single file code view (newer GitHub.com code views use createFileLineContainerToolbarMount)
 */
export function createFileActionsToolbarMount(codeView: HTMLElement): HTMLElement {
    const className = 'github-file-actions-toolbar-mount'
    const existingMount = codeView.querySelector('.' + className) as HTMLElement
    if (existingMount) {
        return existingMount
    }

    const mountElement = document.createElement('div')
    mountElement.className = className

    const fileActions = codeView.querySelector('.file-actions')
    if (!fileActions) {
        throw new Error('Could not find GitHub file actions with selector .file-actions')
    }

    // Add a class to the .file-actions element, so that we can reliably match it in
    // stylesheets without bleeding CSS to other code hosts (GitLab also uses .file-actions elements).
    fileActions.classList.add('sg-github-file-actions')

    // Old GitHub Enterprise PR views have a "☑ show comments" text that we want to insert *after*
    const showCommentsElement = codeView.querySelector('.show-file-notes')
    if (showCommentsElement) {
        showCommentsElement.after(mountElement)
    } else {
        fileActions.prepend(mountElement)
    }

    return mountElement
}

const diffCodeView: Omit<CodeView, 'element'> = {
    dom: diffDomFunctions,
    getToolbarMount: createFileActionsToolbarMount,
    resolveFileInfo: resolveDiffFileInfo,
    getScrollBoundaries: codeView => {
        const fileHeader = codeView.querySelector<HTMLElement>('.file-header')
        if (!fileHeader) {
            throw new Error('Could not find .file-header element in GitHub PR code view')
        }
        return [fileHeader]
    },
}

const diffConversationCodeView: Omit<CodeView, 'element'> = {
    ...diffCodeView,
    getToolbarMount: undefined,
}

const singleFileCodeView: Omit<CodeView, 'element'> = {
    dom: singleFileDOMFunctions,
    getToolbarMount: createFileActionsToolbarMount,
    resolveFileInfo,
    getSelections: getSelectionsFromHash,
    observeSelections: observeSelectionsFromHash,
}

/**
 * Some code snippets get leading white space trimmed. This adjusts based on
 * this. See an example here https://github.com/sourcegraph/browser-extensions/issues/188.
 */
const getSnippetPositionAdjuster = (
    requestGraphQL: PlatformContext['requestGraphQL']
): PositionAdjuster<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec> => ({ direction, codeView, position }) =>
    fetchBlobContentLines({ ...position, requestGraphQL }).pipe(
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
    getPositionAdjuster: getSnippetPositionAdjuster,
    resolveFileInfo: resolveSnippetFileInfo,
})

const snippetCodeView: Omit<CodeView, 'element'> = {
    dom: singleFileDOMFunctions,
    resolveFileInfo: resolveSnippetFileInfo,
    getPositionAdjuster: getSnippetPositionAdjuster,
}

export const createFileLineContainerToolbarMount: NonNullable<CodeView['getToolbarMount']> = (
    codeViewElement: HTMLElement
): HTMLElement => {
    const className = 'sourcegraph-github-file-code-view-toolbar-mount'
    const existingMount = codeViewElement.querySelector(`.${className}`) as HTMLElement
    if (existingMount) {
        return existingMount
    }
    const mountElement = document.createElement('div')
    mountElement.style.display = 'inline-flex'
    mountElement.style.verticalAlign = 'middle'
    mountElement.style.alignItems = 'center'
    mountElement.className = className
    const rawURLLink = codeViewElement.querySelector('#raw-url')
    const buttonGroup = rawURLLink?.closest('.BtnGroup')
    if (!buttonGroup?.parentNode) {
        throw new Error('File actions not found')
    }
    buttonGroup.parentNode.insertBefore(mountElement, buttonGroup)
    return mountElement
}

/**
 * Matches the modern single-file code view, or snippets embedded in comments.
 *
 */
export const fileLineContainerResolver: ViewResolver<CodeView> = {
    selector: '.js-file-line-container',
    resolveView: (fileLineContainer: HTMLElement): CodeView | null => {
        const embeddedBlobWrapper = fileLineContainer.closest('.blob-wrapper-embedded')
        if (embeddedBlobWrapper) {
            // This is a snippet embedded in a comment.
            // Resolve to `.blob-wrapper-embedded`'s parent element,
            // the smallest element that contains both the code and
            // the HTML anchor allowing to resolve the file info.
            const element = embeddedBlobWrapper.parentElement!
            return {
                element,
                ...snippetCodeView,
            }
        }
        const { pageType } = parseURL()
        if (pageType !== 'blob') {
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
    selector: target => {
        const codeViews = new Set<HTMLElement>()

        // Logic to support large diffs that are loaded asynchronously:
        // https://github.com/sourcegraph/sourcegraph/issues/18337
        // - Don't return `.file` elements that have yet to be loaded (loading is triggered by user)
        // - When the user triggers diff loading, the mutation observer will tell us about
        // .js-blob-wrapper, since the actual '.file' has been in the DOM the whole time. Return
        // the closest ancestor '.file'

        for (const file of querySelectorAllOrSelf<HTMLElement>(target, '.file')) {
            if (file.querySelectorAll('.js-diff-load-container').length === 0) {
                codeViews.add(file)
            }
        }

        for (const blobWrapper of querySelectorAllOrSelf(target, '.js-blob-wrapper')) {
            const file = blobWrapper.closest('.file')
            if (file instanceof HTMLElement) {
                codeViews.add(file)
            }
        }

        return [...codeViews]
    },
    resolveView: (element: HTMLElement): CodeView | null => {
        if (element.querySelector('article.markdown-body')) {
            // This code view is rendered markdown, we shouldn't add code intelligence
            return null
        }

        // This is a suggested change on a GitHub PR
        if (element.closest('.js-suggested-changes-blob')) {
            return null
        }

        const { pageType } = parseURL()
        const isSingleCodeFile =
            pageType === 'blob' &&
            document.querySelectorAll('.file').length === 1 &&
            document.querySelectorAll('.diff-view').length === 0

        if (isSingleCodeFile) {
            return { element, ...singleFileCodeView }
        }

        if (element.closest('.discussion-item-body') || element.classList.contains('js-comment-container')) {
            // This code view is embedded on a PR conversation page.
            return { element, ...diffConversationCodeView }
        }

        return { element, ...diffCodeView }
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
export const checkIsGitHubDotCom = (url = window.location.href): boolean => /^https?:\/\/(www\.)?github\.com/.test(url)

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
    pageheadActions.prepend(mount)
    return mount
}

const nativeTooltipResolver: ViewResolver<NativeTooltip> = {
    selector: '.js-tagsearch-popover',
    resolveView: element => ({ element }),
}

const iconClassName = 'icon--github v-align-text-bottom'

const notificationClassNames = {
    [NotificationType.Log]: 'flash',
    [NotificationType.Success]: 'flash flash-success',
    [NotificationType.Info]: 'flash',
    [NotificationType.Warning]: 'flash flash-warn',
    [NotificationType.Error]: 'flash flash-error',
}

const searchEnhancement: CodeHost['searchEnhancement'] = {
    searchViewResolver: {
        selector: '.js-site-search-form input[type="text"][aria-controls="jump-to-results"]',
        resolveView: element => ({ element }),
    },
    resultViewResolver: {
        selector: '#jump-to-suggestion-search-global',
        resolveView: element => ({ element }),
    },
    onChange: ({ value, searchURL, resultElement: ghElement }) => {
        const SEARCH_IN_SOURCEGRAPH_SELECTOR = '#jump-to-sourcegraph-search-global'

        /** Create "Search in Sourcegraph" element based on GH element */
        const createElement = (): HTMLElement => {
            /** SG Base element on top of GH "All Github" element */
            const sgElement = ghElement.cloneNode(true) as HTMLElement
            sgElement.id = SEARCH_IN_SOURCEGRAPH_SELECTOR.replace('#', '')
            sgElement.classList.remove('navigation-focus')
            sgElement.setAttribute('aria-selected', 'false')

            /** Add sourcegraph logo */
            const logo = document.createElement('img')
            logo.src = LogoSVG
            logo.setAttribute('style', 'width: 16px; height: 20px; float: left; margin-right: 2px;')
            logo.setAttribute('alt', 'Sourcegraph Logo Image')

            /** Update badge text */
            const badge = sgElement.querySelector('.js-jump-to-badge-search-text-global') as HTMLElement
            badge.textContent = 'Sourcegraph'
            badge.parentNode?.insertBefore(logo, badge)

            /** Add sourcegraph item after GH item */
            ghElement.parentNode?.insertBefore(sgElement, ghElement.nextElementSibling)

            return sgElement
        }

        /** Update link and display value */
        const updateContent = (sgElement: HTMLElement): void => {
            const displayValue = sgElement.querySelector<HTMLElement>('.jump-to-suggestion-name') as HTMLElement
            displayValue.textContent = value
            displayValue.setAttribute('aria-label', value)

            const link = sgElement.querySelector<HTMLElement>('a') as HTMLLinkElement
            const url = new URL(searchURL)
            url.searchParams.append('q', value)
            link.setAttribute('href', url.href)
            link.setAttribute('target', '_blank')
            sgElement.setAttribute('style', `display: ${value ? 'initial' : 'none !important'}`)
        }

        updateContent(document.querySelector<HTMLElement>(SEARCH_IN_SOURCEGRAPH_SELECTOR) ?? createElement())
    },
}

/**
 * Checks whether repository is private or not using Github API + fallback to DOM element check
 *
 * @description See https://docs.github.com/en/rest/reference/repos#get-a-repository
 * @description see rate limit https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting
 */
export const isPrivateRepository = (
    repoName: string,
    fetchCache = background.fetchCache,
    fallbackSelector = '#repository-container-header h1 span.Label'
): Promise<boolean> => {
    if (window.location.hostname !== 'github.com') {
        return Promise.resolve(true)
    }
    return fetchCache<{ private?: boolean }>({
        url: `https://api.github.com/repos/${repoName}`,
        credentials: 'omit',
        cacheMaxAge: 60 * 60 * 1000, // 1 hour
    })
        .then(response => {
            const rateLimit = response.headers['x-ratelimit-remaining']
            if (Number(rateLimit) <= 0) {
                const rateLimitError = new Error('Github rate limit exceeded.')
                Sentry.captureException(rateLimitError)
                throw rateLimitError
            }
            return response
        })
        .then(({ data }) => typeof data.private !== 'boolean' || data.private)
        .catch(error => {
            // If network error or rate-limit exceeded fallback to DOM check
            console.warn('Failed to fetch if the repository is private.', error)
            return document.querySelector(fallbackSelector)?.textContent?.toLowerCase().trim() !== 'public'
        })
}

export const githubCodeHost: CodeHost = {
    type: 'github',
    name: checkIsGitHubEnterprise() ? 'GitHub Enterprise' : 'GitHub',
    searchEnhancement,
    codeViewResolvers: [genericCodeViewResolver, fileLineContainerResolver, searchResultCodeViewResolver],
    contentViewResolvers: [markdownBodyViewResolver],
    nativeTooltipResolvers: [nativeTooltipResolver],
    getContext: async () => {
        const { repoName, rawRepoName, pageType } = parseURL()

        return {
            rawRepoName,
            revision: pageType === 'blob' || pageType === 'tree' ? resolveFileInfo().blob.revision : undefined,
            privateRepository: await isPrivateRepository(repoName),
        }
    },
    isLightTheme: defer(() => {
        const mode = document.documentElement.dataset.colorMode as 'auto' | 'light' | 'dark' | undefined
        if (mode === 'auto') {
            return observeSystemIsLightTheme().observable
        }
        return of(mode !== 'dark')
    }),
    getViewContextOnSourcegraphMount: createOpenOnSourcegraphIfNotExists,
    viewOnSourcegraphButtonClassProps: {
        className: 'btn btn-sm tooltipped tooltipped-s',
        iconClassName,
    },
    check: checkIsGitHub,
    getCommandPaletteMount,
    notificationClassNames,
    commandPaletteClassProps: {
        buttonClassName: 'Header-link d-flex flex-items-baseline',
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
        iconClassName,
    },
    codeViewToolbarClassProps: {
        className: 'code-view-toolbar--github',
        listItemClass: 'code-view-toolbar__item--github BtnGroup',
        actionItemClass: 'btn btn-sm tooltipped tooltipped-s BtnGroup-item action-item--github',
        actionItemPressedClass: 'selected',
        actionItemIconClass: 'icon--github v-align-text-bottom',
    },
    hoverOverlayClassProps: {
        className: 'Box',
        actionItemClassName: 'btn btn-secondary',
        actionItemPressedClassName: 'active',
        badgeClassName: 'label hover-overlay__badge--github',
        getAlertClassName: createNotificationClassNameGetter(notificationClassNames, 'flash-full'),
        iconClassName,
    },
    setElementTooltip,
    linkPreviewContentClass: 'text-small text-gray p-1 mx-1 border rounded-1 bg-gray text-gray-dark',
    urlToFile: (sourcegraphURL, target, context) => {
        if (target.viewState) {
            // A view state means that a panel must be shown, and panels are currently only supported on
            // Sourcegraph (not code hosts).
            return toAbsoluteBlobURL(sourcegraphURL, target)
        }

        // Make sure the location is also on this github instance, return an absolute URL otherwise.
        const sameCodeHost = target.rawRepoName.startsWith(window.location.hostname)
        if (!sameCodeHost) {
            return toAbsoluteBlobURL(sourcegraphURL, target)
        }

        const revision = target.revision || 'HEAD'
        // If we're provided options, we can make the j2d URL more specific.
        const { rawRepoName } = parseURL()

        // Stay on same page in PR if possible.
        // TODO to be entirely correct, this would need to compare the revision of the code view with the target revision.
        const isSameRepo = rawRepoName === target.rawRepoName
        if (isSameRepo && context.part !== undefined) {
            const containers = getFileContainers()
            for (const container of containers) {
                const header = container.querySelector<HTMLElement & { dataset: { path: string; anchor: string } }>(
                    '.file-header[data-path][data-anchor]'
                )
                if (!header) {
                    // E.g. suggestion snippet
                    continue
                }
                const anchorPath = header.dataset.path
                if (anchorPath === target.filePath) {
                    const anchorUrl = header.dataset.anchor
                    const url = new URL(window.location.href)
                    url.hash = anchorUrl
                    if (target.position) {
                        // GitHub uses L for the left side, R for both right side and the unchanged/white parts
                        url.hash += `${context.part === 'base' ? 'L' : 'R'}${target.position.line}`
                    }
                    // Only use URL if it is visible
                    // TODO: Expand hidden lines to reveal
                    if (!document.querySelector(url.hash)) {
                        break
                    }
                    return url.href
                }
            }
        }

        // Go to blob URL
        const fragment = target.position
            ? `#L${target.position.line}${target.position.character ? `:${target.position.character}` : ''}`
            : ''
        return `https://${target.rawRepoName}/blob/${revision}/${target.filePath}${fragment}`
    },
    codeViewsRequireTokenization: true,
}
