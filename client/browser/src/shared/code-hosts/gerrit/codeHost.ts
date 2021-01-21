import { CodeHost } from '../shared/codeHost'
import { CodeView, DOMFunctions } from '../shared/codeViews'
import { ViewResolver } from '../shared/views'

function checkIsGerrit(): boolean {
    const isGerrit = !!document.querySelector('gr-app#app')
    return isGerrit
}

interface GerritChangeAndPatchSet {
    repoName: string
    changeId: string
    patchSetId: string
    filePath?: string
}

function buildGerritChangeString({ changeId, patchSetId }: GerritChangeAndPatchSet): string {
    const changeDirectoryPrefix = changeId.slice(0, 2).padStart(2, '0')
    patchSetId = patchSetId || '1' // Default patch set if it's not provided.
    return `refs/changes/${changeDirectoryPrefix}/${changeId}/${patchSetId}`
}

function parseGerritChange(): GerritChangeAndPatchSet {
    const path = new URL(window.location.href).pathname
    const pathParts = path.split('/')
    const cPart = pathParts.indexOf('c')
    const repoName = pathParts[cPart + 1]
    const plusSign = pathParts.indexOf('+')
    const changeId = pathParts[plusSign + 1]
    const patchSetId = pathParts[plusSign + 2]
    const filePath = pathParts.slice(plusSign + 3).join('/')
    const repoNameWithServer = window.location.hostname + '/' + repoName
    return { repoName: repoNameWithServer, changeId, patchSetId, filePath }
}

const resolveFileListCodeView: ViewResolver<CodeView> = {
    selector() {
        const fileListElement = querySelectorAcrossShadowRoots(document, [
            '#app',
            '#app-element',
            'gr-change-view',
            '#fileList',
        ])
        const fileRows = fileListElement?.shadowRoot?.querySelectorAll('.file-row.expanded')
        return fileRows
    },
    resolveView(element: HTMLElement): CodeView | null {
        const stickyArea = element.parentElement
        if (!stickyArea) {
            return null
        }
        // From .file-row to diff-table
        const diffTableElement = querySelectorAcrossShadowRoots(stickyArea, [
            'gr-diff-host',
            'gr-diff',
            'table',
        ]) as HTMLElement
        if (!diffTableElement) {
            return null
        }

        // Get the file and revision context
        const gerritChange = parseGerritChange()
        const gerritChangeString = buildGerritChangeString(gerritChange)
        const filePath = element.getAttribute('data-path')
        const parentCommit = getParentCommit() || ''

        return {
            element: diffTableElement,
            dom: diffTableDomFunctions,
            resolveFileInfo() {
                return {
                    base: {
                        filePath: filePath || '',
                        rawRepoName: gerritChange.repoName,
                        commitID: parentCommit,
                    },
                    head: {
                        filePath: filePath || '',
                        rawRepoName: gerritChange.repoName,
                        commitID: gerritChangeString,
                    },
                }
            },
        }
    },
}

const diffTableDomFunctions: DOMFunctions = {
    getLineElementFromLineNumber: (codeView, line, part) => {
        const side = part === 'head' ? 'right' : 'left'
        // Split diff: line element is next sibling
        const lineNumberCell = codeView.querySelector(`td.lineNum.${side}[data-value="${line}"]`)
        const lineRow = lineNumberCell?.closest('tr')
        const codeElement = lineRow?.querySelector(`.contentText[data-side="${side}"`)
        return codeElement?.closest('tr') as HTMLElement
    },
    getCodeElementFromLineNumber: (codeView, line, part) => {
        const side = part === 'head' ? 'right' : 'left'
        const lineNumberCell = codeView.querySelector(`td.lineNum.${side}[data-value="${line}"]`)
        const lineRow = lineNumberCell?.closest('tr')
        const codeElement = lineRow?.querySelector(`.contentText[data-side="${side}"`)
        return codeElement as HTMLElement
    },
    getCodeElementFromTarget: (target: HTMLElement) => target.closest('td.content'),
    getLineNumberFromCodeElement: (codeElement: HTMLElement): number => {
        const side = getSideFromCodeElement(codeElement)
        if (!side) {
            throw new TypeError('Could not find line number')
        }
        const cell = codeElement.closest('td')
        const lineNumberCell = cell?.parentElement?.querySelector(`.lineNum.${side}`)
        const lineNumber = lineNumberCell?.getAttribute('data-value')
        if (!lineNumber) {
            throw new TypeError('Could not find line number')
        }
        return parseInt(lineNumber, 10)
    },
    getDiffCodePart: codeElement => {
        const side = getSideFromCodeElement(codeElement)
        if (side === 'left') {
            return 'base'
        }
        if (side === 'right') {
            return 'head'
        }
        console.error('Cannot tell: base or head')
        return 'base'
    },
}

