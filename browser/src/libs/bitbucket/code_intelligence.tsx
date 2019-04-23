import { AdjustmentDirection, DOMFunctions, PositionAdjuster } from '@sourcegraph/codeintellify'
import { of } from 'rxjs'
import { Omit } from 'utility-types'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../../../../shared/src/util/url'
import { querySelectorOrSelf } from '../../shared/util/dom'
import { CodeHost, MountGetter } from '../code_intelligence'
import { CodeView } from '../code_intelligence/code_views'
import { ViewResolver } from '../code_intelligence/views'
import { getContext } from './context'
import { diffDOMFunctions, singleFileDOMFunctions } from './dom_functions'
import {
    resolveCommitViewFileInfo,
    resolveCompareFileInfo,
    resolveFileInfoForSingleFileSourceView,
    resolvePullRequestFileInfo,
    resolveSingleFileDiffFileInfo,
} from './file_info'
import { isCommitsView, isCompareView, isPullRequestView, isSingleFileView } from './scrape'

/**
 * Gets or creates the toolbar mount for allcode views.
 */
export const getToolbarMount = (codeView: HTMLElement): HTMLElement => {
    const existingMount = codeView.querySelector<HTMLElement>('.sg-toolbar-mount')
    if (existingMount) {
        return existingMount
    }

    const fileActions = codeView.querySelector<HTMLElement>('.file-toolbar .secondary')
    if (!fileActions) {
        throw new Error('Unable to find mount location')
    }

    fileActions.style.display = 'flex'

    const mount = document.createElement('div')
    mount.classList.add('btn-group')
    mount.classList.add('sg-toolbar-mount')
    mount.classList.add('sg-toolbar-mount-bitbucket-server')

    fileActions.insertAdjacentElement('afterbegin', mount)

    return mount
}

/**
 * Sometimes tabs are converted to spaces so we need to adjust. Luckily, there
 * is an attribute `cm-text` that contains the real text.
 */
const createPositionAdjuster = (dom: DOMFunctions) => (
    queryGraphQL: PlatformContext['queryGraphQL']
): PositionAdjuster<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec> => ({ direction, codeView, position }) => {
    const codeElement = dom.getCodeElementFromLineNumber(codeView, position.line, position.part)
    if (!codeElement) {
        throw new Error('(adjustPosition) could not find code element for line provided')
    }

    let delta = 0
    for (const modifiedTextElem of codeElement.querySelectorAll('[cm-text]')) {
        const actualText = modifiedTextElem.getAttribute('cm-text') || ''
        const adjustedText = modifiedTextElem.textContent || ''

        delta += actualText.length - adjustedText.length
    }

    const modifier = direction === AdjustmentDirection.ActualToCodeView ? -1 : 1

    const newPos = {
        line: position.line,
        character: position.character + modifier * delta,
    }

    return of(newPos)
}

const toolbarButtonProps = {
    className: 'aui-button',
}

/**
 * A code view spec for single file code view in the "source" view (not diff).
 */
const singleFileSourceCodeView: Omit<CodeView, 'element'> = {
    getToolbarMount,
    dom: singleFileDOMFunctions,
    resolveFileInfo: resolveFileInfoForSingleFileSourceView,
    getPositionAdjuster: createPositionAdjuster(singleFileDOMFunctions),
    toolbarButtonProps,
}

const baseDiffCodeView: Omit<CodeView, 'element' | 'resolveFileInfo'> = {
    getToolbarMount,
    dom: diffDOMFunctions,
    adjustPosition: createPositionAdjuster(diffDOMFunctions),
    toolbarButtonProps,
}
/**
 * A code view spec for a single file "diff to previous" view
 */
const singleFileDiffCodeView: Omit<CodeView, 'element'> = {
    ...baseDiffCodeView,
    resolveFileInfo: resolveSingleFileDiffFileInfo,
}

/**
 * A code view spec for pull requests
 */
const pullRequestDiffCodeView: Omit<CodeView, 'element'> = {
    ...baseDiffCodeView,
    resolveFileInfo: resolvePullRequestFileInfo,
}

