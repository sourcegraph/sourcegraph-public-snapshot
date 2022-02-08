import classNames from 'classnames'
import React, { useMemo } from 'react'

import { highlightCodeSafe } from '@sourcegraph/common'

interface CodeSnippetProps {
    /** The code to be displayed. */
    code: string
    /** Hint to the language, used for syntax-highlighting the code-snippet. */
    language: string

    className?: string
}

export const CodeSnippet: React.FunctionComponent<CodeSnippetProps> = ({ code, language, className }) => {
    const highlightedInput = useMemo(() => ({ __html: highlightCodeSafe(code, language) }), [code, language])
    return (
        <pre className={classNames('bg-code rounded border p-3', className)}>
            <code dangerouslySetInnerHTML={highlightedInput} />
        </pre>
    )
}
