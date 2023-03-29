/* eslint-disable react/no-array-index-key */
import React, { useEffect } from 'react'

import { renderMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'

interface CodeBlocksProps {
    displayText: string
}

export const CodeBlocks: React.FunctionComponent<CodeBlocksProps> = ({ displayText }) => {
    useEffect(() => {
        const preEls = document.querySelectorAll('pre')
        if (preEls) {
            // Add buttons to copy inner text of each pre elemenet item in preEls
            for (const pre of preEls) {
                const preText = pre.textContent
                if (!pre.querySelector('.chat-code-block-copy-btn') && preText) {
                    const copyBtn = document.createElement('button')
                    copyBtn.textContent = 'Copy'
                    copyBtn.className = 'chat-code-block-copy-btn'
                    copyBtn.addEventListener('click', () => {
                        navigator.clipboard.writeText(preText).catch(error => console.error(error))
                        copyBtn.textContent = 'Copied!'
                        setTimeout(() => {
                            copyBtn.textContent = 'Copy snippet'
                        }, 3000)
                    })
                    pre.append(copyBtn)
                }
            }
        }
    }, [displayText])

    return <p className="chat-code-block-container" dangerouslySetInnerHTML={{ __html: renderMarkdown(displayText) }} />
}
