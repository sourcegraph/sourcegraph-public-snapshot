import React, { useMemo } from 'react'

import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'

import { highlightCodeSafe } from '@sourcegraph/common'
import { Button } from '@sourcegraph/wildcard'

import styles from './CodeSnippet.module.scss'

interface CodeSnippetProps {
    /** The code to be displayed. */
    code: string
    /** Hint to the language, used for syntax-highlighting the code-snippet. */
    language: string
    /** Enable inclusion of a "Copy" button in snippet */
    withCopyButton?: boolean

    className?: string
}

export const CodeSnippet: React.FunctionComponent<React.PropsWithChildren<CodeSnippetProps>> = ({
    code,
    language,
    className,
    withCopyButton = false,
}) => {
    const highlightedInput = useMemo(() => ({ __html: highlightCodeSafe(code, language) }), [code, language])
    return (
        <pre className={classNames('bg-code rounded border p-3 position-relative', className)}>
            {withCopyButton && (
                <Button className={styles.copyButton} onClick={() => copy(code)}>
                    <ContentCopyIcon className="pr-2 pt-2" />
                </Button>
            )}
            <code dangerouslySetInnerHTML={highlightedInput} />
        </pre>
    )
}
