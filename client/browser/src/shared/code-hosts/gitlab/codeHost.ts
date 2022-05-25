import * as Sentry from '@sentry/browser'
import classNames from 'classnames'
import { fromEvent } from 'rxjs'
import { filter, map } from 'rxjs/operators'
import { Omit } from 'utility-types'

import { LineOrPositionOrRange, subtypeOf } from '@sourcegraph/common'
import { NotificationType } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { toAbsoluteBlobURL } from '@sourcegraph/shared/src/util/url'

import { background } from '../../../browser-extension/web-extension-api/runtime'
import { CodeHost, OverlayPosition } from '../shared/codeHost'
import { CodeView } from '../shared/codeViews'
import { createNotificationClassNameGetter } from '../shared/getNotificationClassName'
import { getSelectionsFromHash, observeSelectionsFromHash } from '../shared/util/selections'
import { queryWithSelector, ViewResolver } from '../shared/views'

import { diffDOMFunctions, singleFileDOMFunctions } from './domFunctions'
import { getCommandPaletteMount } from './extensions'
import { resolveCommitFileInfo, resolveDiffFileInfo, resolveFileInfo } from './fileInfo'
import { getPageInfo, GitLabPageKind, getFilePathsFromCodeView } from './scrape'

import styles from './codeHost.module.scss'

export function checkIsGitlab(): boolean {
    return !!document.head.querySelector('meta[content="GitLab"]')
}

const adjustOverlayPosition: CodeHost['adjustOverlayPosition'] = args => {
    const topOrBottom = 'top' in args ? 'top' : 'bottom'
    let topOrBottomValue = 'top' in args ? args.top : args.bottom

    const header = document.querySelector('header')
    if (header) {
        topOrBottomValue += header.getBoundingClientRect().height
    }
    // When running GitLab from source, we also need to take into account
    // the debug header shown at the top of the page.
    const debugHeader = document.querySelector('#js-peek.development')
    if (debugHeader) {
        topOrBottomValue += debugHeader.getBoundingClientRect().height
    }

    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    return {
        [topOrBottom]: topOrBottomValue,
        left: args.left,
    } as OverlayPosition
}

export const getToolbarMount = (codeView: HTMLElement, pageKind?: GitLabPageKind): HTMLElement => {
    const existingMount: HTMLElement | null = codeView.querySelector('.sg-toolbar-mount-gitlab')
    if (existingMount) {
        return existingMount
    }

    const fileActions = codeView.querySelector('.file-actions')
    if (!fileActions) {
        throw new Error('Unable to find mount location')
    }

    const mount = document.createElement('div')
    mount.classList.add('sg-toolbar-mount', 'sg-toolbar-mount-gitlab')
    if (pageKind === GitLabPageKind.Commit) {
        mount.classList.add('gl-mr-3')
    }

    fileActions.prepend(mount)

    return mount
}

const singleFileCodeView: Omit<CodeView, 'element'> = {
    dom: singleFileDOMFunctions,
    getToolbarMount: (codeView: HTMLElement) => getToolbarMount(codeView, GitLabPageKind.File),
    resolveFileInfo,
    getSelections: getSelectionsFromHash,
    observeSelections: observeSelectionsFromHash,
}

const getFileTitle = (codeView: HTMLElement): HTMLElement[] => {
    const fileTitle = codeView.querySelector<HTMLElement>('.js-file-title')
    if (!fileTitle) {
        throw new Error('Could not find .file-title element')
    }
    return [fileTitle]
}

const mergeRequestCodeView: Omit<CodeView, 'element'> = {
    dom: diffDOMFunctions,
    getToolbarMount: (codeView: HTMLElement) => getToolbarMount(codeView, GitLabPageKind.MergeRequest),
    resolveFileInfo: resolveDiffFileInfo,
    getScrollBoundaries: getFileTitle,
}

const commitCodeView: Omit<CodeView, 'element'> = {
    dom: diffDOMFunctions,
    getToolbarMount: (codeView: HTMLElement) => getToolbarMount(codeView, GitLabPageKind.Commit),
    resolveFileInfo: resolveCommitFileInfo,
    getScrollBoundaries: getFileTitle,
}

const resolveView: ViewResolver<CodeView>['resolveView'] = (element: HTMLElement): CodeView | null => {
    if (element.classList.contains('discussion-wrapper')) {
        // This is a commented snippet in a merge request discussion timeline
        // (a snippet where somebody added a review comment on a piece of code in the MR),
        // we don't support adding code intelligence on those.
        return null
    }
    const { pageKind } = getPageInfo()

    if (pageKind === GitLabPageKind.Other) {
        return null
    }

    if (pageKind === GitLabPageKind.File) {
        return { element, ...singleFileCodeView }
    }

    if (pageKind === GitLabPageKind.MergeRequest) {
        if (!element.querySelector('.file-actions')) {
            // If the code view has no file actions, we cannot resolve its head commit ID.
            // This can be the case for code views representing added git submodules.
            return null
        }
        return { element, ...mergeRequestCodeView }
    }

    return { element, ...commitCodeView }
}

const codeViewResolver: ViewResolver<CodeView> = {
    selector: '.file-holder',
    resolveView,
}

const notificationClassNames = {
    [NotificationType.Log]: 'alert alert-secondary',
    [NotificationType.Success]: 'alert alert-success',
    [NotificationType.Info]: 'alert alert-info',
    [NotificationType.Warning]: 'alert alert-warning',
    [NotificationType.Error]: 'alert alert-danger',
}

