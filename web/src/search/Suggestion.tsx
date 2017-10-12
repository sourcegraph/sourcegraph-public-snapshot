
import escapeRegexp from 'escape-string-regexp'
import * as React from 'react'

export interface SuggestionProps {

    icon: React.ComponentType<{ className: string }>

    label: string

    /** The query to highlight */
    query: string

    isSelected?: boolean

    /** Called when the user clicks on the suggestion */
    onClick?: () => void

    /** Get a reference to the HTML element for scroll management */
    liRef?: (ref: HTMLLIElement | null) => void
}

export const Suggestion = (props: SuggestionProps) => {
    const lowerQuery = props.query.toLowerCase()
    const splitRegexp = new RegExp(`(${escapeRegexp(lowerQuery)})`, 'gi')
    const parts = props.label.split(splitRegexp)
    const Icon = props.icon
    return (
        <li className={'suggestion' + (props.isSelected ? ' suggestion--selected' : '')} onClick={props.onClick} ref={props.liRef}>
            <Icon className='icon-inline' />
            <div className='suggestion__label'>
                {
                    parts.map((part, i) =>
                        <span key={i} className={part.toLowerCase() === lowerQuery ? 'suggestion__highlighted-query' : ''}>
                            {part}
                        </span>,
                    )
                }
            </div>
            <div className='suggestion__tip' hidden={!props.isSelected}><kbd>enter</kbd> to add as filter</div>
        </li>
    )
}
