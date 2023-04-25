import React, { useEffect } from 'react'

import classNames from 'classnames'

import { renderMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'

import { CopyButtonProps } from '../Chat'

import styles from './CodeBlocks.module.css'

interface CodeBlocksProps {
    displayText: string

    copyButtonClassName?: string

    CopyButtonProps?: CopyButtonProps['copyButtonOnSubmit']
}

function wrapElement(element: HTMLElement, wrapperElement: HTMLElement): void {
    if (!element.parentNode) {
        return
    }
    element.parentNode.insertBefore(wrapperElement, element)
    wrapperElement.append(element)
}

function createCopyButtonWithContainer(
    text: string,
    className: string,
    copyButtonOnSubmit?: CopyButtonProps['copyButtonOnSubmit']
): HTMLElement {
    const copyButton = document.createElement('button')
    copyButton.textContent = 'Copy'
    copyButton.className = className
    copyButton.addEventListener('click', () => {
        navigator.clipboard.writeText(text).catch(error => console.error(error))
        copyButton.textContent = 'Copied!'
        setTimeout(() => (copyButton.textContent = 'Copy'), 3000)
        if (copyButtonOnSubmit) {
            copyButtonOnSubmit('copyButton')
        }
    })

    // The container will contain the copy button and the <pre> element with the code.
    // This allows us to position the copy button independent of the code.
    const container = document.createElement('div')
    container.className = styles.container
    container.append(copyButton)
    return container
}

export const CodeBlocks: React.FunctionComponent<CodeBlocksProps> = ({
    displayText,
    copyButtonClassName,
    CopyButtonProps,
}) => {
    useEffect(() => {
        const preElements = document.querySelectorAll('pre')
        for (const preElement of preElements) {
            const preText = preElement.textContent
            const hasCopyButton = preElement.querySelector(`.${styles.container}`)
            if (!hasCopyButton && preText && preText.trim().length > 0) {
                // We have to wrap the `<pre>` tag in the copy button container, otherwise
                // the Copy button scrolls along with the code.
                wrapElement(
                    preElement,
                    createCopyButtonWithContainer(
                        preText,
                        classNames(styles.copyButton, copyButtonClassName),
                        CopyButtonProps
                    )
                )
            }
        }
    }, [copyButtonClassName, displayText, CopyButtonProps])

    return <div dangerouslySetInnerHTML={{ __html: renderMarkdown(displayText) }} />
}
