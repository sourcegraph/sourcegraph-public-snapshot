import React, { useMemo } from 'react'
import { highlightCodeSafe } from '../../../shared/src/util/markdown'

interface CodeSnippetProps {
    code: string
    language: string
}

export const CodeSnippet: React.FunctionComponent<CodeSnippetProps> = ({ code, language }) => {
    const highlightedInput = useMemo(() => ({ __html: highlightCodeSafe(code, language) }), [code, language])
    return (
        <div className="code-snippet__container rounded p-2">
            <pre className="m-0" dangerouslySetInnerHTML={highlightedInput} />
        </div>
    )
}