/**
 * Checks whether repository is private or not using Gitlab API
 *
 * @description see https://docs.gitlab.com/ee/api/projects.html#get-single-project
 * @description see rate limit https://docs.gitlab.com/ee/user/admin_area/settings/user_and_ip_rate_limits.html#response-headers
 */
export const isPrivateRepository = (repoName: string, fetchCache = background.fetchCache): Promise<boolean> => {
    if (window.location.hostname !== 'gitlab.com' || !repoName) {
        return Promise.resolve(true)
    }
    return fetchCache({
        url: `https://gitlab.com/api/v4/projects/${encodeURIComponent(repoName)}`,
        credentials: 'omit', // it returns different response based on auth state
        cacheMaxAge: 60 * 60 * 1000, // 1 hour
    })
        .then(response => {
            const rateLimit = response.headers['ratelimit-remaining']
            if (Number(rateLimit) <= 0) {
                const rateLimitError = new Error('Gitlab rate limit exceeded.')
                Sentry.captureException(rateLimitError)
                throw rateLimitError
            }
            return response
        })
        .then(({ status }) => status !== 200)
        .catch(error => {
            console.warn('Failed to fetch repository visibility info.', error)
            return true
        })
}

export const parseHash = (hash: string): LineOrPositionOrRange => {
    if (hash.startsWith('#')) {
        hash = hash.slice(1)
    }

    if (!/^L\d+(-\d+)?$/.test(hash)) {
        return {}
    }

    const lpr = {} as LineOrPositionOrRange
    const [startString, endString] = hash.slice(1).split('-')

    lpr.line = parseInt(startString, 10)
    if (endString) {
        lpr.endLine = parseInt(endString, 10)
    }

    return lpr
}

export const gitlabCodeHost = subtypeOf<CodeHost>()({
    type: 'gitlab',
    name: 'GitLab',
    check: checkIsGitlab,
    codeViewResolvers: [codeViewResolver],
    adjustOverlayPosition,
    getCommandPaletteMount,
    getContext: async () => {
        const { repoName, ...pageInfo } = getPageInfo()
        return {
            ...pageInfo,
            privateRepository: await isPrivateRepository(repoName),
        }
    },
    urlToFile: (sourcegraphURL, target, context): string => {
        // A view state means that a panel must be shown, and panels are currently only supported on
        // Sourcegraph (not code hosts).
        // Make sure the location is also on this Gitlab instance, return an absolute URL otherwise.
        if (target.viewState || !target.rawRepoName.startsWith(window.location.hostname)) {
            return toAbsoluteBlobURL(sourcegraphURL, target)
        }

        // Stay on same page in MR if possible.
        // TODO to be entirely correct, this would need to compare the revision of the code view with the target revision.
        const currentPage = getPageInfo()
        if (currentPage.rawRepoName === target.rawRepoName && context.part !== undefined) {
            const codeViews = queryWithSelector(document.body, codeViewResolver.selector)
            for (const codeView of codeViews) {
                const { headFilePath, baseFilePath } = getFilePathsFromCodeView(codeView)
                if (headFilePath !== target.filePath && baseFilePath !== target.filePath) {
                    continue
                }
                if (!target.position) {
                    const url = new URL(window.location.href)
                    url.hash = codeView.id
                    return url.href
                }
                const partSelector = context.part !== null ? { head: '.new_line', base: '.old_line' }[context.part] : ''
                const link = codeView.querySelector<HTMLAnchorElement>(
                    `${partSelector} a[data-linenumber="${target.position.line}"]`
                )
                if (!link) {
                    break
                }
                return new URL(link.href).href
            }
        }

        // Go to specific URL on this Gitlab instance.
        const url = new URL(`https://${target.rawRepoName}/blob/${target.revision}/${target.filePath}`)
        if (target.position) {
            const { line } = target.position
            url.hash = `#L${line}`
        }
        return url.href
    },
    notificationClassNames,
    commandPaletteClassProps: {
        popoverClassName: classNames('dropdown-menu', styles.commandListPopover),
        formClassName: 'dropdown-input',
        inputClassName: 'dropdown-input-field',
        resultsContainerClassName: 'dropdown-content',
        selectedActionItemClassName: 'is-focused',
        noResultsClassName: 'px-3',
        iconClassName: 's16 align-bottom',
    },
    codeViewToolbarClassProps: {
        className: 'pl-0',
        actionItemClass: 'btn btn-md gl-button btn-icon',
        actionItemPressedClass: 'active',
        actionItemIconClass: 'gl-button-icon gl-icon s16',
    },
    hoverOverlayClassProps: {
        className: classNames('card', styles.hoverOverlay),
        actionItemClassName: 'btn btn-secondary',
        actionItemPressedClassName: 'active',
        iconClassName: 'square s16',
        getAlertClassName: createNotificationClassNameGetter(notificationClassNames),
    },
    codeViewsRequireTokenization: true,
    getHoverOverlayMountLocation: (): string | null => {
        const { pageKind } = getPageInfo()
        // On merge request pages only, mount the hover overlay to the diffs tab container.
        if (pageKind === GitLabPageKind.MergeRequest) {
            return 'div.tab-pane.diffs'
        }
        return null
    },
    // We listen to links clicks instead of 'hashchange' event as GitLab uses anchor links
    // to scroll to the selected line. Link click doesn't trigger 'hashchange' event
    // despite the URL hash is updated.
    observeLineSelection: fromEvent(document, 'click').pipe(
        filter(event => (event.target as HTMLElement).matches('a[data-line-number]')),
        map(() => parseHash(window.location.hash))
    ),
})
