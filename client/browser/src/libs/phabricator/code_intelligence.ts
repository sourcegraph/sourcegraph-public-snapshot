import { AdjustmentDirection, PositionAdjuster } from '@sourcegraph/codeintellify'
import { Position } from '@sourcegraph/extension-api-types'
import { of } from 'rxjs'
import { map } from 'rxjs/operators'
import { convertSpacesToTabs, spacesToTabsAdjustment } from '.'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../../../../../shared/src/util/url'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { CodeHost, CodeViewSpec, CodeViewSpecResolver, CodeViewSpecWithOutSelector } from '../code_intelligence'
import { diffDomFunctions, diffusionDOMFns } from './dom_functions'
import { resolveDiffFileInfo, resolveDiffusionFileInfo, resolveRevisionFileInfo } from './file_info'

/**
 * Gets the actual text content we care about and returns the number of characters we have stripped
 * so that we can adjust accordingly.
 */
const getTextContent = (element: HTMLElement): { textContent: string; adjust: number } => {
    let textContent = element.textContent || ''
    let adjust = 0

    // For some reason, phabricator adds an invisible element to the beginning of lines containing the diff indicator
    // followed by a space (ex: '+ '). We need to adjust the position accordingly.
    if (element.firstElementChild && element.firstElementChild.classList.contains('aural-only')) {
        const pre = element.firstElementChild.textContent || ''
        // Codeintellify handles ignoring one character for diff indicators so we'll allow it to adjust for that.
        adjust = pre.replace(/^(\+|-)/, '').length

        // Get rid of the characters we have adjusted for.
        textContent = textContent.substr(pre.length - adjust)
    }

    // Phabricator adds a no-width-space to the beginning of the line in some cases.
    // We need to strip that and account for it here.
    if (textContent.charCodeAt(0) === 8203) {
        textContent = textContent.substr(1)
        adjust++
    }

    return { textContent, adjust }
}

const adjustCharacter = (position: Position, adjustment: number): Position => ({
    line: position.line,
    character: position.character + adjustment,
})

const adjustPosition: PositionAdjuster<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec> = ({
    direction,
    codeView,
    position,
}) =>
    fetchBlobContentLines(position).pipe(
        map(lines => {
            const codeElement = diffDomFunctions.getCodeElementFromLineNumber(codeView, position.line, position.part)
            if (!codeElement) {
                throw new Error('(adjustPosition) could not find code element for line provided')
            }

            const textContentInfo = getTextContent(codeElement)

            const documentLineContent = textContentInfo.textContent
            const actualLineContent = lines[position.line - 1]

            // See if we should adjust for whitespace changes.
            const convertSpaces = convertSpacesToTabs(actualLineContent, documentLineContent)

            // Whether the adjustment should add or subtract from the given position.
            const modifier = direction === AdjustmentDirection.CodeViewToActual ? -1 : 1

            return convertSpaces
                ? adjustCharacter(
                      position,
                      (spacesToTabsAdjustment(documentLineContent) + textContentInfo.adjust) * modifier
                  )
                : adjustCharacter(position, textContentInfo.adjust * modifier)
        })
    )

const toolbarButtonProps = {
    className: 'button grey button-grey has-icon has-text phui-button-default msl',
}
const commitCodeView: CodeViewSpecWithOutSelector = {
    dom: diffDomFunctions,
    resolveFileInfo: resolveRevisionFileInfo,
    adjustPosition,
    getToolbarMount: codeView => {
        const actions = codeView.querySelector('.differential-changeset-buttons')
        if (!actions) {
            throw new Error('Unable to find action links for revision')
        }

        const mount = document.createElement('div')
        mount.style.display = 'inline-block'
        mount.classList.add('sourcegraph-app-annotator')

        actions.insertAdjacentElement('afterbegin', mount)

        return mount
    },
    toolbarButtonProps,
    isDiff: true,
}

export const diffCodeView = {
    dom: diffDomFunctions,
    resolveFileInfo: resolveDiffFileInfo,
    adjustPosition,
    getToolbarMount: (codeView: HTMLElement): HTMLElement => {
        const className = 'sourcegraph-app-annotator'
        const existingMount = codeView.querySelector<HTMLElement>('.' + className)
        if (existingMount) {
            return existingMount
        }
        const mountLocation = codeView.querySelector('.differential-changeset-buttons')
        if (!mountLocation) {
            throw new Error('Unable to find action links for changeset')
        }
        const mount = document.createElement('div')
        mount.style.display = 'inline-block'
        mount.classList.add(className)
        mountLocation.prepend(mount, ' ')
        return mount
    },
    toolbarButtonProps,
    isDiff: true,
}

const codeViewSpecResolver: CodeViewSpecResolver = {
    selector: '.differential-changeset',
    resolveCodeViewSpec: (codeView: HTMLElement): CodeViewSpecWithOutSelector => {
        if (window.location.pathname.match(/^\/r/)) {
            return commitCodeView
        }
        return diffCodeView
    },
}

const phabCodeViews: CodeViewSpec[] = [
    {
        // TODO this code view does not include the toolbar,
        // which makes it not possible to test getToolbarMount()
        // Fix after https://github.com/sourcegraph/sourcegraph/issues/3271
        selector: '.diffusion-source',
        dom: diffusionDOMFns,
        resolveFileInfo: resolveDiffusionFileInfo,
        getToolbarMount: () => {
            const actions = document.querySelector<HTMLElement>('.phui-two-column-content .phui-header-action-links')
            if (!actions) {
                throw new Error('unable to find file actions')
            }

            const mount = document.createElement('div')
            mount.style.display = 'inline-block'
            mount.classList.add('sourcegraph-app-annotator')

            actions.insertAdjacentElement('afterbegin', mount)

            return mount
        },
        toolbarButtonProps,
        isDiff: false,
    },
]

export const checkIsPhabricator = () => !!document.querySelector('.phabricator-wordmark')

export const phabricatorCodeHost: CodeHost = {
    codeViewSpecs: phabCodeViews,
    codeViewSpecResolver,
    name: 'phabricator',
    check: checkIsPhabricator,

    // TODO: handle parsing selected line number from Phabricator href,
    // and find a way to listen to changes (Phabricator does not emit popstate events).
    selectionsChanges: () => of([]),
    codeViewToolbarClassProps: {
        actionItemClass: 'button grey action-item--phabricator',
        actionItemIconClass: 'action-item__icon--phabricator',
    },
    hoverOverlayClassProps: {
        actionItemClassName: 'button grey hover-overlay-action-item--phabricator',
    },
}