/**
 * A code view spec for compare pages
 */
const compareDiffCodeView: Omit<CodeView, 'element'> = {
    ...baseDiffCodeView,
    resolveFileInfo: resolveCompareFileInfo,
}

/**
 * A code view spec for commit pages
 */
const commitDiffCodeView: Omit<CodeView, 'element'> = {
    ...baseDiffCodeView,
    resolveFileInfo: resolveCommitViewFileInfo,
}

const codeViewResolver: ViewResolver<CodeView> = {
    selector: '.file-content',
    resolveView: element => {
        const contentView = element.querySelector('.content-view')
        if (!contentView) {
            return null
        }
        if (isCompareView()) {
            return { element, ...compareDiffCodeView }
        }
        if (isCommitsView()) {
            return { element, ...commitDiffCodeView }
        }
        if (isSingleFileView(element)) {
            const isDiff = contentView.classList.contains('diff-view')
            return isDiff ? { element, ...singleFileDiffCodeView } : { element, ...singleFileSourceCodeView }
        }
        if (isPullRequestView()) {
            return { element, ...pullRequestDiffCodeView }
        }
        console.error('Unknown code view', element)
        return null
    },
}

const getCommandPaletteMount: MountGetter = (container: HTMLElement): HTMLElement | null => {
    const headerElement = querySelectorOrSelf(container, '.aui-header-primary .aui-nav')
    if (!headerElement) {
        return null
    }
    const classes = ['command-palette-button', 'command-palette-button--bitbucket-server']
    const create = (): HTMLElement => {
        const mount = document.createElement('li')
        mount.className = classes.join(' ')
        headerElement.insertAdjacentElement('beforeend', mount)
        return mount
    }
    const preexisting = headerElement.querySelector<HTMLElement>(classes.map(c => `.${c}`).join(''))
    return preexisting || create()
}

function getViewContextOnSourcegraphMount(container: HTMLElement): HTMLElement | null {
    const branchSelectorButtons = querySelectorOrSelf(container, '.branch-selector-toolbar .aui-buttons')
    if (!branchSelectorButtons) {
        return null
    }
    const preexisting = branchSelectorButtons.querySelector<HTMLElement>('#open-on-sourcegraph')
    if (preexisting) {
        return preexisting
    }
    const mount = document.createElement('span')
    mount.id = 'open-on-sourcegraph'
    mount.className = 'open-on-sourcegraph--bitbucket-server'
    branchSelectorButtons.insertAdjacentElement('beforeend', mount)
    return mount
}

export const checkIsBitbucket = (): boolean =>
    !!document.querySelector('.bitbucket-header-logo') ||
    !!document.querySelector('.aui-header-logo.aui-header-logo-bitbucket')

export const bitbucketServerCodeHost: CodeHost = {
    name: 'bitbucket-server',
    check: checkIsBitbucket,
    codeViewResolvers: [codeViewResolver],
    getCommandPaletteMount,
    commandPaletteClassProps: {
        popoverClassName: 'searchable-selector command-palette-popover--bitbucket-server',
        resultsContainerClassName: 'results',
        listClassName: 'results-list',
        listItemClassName: 'result',
        selectedListItemClassName: 'focused',
        noResultsClassName: 'no-results',
    },
    codeViewToolbarClassProps: {
        className: 'aui-buttons',
        actionItemClass: 'aui-button action-item--bitbucket-server',
        // actionItemPressedClass is not needed because Bitbucket applies styling to aria-pressed="true"
        actionItemIconClass: 'aui-icon',
        listItemClass: 'action-nav-item--bitbucket',
    },
    hoverOverlayClassProps: {
        actionItemClassName: 'aui-button hover-action-item--bitbucket-server',
        closeButtonClassName: 'aui-button',
    },
    getViewContextOnSourcegraphMount,
    getContext,
    viewOnSourcegraphButtonClassProps: {
        className: 'aui-button',
        iconClassName: 'aui-icon',
    },
}
