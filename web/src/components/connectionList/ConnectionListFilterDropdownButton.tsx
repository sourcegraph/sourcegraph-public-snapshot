/* eslint-disable react/jsx-no-bind */
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { pluralize } from '../../../../shared/src/util/strings'
import { DropdownMenuFilter } from '../dropdownMenuFilter/DropdownMenuFilter'
import { QueryParameterProps } from '../../util/useQueryParameter'

export interface ConnectionListFilterContext<F extends {}> extends QueryParameterProps {
    connection: undefined | { filters: F } | ErrorLike

    className?: string
}

export interface ConnectionListFilterItem {
    beforeText?: React.ReactFragment
    text: string
    count: number | null
    queryPart: string
    isApplied: boolean
}

// tslint:disable-next-line: no-any
interface Props<F extends { [key in string]: any[] }, P extends keyof Pick<F, Exclude<keyof F, '__typename'>>>
    extends ConnectionListFilterContext<F>,
        QueryParameterProps {
    filterKey: P
    itemFunc: (filter: F[P][number]) => ConnectionListFilterItem

    buttonText: string
    noun: string
    pluralNoun?: string

    className?: string
}

export const ConnectionListFilterDropdownButton = <
    // tslint:disable-next-line: no-any
    F extends { [property in P]: any[] },
    P extends keyof Pick<F, Exclude<keyof F, '__typename'>>
>({
    connection,
    filterKey,
    itemFunc,
    buttonText,
    noun,
    pluralNoun = pluralize(noun, 2),
    query: parentQuery,
    onQueryChange: parentOnQueryChange,
    className = '',
}: Props<F, P>): React.ReactElement => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onSelect = useCallback(
        (item: ConnectionListFilterItem) => {
            parentOnQueryChange(`${parentQuery}${parentQuery ? ' ' : ''}${item.queryPart}`)
        },
        [parentOnQueryChange, parentQuery]
    )

    const [query, setQuery] = useState('')
    const isEmpty = connection !== undefined && !isErrorLike(connection) && connection.filters[filterKey].length === 0
    const itemsFiltered: undefined | ConnectionListFilterItem[] | ErrorLike =
        connection !== undefined && !isErrorLike(connection)
            ? // TODO!(sqs): this type error is erroneous
              connection.filters[filterKey]
                  .map(itemFunc)
                  .filter(({ text }) => text.toLowerCase().includes(query.toLowerCase()))
            : connection

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className} direction="down">
            <DropdownToggle className="btn bg-transparent border-0" caret={true}>
                {buttonText}
            </DropdownToggle>
            <DropdownMenu>
                <DropdownMenuFilter
                    value={query}
                    onChange={setQuery}
                    placeholder={`Filter ${pluralNoun}`}
                    header={`Filter by ${noun}`}
                />
                {itemsFiltered === undefined ? (
                    <DropdownItem header={true} className="py-1">
                        Loading {pluralNoun}...
                    </DropdownItem>
                ) : isErrorLike(itemsFiltered) ? (
                    <DropdownItem header={true} className="py-1">
                        Error loading {pluralNoun}
                    </DropdownItem>
                ) : itemsFiltered.length === 0 ? (
                    <DropdownItem header={true}>
                        No {pluralNoun} {!isEmpty && 'match'}
                    </DropdownItem>
                ) : (
                    itemsFiltered.slice(0, 15 /* TODO!(sqs) hack */).map(item => (
                        <DropdownItem
                            key={item.queryPart}
                            onClick={() => onSelect(item)}
                            className="d-flex align-items-center"
                        >
                            {item.beforeText}
                            <span className="text-truncate">{item.text}</span>
                            <span className="flex-1" />
                            {typeof item.count === 'number' && (
                                <span className="ml-3 badge badge-secondary">{item.count}</span>
                            )}
                        </DropdownItem>
                    ))
                )}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
