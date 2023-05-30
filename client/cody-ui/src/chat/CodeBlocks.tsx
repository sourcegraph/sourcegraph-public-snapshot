import React, { useEffect, useMemo } from 'react'

import classNames from 'classnames'

import { renderCodyMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'

import { CopyButtonProps } from '../Chat'

import styles from './CodeBlocks.module.css'

interface CodeBlocksProps {
    displayText: string

    copyButtonClassName?: string
    insertButtonClassName?: string

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
    copyButtonOnSubmit?: CopyButtonProps['copyButtonOnSubmit'],
    insertButtonClassName?: string
): HTMLElement {
    const copyButton = document.createElement('button')
    copyButton.textContent = 'Copy'
    copyButton.className = className
    copyButton.addEventListener('click', () => {
        navigator.clipboard.writeText(text).catch(error => console.error(error))
        copyButton.textContent = 'Copied'
        setTimeout(() => (copyButton.textContent = 'Copy'), 3000)
        if (copyButtonOnSubmit) {
            copyButtonOnSubmit('copyButton')
        }
    })
    // The insert button is for IDE integrations. It allows the user to insert the code into the editor.
    const insertButton = createInsertButton(text, insertButtonClassName, copyButtonOnSubmit)

    // The container will contain the buttons and the <pre> element with the code.
    // This allows us to position the buttons independent of the code.
    const buttons = document.createElement('div')
    buttons.className = styles.buttons
    buttons.append(insertButton, copyButton)

    const container = document.createElement('div')
    container.className = styles.container
    container.append(buttons)

    return container
}

function createInsertButton(
    text: string,
    insertButtonClassName?: string,
    copyButtonOnSubmit?: CopyButtonProps['copyButtonOnSubmit']
): HTMLElement {
    if (!insertButtonClassName || !copyButtonOnSubmit) {
        return document.createElement('span')
    }
    const insertButton = document.createElement('button')
    insertButton.textContent = 'Insert at Cursor'
    insertButton.title = 'Insert this at the current cursor position'
    insertButton.className = classNames(styles.insertButton, insertButtonClassName)
    insertButton.addEventListener('click', () => {
        copyButtonOnSubmit(text, true)
    })
    return insertButton
}

export const CodeBlocks: React.FunctionComponent<CodeBlocksProps> = ({
    displayText,
    copyButtonClassName,
    CopyButtonProps,
    insertButtonClassName,
}) => {
    useEffect(() => {
        const preElements = document.querySelectorAll('pre')
        for (const preElement of preElements) {
            const preText = preElement.textContent
            // check if preElement has button element
            const hasCopyButton = preElement.querySelector(`.${styles.container}`)
            if (!hasCopyButton && preText?.trim()) {
                // We have to wrap the `<pre>` tag in the copy button container, otherwise
                // the Copy button scrolls along with the code.
                wrapElement(
                    preElement,
                    createCopyButtonWithContainer(
                        preText,
                        classNames(styles.copyButton, copyButtonClassName),
                        CopyButtonProps,
                        classNames(styles.insertButton, insertButtonClassName)
                    )
                )
            }
        }
    }, [CopyButtonProps, copyButtonClassName, displayText, insertButtonClassName])

    return useMemo(() => <div dangerouslySetInnerHTML={{ __html: renderCodyMarkdown(displayText) }} />, [displayText])
}
