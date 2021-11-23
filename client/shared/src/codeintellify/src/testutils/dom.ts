import githubCode from '../../testdata/generated/github.html'
import sourcegraphCode from '../../testdata/generated/sourcegraph.html'
import { DOMFunctions } from '../token_position'
import { TEST_DATA_REVSPEC } from './rev'

const createElementFromString = (html: string): HTMLDivElement => {
    const elem = document.createElement('div')

    elem.innerHTML = html
    elem.style.height = 'auto'
    elem.style.width = 'auto'
    elem.style.whiteSpace = 'pre'
    elem.style.cssFloat = 'left'
    elem.style.display = 'block'
    elem.style.clear = 'both'

    return elem
}

/** Gets all of the text nodes within a given node. Used for testing. */
export const getTextNodes = (node: Node): Node[] => {
    if (node.childNodes.length === 0 && node.TEXT_NODE === node.nodeType && node.nodeValue) {
        return [node]
    }

    const nodes: Node[] = []

    for (const child of Array.from(node.childNodes)) {
        nodes.push(...getTextNodes(child))
    }

    return nodes
}

/** The props used for the generated test cases(e.g. GitHub and Sourcegraph flavored dom). */
export interface CodeViewProps extends DOMFunctions {
    /** The code view for the given test case(e.g. a <code> element in Sourcegraph and <table> in GitHub) */
    codeView: HTMLElement
    /** The container of the code view. (e.g. The scrollable contaienr in Sourcegraph and a parent <div> in GitHub) */
    container: HTMLElement
    /** The revision and repository information for the file used in the generated test cases. */
    revSpec: typeof TEST_DATA_REVSPEC
}

// BEGIN setup test cases

// Abstract implementation for GitHub and Sourcegraph. Could potentially be sufficient for any code host
// but we may want to keep this as a configuration point.
// Commented out cause we only have tests for non-diff code views so far
// const getDiffCodePart = (codeElement: HTMLElement): DiffPart => {
//     switch (codeElement.textContent!.charAt(0)) {
//         case '+':
//             return 'head'
//         case '-':
//             return 'base'
//         default:
//             return null
//     }
// }

const createGitHubCodeView = (): CodeViewProps => {
    const codeView = document.createElement('div')

    codeView.innerHTML = githubCode
    codeView.style.clear = 'both'

    const getCodeElementFromTarget = (target: HTMLElement): HTMLElement | null => {
        const row = target.closest('tr')
        if (!row) {
            return null
        }

        const codeCell = row.children.item(1) as HTMLElement

        if (!codeCell.classList.contains('blob-code')) {
            // Line element mouse overs probably
            return null
        }

        return codeCell
    }

    const getCodeElementFromLineNumber = (b: HTMLElement, line: number): HTMLElement | null => {
        const numCell = b.querySelector(`[data-line-number="${line}"]`)
        if (!numCell) {
            return null
        }

        const row = numCell.closest('tr') as HTMLElement
        if (!row) {
            return row
        }

        return row.children.item(1) as HTMLElement | null
    }

    const getLineNumberFromCodeElement = (codeCell: HTMLElement): number => {
        const row = codeCell.closest('tr')
        if (!row) {
            return -1
        }
        const numCell = row.children.item(0) as HTMLElement
        if (!numCell || (numCell && !numCell.dataset.lineNumber)) {
            return -1
        }

        return parseInt(numCell.dataset.lineNumber as string, 10)
    }

    return {
        container: codeView,
        codeView,

        revSpec: TEST_DATA_REVSPEC,
        getCodeElementFromTarget,
        getCodeElementFromLineNumber,
        getLineNumberFromCodeElement,
    }
}

const createSourcegraphCodeView = (): CodeViewProps => {
    const codeView = document.createElement('div')

    codeView.innerHTML = sourcegraphCode
    codeView.style.clear = 'both'

    const getCodeElementFromTarget = (target: HTMLElement): HTMLElement | null => {
        const row = target.closest('tr')
        if (!row) {
            return null
        }

        const codeCell = row.children.item(1) as HTMLElement

        if (!codeCell.classList.contains('code')) {
            // Line element mouse overs probably
            return null
        }

        return codeCell
    }

    const getCodeElementFromLineNumber = (b: HTMLElement, line: number): HTMLElement | null => {
        const numCell = b.querySelector(`[data-line="${line}"]`)
        if (!numCell) {
            return null
        }

        const row = numCell.closest('tr') as HTMLElement
        if (!row) {
            return row
        }

        return row.children.item(1) as HTMLElement | null
    }

    const getLineNumberFromCodeElement = (codeCell: HTMLElement): number => {
        const row = codeCell.closest('tr')
        if (!row) {
            return -1
        }

        const numCell = row.children.item(0) as HTMLElement
        if (!numCell || (numCell && !numCell.dataset.line)) {
            return -1
        }

        return parseInt(numCell.dataset.line as string, 10)
    }

    return {
        container: codeView,

        codeView: codeView.querySelector('code')!,
        revSpec: TEST_DATA_REVSPEC,
        getCodeElementFromTarget,
        getCodeElementFromLineNumber,
        getLineNumberFromCodeElement,
    }
}

// END setup test cases

/**
 * DOM is a testing utility class that keeps track of all elements a test suite is adding to the DOM
 * so that we can clean up after the test suite has finished.
 */
export class DOM {
    /** The inserted nodes. We save them so that we can remove them on cleanup. */
    private readonly nodes = new Set<Element>()

    /**
     * Creates and inserts the generated test cases into the DOM
     *
     * @returns the CodeViewProps for the test cases added to the DOM.
     */
    public createCodeViews(): CodeViewProps[] {
        const codeViews: CodeViewProps[] = [createSourcegraphCodeView(), createGitHubCodeView()]

        for (const { container } of codeViews) {
            this.insert(container)
        }

        return codeViews
    }

    /**
     * Creates a div with some arbitrary content.
     *
     * @param html the content of the element you wish to create.
     * @returns the created div.
     */
    public createElementFromString(html: string): HTMLDivElement {
        const element = createElementFromString(html)
        this.insert(element)
        return element
    }

    /**
     * Removes all nodes that were inserted from the DOM. This should be called after a test suite has ran.
     */
    public cleanup = (): void => {
        for (const node of this.nodes) {
            document.body.removeChild(node)
        }
    }

    /** The funnel for inserting elements into the DOM so that we know to remove it in `cleanup()`. */
    private insert(node: Element): void {
        document.body.appendChild(node)

        this.nodes.add(node)
    }
}
