import { Unsubscribable } from 'rxjs'
import { TextDocumentDecoration } from '../../../../../shared/src/api/protocol/plainTypes'

const combineUnsubscribables = (...unsubscribables: Unsubscribable[]): Unsubscribable => ({
    unsubscribe: () => {
        for (const unsubscribable of unsubscribables) {
            unsubscribable.unsubscribe()
        }
    },
})

// This applies a decoration to a GitHub blob page. This doesn't work with any other code host yet.
export const applyDecoration = ({
    fileElement,
    decoration,
}: {
    fileElement: HTMLElement
    decoration: TextDocumentDecoration
}): Unsubscribable => {
    const unsubscribables: Unsubscribable[] = []
    const ghLineNumber = decoration.range.start.line + 1
    const lineNumberElements: NodeListOf<HTMLElement> = fileElement.querySelectorAll(
        `td[data-line-number="${ghLineNumber}"]`
    )
    if (!lineNumberElements) {
        throw new Error(`Line number ${ghLineNumber} not found`)
    }
    if (lineNumberElements.length !== 1) {
        throw new Error(`Line number ${ghLineNumber} matched ${lineNumberElements.length} elements (expected 1)`)
    }
    const lineNumberElement = lineNumberElements[0]
    if (!lineNumberElement) {
        throw new Error(`Line number ${ghLineNumber} is falsy: ${lineNumberElement}`)
    }
    const lineElement = lineNumberElement.nextElementSibling as HTMLElement | undefined
    if (!lineElement) {
        throw new Error(`Line ${ghLineNumber} is falsy: ${lineNumberElement}`)
    }
    if (decoration.backgroundColor) {
        lineElement.style.backgroundColor = decoration.backgroundColor
        unsubscribables.push({
            unsubscribe: () => {
                lineElement.style.backgroundColor = null
            },
        })
    }
    if (decoration.after) {
        const linkTo = (url: string) => (e: HTMLElement): HTMLElement => {
            const link = document.createElement('a')
            link.setAttribute('href', url)
            link.style.color = decoration.after!.color || null
            link.appendChild(e)
            return link
        }
        const after = document.createElement('span')
        after.style.backgroundColor = decoration.after.backgroundColor || null
        after.textContent = decoration.after.contentText || null
        const annotation = decoration.after.linkURL ? linkTo(decoration.after.linkURL)(after) : after
        lineElement.appendChild(annotation)
        unsubscribables.push({
            unsubscribe: () => {
                annotation.remove()
            },
        })
    }
    return combineUnsubscribables(...unsubscribables)
}