const resolveFilePageCodeView: ViewResolver<CodeView> = {
    selector() {
        // TODO: rewrite query using querySelectorAcrossShadowRoots
        const diffTableElement = document.body
            .querySelector('#app')
            ?.shadowRoot?.querySelector('#app-element')
            ?.shadowRoot?.querySelector('main > gr-diff-view')
            ?.shadowRoot?.querySelector('#diffHost')
            ?.shadowRoot?.querySelector('#diff')
            ?.shadowRoot?.querySelector('#diffTable')

        if (diffTableElement) {
            return [diffTableElement as HTMLElement]
        }
        return []
    },
    resolveView(element: HTMLElement): CodeView | null {
        const gerritChange = parseGerritChange()
        const gerritChangeString = buildGerritChangeString(gerritChange)
        let parent = getParentCommit() || gerritChangeString + '^'
        // Possible situation: we cannot get the parent commit on the page
        if (!parent) {
            // TODO: determine if this fallback works.
            parent = gerritChangeString + '^'
        }

        return {
            element,
            dom: diffTableDomFunctions,
            getToolbarMount(codeView) {
                const subheaderElement = getSubheaderFromCodeView(codeView)?.querySelector('.patchRangeLeft')
                if (!subheaderElement) {
                    throw new Error('Could not find subheader')
                }
                const mountElement = document.createElement('div')
                subheaderElement.append(mountElement)
                subheaderElement.append(createStyleElement(toolbarStyles))
                return mountElement
            },
            resolveFileInfo() {
                return {
                    base: {
                        filePath: gerritChange.filePath || '',
                        rawRepoName: gerritChange.repoName,
                        commitID: parent,
                    },
                    head: {
                        filePath: gerritChange.filePath || '',
                        rawRepoName: gerritChange.repoName,
                        commitID: gerritChangeString,
                    },
                }
            },
        }
    },
}

export const gerritCodeHost: CodeHost = {
    type: 'gerrit',
    name: 'Gerrit',
    codeViewResolvers: [resolveFilePageCodeView, resolveFileListCodeView],
    contentViewResolvers: [],
    textFieldResolvers: [],
    nativeTooltipResolvers: [],
    codeViewsRequireTokenization: true,
    getContext() {
        const gerritChange = parseGerritChange()
        const gerritChangeString = buildGerritChangeString(gerritChange)
        return {
            privateRepository: true, // Gerrit is always private. Despite the fact that permissions can be set to be publicly viewable.
            rawRepoName: gerritChange.repoName,
            revision: gerritChangeString,
        }
    },
    check: checkIsGerrit,
    notificationClassNames: { 1: '', 2: '', 3: '', 4: '', 5: '' },
    hoverOverlayClassProps: {
        className: 'hover-overlay--gerrit',
    },
    codeViewToolbarClassProps: {
        className: 'code-view-toolbar--gerrit',
        actionItemIconClass: 'icon--gerrit',
    },
}

function getParentCommit(): string | null | undefined {
    // TODO: rewrite using querySelectorAcrossShadowRoots
    const metadataPanel = document
        .querySelector('#app')
        ?.shadowRoot?.querySelector('#app-element')
        ?.shadowRoot?.querySelector('main > gr-change-view')
        ?.shadowRoot?.querySelector('#metadata')
        ?.shadowRoot?.querySelector('gr-commit-info')
    if (!metadataPanel) {
        return null
    }
    return metadataPanel?.shadowRoot?.querySelector('.container')?.textContent?.trim()
}

function getSideFromCodeElement(codeElement: HTMLElement): string | undefined {
    return codeElement.querySelector('.contentText')?.getAttribute('data-side') || undefined
}

function querySelectorAcrossShadowRoots(element: ParentNode, selectors: string[]): Element | null {
    let currentElement: ParentNode | null = element
    const selectorsExceptLast = selectors.slice(0, -1)
    const lastSelector = selectors[selectors.length - 1]
    for (const selector of selectorsExceptLast) {
        if (!currentElement) {
            return null
        }
        currentElement = currentElement.querySelector(selector)?.shadowRoot || null
    }
    return currentElement?.querySelector(lastSelector) as Element
}

function closestParentAcrossShadowRoots(element: Element, selector: string): Element | null {
    while (true) {
        const result = element.closest(selector)
        if (result) {
            return result
        }
        element = (element.getRootNode() as ShadowRoot)?.host
        if (!element || element === document.getRootNode()) {
            return null
        }
    }
}

function getSubheaderFromCodeView(codeView: HTMLElement): HTMLElement | null | undefined {
    const grDiffView = closestParentAcrossShadowRoots(codeView, 'gr-diff-view')
    const subheader = grDiffView?.shadowRoot?.querySelector('gr-fixed-panel .subHeader') as HTMLElement
    return subheader
}

function createStyleElement(styles: string): HTMLStyleElement {
    const styleElement = document.createElement('style')
    styleElement.textContent = styles
    return styleElement
}

const toolbarStyles = `
.icon--gerrit {
    height: 1.3rem;
    width: 1.3rem;
    padding: 5px 4px;
}
`
