import { CodeHost } from '../shared/codeHost'
import { CodeView } from '../shared/codeViews'
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

const resolveFileCodeView: ViewResolver<CodeView> = {
    // The selector is used not only to select, but to filter.
    selector: 'div,body',
    resolveView(element: HTMLElement): CodeView | null {
        console.log('Sourcegraph: resolveView is running')
        const diffTableElement = document.body
            .querySelector('#app')
            ?.shadowRoot?.querySelector('#app-element')
            ?.shadowRoot?.querySelector('main > gr-diff-view')
            ?.shadowRoot?.querySelector('#diffHost')
            ?.shadowRoot?.querySelector('#diff')
            ?.shadowRoot?.querySelector('#diffTable')

        if (!diffTableElement) {
            return null
        }

        console.log('Sourcegraph: Found diff', diffTableElement)
        const gerritChange = parseGerritChange()
        const gerritChangeString = buildGerritChangeString(gerritChange)
        const parent = getParentCommit()
        console.log('Sourcegraph: gerrit diff:', { gerritChange, gerritChangeString, parent })

        return {
            element: diffTableElement as HTMLElement,
            dom: {
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
                        console.error('getLineNumberFromCodeElement: could not get side')
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
            },
            resolveFileInfo() {
                return {
                    base: {
                        filePath: gerritChange.filePath || '',
                        rawRepoName: gerritChange.repoName,
                        commitID: gerritChangeString + '^', // Does this
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
    codeViewResolvers: [resolveFileCodeView],
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
}

function getParentCommit(): string | null | undefined {
    const metadataPanel = document
        .querySelector('#app')
        ?.shadowRoot?.querySelector('#app-element')
        ?.shadowRoot?.querySelector('main > gr-change-view')
        ?.shadowRoot?.querySelector('#metadata')
        ?.shadowRoot?.querySelector('gr-commit-info')
    if (!metadataPanel) {
        return null
    }
    return metadataPanel?.shadowRoot?.querySelector('.container')?.textContent
}

function getSideFromCodeElement(codeElement: HTMLElement): string | undefined {
    return codeElement.querySelector('.contentText')?.getAttribute('data-side') || undefined
}
