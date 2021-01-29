import { compact, find, head } from 'lodash'
import { interval, Observable, Subject } from 'rxjs'
import { filter, map, refCount, publishReplay } from 'rxjs/operators'
import { MutationRecordLike } from '../../util/dom'
import { CodeHost } from '../shared/codeHost'
import { CodeView, DOMFunctions } from '../shared/codeViews'
import { queryWithSelector, ViewResolver, CustomSelectorFunction } from '../shared/views'

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
    selector(target: HTMLElement) {
        // The Gerrit mutation observer uses this selector to emit added nodes.
        // But `trackViews` calls this selector also, to check if an added node
        // contains a code view, even though in Gerrit this is redundant. When
        // the selector is called with an existing matching element, then it
        // must return that element (and only that one element) rather than all
        // matching elements on the page.
        if (target.matches('#diffTable')) {
            return [target]
        }

        const fileListElement = querySelectorAcrossShadowRoots(document, [
            '#app',
            '#app-element',
            'gr-change-view',
            '#fileList',
        ])
        // Usually each `.file-row` which is `.expanded` will have a
        // corresponding `diffTable` under a sibling, with the common parent
        // being `.stickArea`.
        const fileRows = fileListElement?.shadowRoot?.querySelectorAll('.file-row.expanded')
        const fileRowElements = [...(fileRows || [])]
        const diffTables = fileRowElements.map(fileRow => {
            const stickyArea = fileRow.parentElement
            if (!stickyArea) {
                return
            }
            return querySelectorAcrossShadowRoots(stickyArea, ['gr-diff-host', 'gr-diff', '#diffTable']) as HTMLElement
        })
        return compact(diffTables)
    },
    resolveView(diffTableElement: HTMLElement): CodeView | null {
        // Although we already obtained code view's element, `diffTable`, from
        // the `fileRow` in the selector function above, we have to revisit the
        // `fileRow` to get the file path.
        const stickyArea = closestParentAcrossShadowRoots(diffTableElement, '.stickyArea')

        if (!stickyArea) {
            return null
        }

        const fileRow = stickyArea.querySelector('.file-row')
        const filePath = fileRow?.getAttribute('data-path')
        if (!filePath) {
            return null
        }

        const gerritChange = parseGerritChange()
        const gerritChangeString = buildGerritChangeString(gerritChange)
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

const getLineElementFromLineNumber: DOMFunctions['getCodeElementFromLineNumber'] = (codeView, line, part) => {
    const side = part === 'head' ? 'right' : 'left'
    // Split diff: line element is next sibling
    const lineNumberCell = codeView.querySelector(`td.lineNum.${side}[data-value="${line}"]`)
    if (!lineNumberCell) {
        return null
    }
    const contentCell = nextMatchingSibling(lineNumberCell, 'td.content')
    const codeElement = contentCell?.querySelector('.contentText')
    // const lineRow = lineNumberCell?.closest('tr')
    // const codeElement = lineRow?.querySelector(`.contentText[data-side="${side}]"`)
    return codeElement as HTMLElement
}

const diffTableDomFunctions: DOMFunctions = {
    getLineElementFromLineNumber,
    getCodeElementFromLineNumber: getLineElementFromLineNumber,
    getCodeElementFromTarget: (target: HTMLElement) => {
        const codeElement = target.closest('td.content')
        // Check if we are on a line which has "File" in the line number cell.
        const fileElement = codeElement?.closest('tr')?.querySelector('[data-line-number]')
        if (fileElement) {
            const lineNumberString = fileElement.getAttribute('data-line-number')
            if (lineNumberString === 'FILE') {
                return null
            }
        }
        return codeElement as HTMLElement
    },
    getLineNumberFromCodeElement: (codeElement: HTMLElement): number => {
        const side = getSideFromCodeElement(codeElement)
        if (!side) {
            console.error(codeElement)
            throw new TypeError('Could not find line number (no side)')
        }
        const cell = codeElement.closest('td')
        const lineNumberCell = cell?.parentElement?.querySelector(`.lineNum.${side}`)
        const lineNumber = lineNumberCell?.getAttribute('data-value')
        if (!lineNumber) {
            throw new TypeError(`Could not find line number (${side})`)
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
    selector(target: HTMLElement) {
        // Because we expect only one code view to be present on an individual file
        // page, we don't have to consider the existing element (`target`) here in
        // the same way as in resolveFileListCodeView.
        const diffTableElement = querySelectorAcrossShadowRoots(document.body, [
            '#app',
            '#app-element',
            'main > gr-diff-view',
            '#diffHost',
            '#diff',
            '#diffTable',
        ])

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

/** Storing  */
interface ElementWithFilePath {
    element: HTMLElement
    filePath: string
}

const POLLING_INTERVAL = 4000
export const observeMutations = (
    target: Node,
    options?: MutationObserverInit,
    paused?: Subject<boolean>
): Observable<MutationRecordLike[]> => {
    let knownElements: ElementWithFilePath[] = []
    return interval(POLLING_INTERVAL).pipe(
        map(() => {
            const { filePath = '' } = parseGerritChange()
            const selectors = [...codeViewResolvers.map(resolver => resolver.selector), toolbarSelector]
            const foundElements = selectors.map(selector => [...queryWithSelector(document.body, selector)]).flat()
            const addedNodes = foundElements.filter(
                foundElement => !find(knownElements, { element: foundElement, filePath })
            )
            const removedNodes = knownElements.filter(
                knownElement =>
                    !(
                        knownElement.filePath === filePath &&
                        foundElements.find(foundElement => foundElement === knownElement.element)
                    )
            )

            // Add to known elements
            for (const addedNode of addedNodes) {
                knownElements.push({ element: addedNode, filePath })
            }
            // Remove from known elements
            knownElements = knownElements.filter(knownElement => !find(removedNodes, knownElement))

            console.log('knownElements', knownElements)
            console.log({ addedNodes, removedNodes })
            return { addedNodes, removedNodes: removedNodes.map(node => node.element) }
        }),
        // Filter to emit only non-empty records.
        filter(({ addedNodes, removedNodes }) => !!addedNodes.length || !!removedNodes.length),
        // Wrap in an array, because that's how mutation observers emit events.
        map(mutationRecord => [mutationRecord]),
        publishReplay(),
        refCount()
    )
}

const codeViewResolvers = [resolveFilePageCodeView, resolveFileListCodeView]
export const gerritCodeHost: CodeHost = {
    type: 'gerrit',
    name: 'Gerrit',
    codeViewResolvers,
    contentViewResolvers: [],
    textFieldResolvers: [],
    nativeTooltipResolvers: [],
    codeViewsRequireTokenization: true,
    // This overrides the default observeMutations because we need to handle shadow DOMS.
    observeMutations,
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
    viewOnSourcegraphButtonClassProps: {
        className: 'open-on-sourcegraph--gerrit',
        iconClassName: 'open-on-sourcegraph-icon--gerrit',
    },
    getViewContextOnSourcegraphMount: target => {
        const secondaryActionsElement = head(toolbarSelector(target))
        if (!secondaryActionsElement) {
            return null
        }
        console.log('secondaryActionsElement', secondaryActionsElement)
        const mount = document.createElement('gr-button')
        mount.setAttribute('link', 'link')
        mount.classList.add('OPEN_IN_SOURCEGRAPH')
        secondaryActionsElement.prepend(mount)
        secondaryActionsElement.append(createStyleElement(toolbarStyles))

        return mount
    },
}

const toolbarSelector: CustomSelectorFunction = () => {
    const toolbar = querySelectorAcrossShadowRoots(document.body, [
        '#app',
        '#app-element',
        'main > gr-change-view',
        '#actions',
        '#secondaryActions',
    ])
    return toolbar ? [toolbar as HTMLElement] : null
}

function getParentCommit(): string | null | undefined {
    const metadataPanel = querySelectorAcrossShadowRoots(document.body, [
        '#app',
        '#app-element',
        'gr-change-view',
        '#metadata',
        'gr-commit-info',
    ])
    if (!metadataPanel) {
        return null
    }
    return metadataPanel?.shadowRoot?.querySelector('.container')?.textContent?.trim()
}

function getDiffTypeFromCodeElement(codeElement: HTMLElement): 'side-by-side' | 'unified' {
    const rowElement = codeElement.closest('tr')
    if (rowElement?.matches('.unified')) {
        return 'unified'
    }
    return 'side-by-side'
}

function getSideFromCodeElement(codeElement: HTMLElement): 'left' | 'right' {
    const diffType = getDiffTypeFromCodeElement(codeElement)
    if (diffType === 'unified') {
        // If there is a line number cell for the right side, in this row, then we're on the right.
        // Otherwise we're on the left.
        const rightLineNumber = codeElement.closest('tr')?.querySelector('td.right')?.getAttribute('data-value')

        if (rightLineNumber) {
            return 'right'
        }
        return 'left'
    }

    // Side-by-side
    if (codeElement.closest('td')?.previousElementSibling?.matches('.right')) {
        return 'right'
    }
    return 'left'
}

function nextMatchingSibling(element: Element, selector: string): HTMLElement | null {
    const allSiblings = [...(element.parentElement?.childNodes || [])]
    const nextSiblings = allSiblings.slice(allSiblings.indexOf(element) + 1)
    for (const sibling of nextSiblings) {
        if (sibling instanceof HTMLElement) {
            if (sibling.matches(selector)) {
                return sibling
            }
        }
    }
    return null
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
    console.log('Created style element', styleElement)
    return styleElement
}

const toolbarStyles = `
.icon--gerrit {
    height: 1.3rem;
    width: 1.3rem;
    padding: 5px 4px;
}
.open-on-sourcegraph-icon--gerrit {
    height: 1rem;
    width: 1rem;
}
.open-on-sourcegraph--gerrit {
    text-decoration: none;
}
`
