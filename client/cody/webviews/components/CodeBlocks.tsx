import React, { useEffect } from 'react'

import { renderMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'

interface CodeBlocksProps {
    displayText: string
}

const copyButtonContainerClass = 'chat-code-block-copy-button-container'
const copyButtonClass = 'chat-code-block-copy-button'

function wrapElement(element: HTMLElement, wrapperElement: HTMLElement): void {
    if (!element.parentNode) {
        return
    }
    element.parentNode.insertBefore(wrapperElement, element)
    wrapperElement.append(element)
}

function createCopyButtonWithContainer(text: string): HTMLElement {
    const copyButton = document.createElement('button')
    copyButton.textContent = 'Copy'
    copyButton.className = copyButtonClass
    copyButton.addEventListener('click', () => {
        navigator.clipboard.writeText(text).catch(error => console.error(error))
        copyButton.textContent = 'Copied!'
        setTimeout(() => (copyButton.textContent = 'Copy'), 3000)
    })

    // The container will contain the copy button and the <pre> element with the code.
    // This allows us to position the copy button independent of the code.
    const container = document.createElement('div')
    container.className = copyButtonContainerClass
    container.append(copyButton)
    return container
}

export const CodeBlocks: React.FunctionComponent<CodeBlocksProps> = ({ displayText }) => {
    useEffect(() => {
        const preElements = document.querySelectorAll('pre')
        for (const preElement of preElements) {
            const preText = preElement.textContent
            const hasCopyButton = preElement.querySelector(`.${copyButtonContainerClass}`)
            if (!hasCopyButton && preText && preText.trim().length > 0) {
                // We have to wrap the `<pre>` tag in the copy button container, otherwise
                // the Copy button scrolls along with the code.
                wrapElement(preElement, createCopyButtonWithContainer(preText))
            }
        }
    }, [displayText])

    return <p dangerouslySetInnerHTML={{ __html: renderMarkdown(displayText) }} />
}
