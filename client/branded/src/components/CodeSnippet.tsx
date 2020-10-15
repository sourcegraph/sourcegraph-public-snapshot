import classNames from 'classnames'
import React, { useMemo } from 'react'
import { highlightCodeSafe } from '../../../shared/src/util/markdown'

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
        <div className={classNames('code-snippet__container rounded p-3', className)}>
            <pre className="m-0" dangerouslySetInnerHTML={highlightedInput} />
        </div>
    )
}
