import H from 'history'
import React from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { threadsQueryMatches, threadsQueryWithValues } from '../url'

/** A link in {@link ListHeaderQueryLinks}. */
export interface QueryLink {
    label: string
    queryField: string
    queryValues: string[]
    removeQueryFields?: string[]
    count?: number
    icon?: React.ComponentType<{ className?: string }>
}

export const ListHeaderQueryLink: React.FunctionComponent<
    QueryLink &
        Pick<QueryParameterProps, 'query'> & {
            className?: string
            activeClassName?: string
            inactiveClassName?: string
            location: H.Location
        }
> = ({
    label,
    queryField,
    queryValues,
    removeQueryFields,
    count,
    icon: Icon,
    query,
    className = '',
    activeClassName = 'active',
    inactiveClassName = 'text-muted',
    location,
}) => {
    const queryFieldValues: { [name: string]: null | string[] } = { [queryField]: queryValues }
    if (removeQueryFields) {
        for (const f of removeQueryFields) {
            if (f !== queryField) {
                queryFieldValues[f] = null
            }
        }
    }
    return (
        <Link
            to={urlForThreadsQuery(location, threadsQueryWithValues(query, queryFieldValues))}
            className={`${className} ${
                queryValues.every(queryValue => threadsQueryMatches(query, { [queryField]: queryValue }))
                    ? activeClassName
                    : inactiveClassName
            }`}
        >
            {Icon && <Icon className="icon-inline mr-1" />}
            {count} {label}
        </Link>
    )
}

interface Props extends Pick<QueryParameterProps, 'query'> {
    /** The links to display. */
    links: QueryLink[]

    className?: string
    itemClassName?: string
    itemActiveClassName?: string
    itemInactiveClassName?: string

    location: H.Location
}

/**
 * A button group with links that interact with the query that determines the contents of a list,
 * such as showing "4 open" and "3 closed".
 */
export const ListHeaderQueryLinksButtonGroup: React.FunctionComponent<Props> = ({
    query,
    links,
    location,
    className = '',
    itemClassName = '',
    itemActiveClassName = '',
    itemInactiveClassName = '',
}) => (
    <div className={`btn-group ${className}`}>
        {links.map((linkProps, i) => (
            <ListHeaderQueryLink
                key={i}
                {...linkProps}
                query={query}
                className={`btn ${itemClassName}`}
                activeClassName={itemActiveClassName}
                inactiveClassName={itemInactiveClassName}
                location={location}
            />
        ))}
    </div>
)

/**
 * A nav with links that interact with the query that determines the contents of a list, such as
 * showing "4 open" and "3 closed".
 */
export const ListHeaderQueryLinksNav: React.FunctionComponent<Props> = ({
    query,
    links,
    location,
    className = '',
    itemClassName = '',
    itemActiveClassName = '',
    itemInactiveClassName = '',
}) => (
    <ul className={`nav ${className}`}>
        {links.map((linkProps, i) => (
            <li key={i} className="nav-item">
                <ListHeaderQueryLink
                    {...linkProps}
                    query={query}
                    className={`nav-link p-1 ${itemClassName}`}
                    activeClassName={itemActiveClassName}
                    inactiveClassName={itemInactiveClassName}
                    location={location}
                />
            </li>
        ))}
    </ul>
)

function urlForThreadsQuery(location: Pick<H.Location, 'search'>, threadsQuery: string): H.LocationDescriptor {
    const params = new URLSearchParams(location.search)
    params.set('q', threadsQuery)
    return { ...location, search: `${params}` }
}
