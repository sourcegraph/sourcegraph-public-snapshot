import { LocationDescriptor } from 'history'
import React, { PropsWithChildren } from 'react'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import { pluralize } from '../../../../shared/src/util/strings'

export interface SummaryCountItemDescriptor<C> {
    noun: string
    pluralNoun?: string
    icon: React.ComponentType<{ className?: string }>
    count: number | ((context: C) => number)
    url?: LocationDescriptor | ((context: C) => LocationDescriptor | undefined)
    after?: (context: C) => JSX.Element
    condition?: (context: C) => boolean
}

interface Props<C> {
    itemDescriptors: SummaryCountItemDescriptor<C>[]
    context: C
    vertical?: boolean

    className?: string
}

/**
 * A horizontal bar with item counts.
 */
export const SummaryCountBar = <C extends {}>({
    itemDescriptors,
    context,
    vertical,
    className = '',
}: PropsWithChildren<Props<C>>): React.ReactElement => (
    <nav className={`summary-count-bar border ${className}`}>
        <ul className={`nav w-100 ${vertical ? 'flex-column align-items-start' : ''}`}>
            {itemDescriptors
                .filter(({ condition }) => !condition || condition(context))
                .map(({ icon: Icon, ...item }, i) => {
                    const count = typeof item.count === 'function' ? item.count(context) : item.count
                    return (
                        <li key={i} className="nav-item flex-1 text-center">
                            <LinkOrSpan
                                to={typeof item.url === 'function' ? item.url(context) : item.url}
                                className="nav-link text-nowrap"
                            >
                                <Icon className="icon-inline text-muted" /> <strong>{count}</strong>{' '}
                                <span className="text-muted">{pluralize(item.noun, count || 0, item.pluralNoun)}</span>
                                {item.after && item.after(context)}
                            </LinkOrSpan>
                        </li>
                    )
                })}
        </ul>
    </nav>
)
