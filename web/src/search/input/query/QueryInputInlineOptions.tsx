import AlphabeticalIcon from 'mdi-react/AlphabeticalIcon'
import FormatLetterCaseIcon from 'mdi-react/FormatLetterCaseIcon'
import RegexIcon from 'mdi-react/RegexIcon'
import React, { useState } from 'react'

const buttonClass = (value: boolean) => `query-input-inline-options__option-${value ? 'on' : 'off'}`

export interface InlineQueryOptions {
    matchCaseSensitive: boolean
    matchWholeWord: boolean
    useRegexp: boolean
}

// tslint:disable: jsx-no-lambda
export const QueryInputInlineOptions: React.FunctionComponent<{ className?: string }> = ({ className = '' }) => {
    const [matchCaseSensitive, setMatchCaseSensitive] = useState(false)
    const [matchWholeWord, setMatchWholeWord] = useState(false)
    const [useRegexp, setUseRegexp] = useState(false)

    return (
        <div className={`query-input-inline-options ${className}`}>
            <button
                className={`btn btn-secondary border-0 btn-sm bg-transparent ${buttonClass(matchCaseSensitive)}`}
                type="button"
                onClick={() => setMatchCaseSensitive(!matchCaseSensitive)}
                title="Match case sensitive"
            >
                <FormatLetterCaseIcon className="icon-inline" />
            </button>
            <button
                className={`btn btn-secondary border-0 btn-sm bg-transparent ${buttonClass(matchWholeWord)}`}
                type="button"
                onClick={() => setMatchWholeWord(!matchWholeWord)}
                title="Match whole word"
            >
                <AlphabeticalIcon className="icon-inline" />
            </button>
            <button
                className={`btn btn-secondary border-0 btn-sm bg-transparent ${buttonClass(useRegexp)}`}
                type="button"
                onClick={() => setUseRegexp(!useRegexp)}
                title="Use regular expression"
            >
                <RegexIcon className="icon-inline" />
            </button>
        </div>
    )
}
