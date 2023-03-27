/* eslint-disable react/no-array-index-key */
import React, { useCallback, useState } from 'react'

import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'
import parse from 'html-react-parser'

interface CodeBlocksProps {
    displayText: string
}

export const CodeBlocks: React.FunctionComponent<CodeBlocksProps> = ({ displayText }) => {
    const [copiedText, setCopiedText] = useState('')

    const preBlocks = displayText.match(/<(\w+)[^>]*>(.*?)<\/\1>|<pre[^>]*>[\S\s]*?<\/pre>/g) || []

    const createDivForCopy = useCallback((text: string) => {
        const element = document.createElement('div')
        element.innerHTML = text
        return element.innerText
    }, [])

    const copyTextToClipboard = useCallback(
        (text: string) => {
            const plainText = createDivForCopy(text)
            navigator.clipboard
                .writeText(plainText)
                .then(() => {
                    setCopiedText(text)
                    setTimeout(() => {
                        setCopiedText('')
                    }, 3000)
                })
                .catch(error => {
                    console.error(`Failed to copy text to clipboard: ${error}`)
                })
        },
        [createDivForCopy]
    )

    return (
        <p>
            {preBlocks.map((block, index) => {
                if (block.match(/^<pre/)) {
                    return (
                        <p className="chat-code-block-container" key={index}>
                            {parse(block)}
                            <VSCodeButton
                                title="Copy code"
                                className="chat-code-block-copy-btn"
                                appearance="icon"
                                onClick={() => copyTextToClipboard(block)}
                            >
                                {copiedText === block ? (
                                    <i className="codicon codicon-check" />
                                ) : (
                                    <i className="codicon codicon-copy" />
                                )}
                            </VSCodeButton>
                        </p>
                    )
                }
                return <p key={index}>{parse(block)}</p>
            })}
        </p>
    )
}
