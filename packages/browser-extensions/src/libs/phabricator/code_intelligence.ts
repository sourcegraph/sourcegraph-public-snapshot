import { AdjustmentDirection, DiffPart, PositionAdjuster } from '@sourcegraph/codeintellify'
import { map } from 'rxjs/operators'
import { Position } from 'vscode-languageserver-types'
import { convertSpacesToTabs, spacesToTabsAdjustment } from '.'
import storage from '../../browser/storage'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { CodeHost, CodeView } from '../code_intelligence'
import { diffDomFunctions, diffusionDOMFns, getLineRanges } from './dom_functions'
import { resolveDiffFileInfo, resolveDiffusionFileInfo } from './file_info'

function createMount(
    findMountLocation: (file: HTMLElement, part?: DiffPart) => HTMLElement
): (file: HTMLElement, part: DiffPart) => HTMLElement {
    return (file: HTMLElement, part: DiffPart) => {
        const className = 'sourcegraph-app-annotator' + (part === 'base' ? '-base' : '')
        const existingMount = file.querySelector('.' + className)
        if (existingMount) {
            // Make this function idempotent; no need to create a mount twice.
            return existingMount as HTMLElement
        }

        const mount = document.createElement('div')
        mount.style.display = 'inline-block'
        mount.classList.add(className)

        const mountLocation = findMountLocation(file, part)
        mountLocation.appendChild(mount)

        return mount
    }
}

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

const adjustPosition: PositionAdjuster = ({ direction, codeView, position }) =>
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
    className: 'button button-grey has-icon has-text phui-button-default msl',
    iconStyle: { marginTop: '-1px', paddingRight: '4px', fontSize: '18px', height: '.8em', width: '.8em' },
    style: {},
}

export const phabCodeViews: CodeView[] = [
    {
        selector: '.differential-changeset',
        dom: diffDomFunctions,
        resolveFileInfo: resolveDiffFileInfo,
        adjustPosition,
        getToolbarMount: createMount(file => {
            const actionLinks = file.querySelector('.differential-changeset-buttons')
            if (!actionLinks) {
                throw new Error('Unable to find action links for changeset')
            }

            return actionLinks as HTMLElement
        }),
        toolbarButtonProps,
        getLineRanges,
        isDiff: true,
    },
    {
        selector: '.phabricator-source-code-container',
        dom: diffusionDOMFns,
        resolveFileInfo: resolveDiffusionFileInfo,
        getToolbarMount: createMount(() => {
            const actionLinks = document.querySelector('.diffusion-action-bar .phui-right-view')
            if (!actionLinks) {
                throw new Error('Unable to find action links for diffusion')
            }

            return actionLinks as HTMLElement
        }),
        toolbarButtonProps,
        getLineRanges,
        isDiff: false,
    },
]

function checkIsPhabricator(): Promise<boolean> {
    if (document.querySelector('.phabricator-wordmark')) {
        return Promise.resolve(true)
    }

    return new Promise<boolean>(resolve =>
        storage.getSync(items => resolve(!!items.enterpriseUrls.find(url => url === window.location.origin)))
    )
}

export const phabricatorCodeHost: CodeHost = {
    codeViews: phabCodeViews,
    name: 'phabricator',
    check: checkIsPhabricator,
}
