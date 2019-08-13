import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { DropdownMenuFilter } from '../../../../components/dropdownMenuFilter/DropdownMenuFilter'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'

const LOADING = 'loading' as const

export interface ThreadListFilterContext extends QueryParameterProps {
    threadConnection:
        | typeof LOADING
        | Pick<GQL.IThreadConnection | GQL.IThreadOrThreadPreviewConnection, 'filters'>
        | ErrorLike

    className?: string
}

export interface ThreadListFilterItem {
    beforeText?: React.ReactFragment
    text: string
    count: number | null
    queryPart: string
    isApplied: boolean
}

interface Props<
    P extends keyof Pick<GQL.IThreadConnectionFilters, Exclude<keyof GQL.IThreadConnectionFilters, '__typename'>>
> extends ThreadListFilterContext, QueryParameterProps {
    filterKey: P
    itemFunc: (filter: GQL.IThreadConnectionFilters[P][number]) => ThreadListFilterItem

    buttonText: string
    noun: string
    pluralNoun?: string

    className?: string
}

export const ThreadListFilterDropdownButton = <
    P extends keyof Pick<GQL.IThreadConnectionFilters, Exclude<keyof GQL.IThreadConnectionFilters, '__typename'>>
>({
    threadConnection,
    filterKey,
    itemFunc,
    buttonText,
    noun,
    pluralNoun = pluralize(noun, 2),
    query: parentQuery,
    onQueryChange: parentOnQueryChange,
    className = '',
}: Props<P>): React.ReactElement => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onSelect = useCallback(
        (item: ThreadListFilterItem) => {
            parentOnQueryChange(`${parentQuery}${parentQuery ? ' ' : ''}${item.queryPart}`)
        },
        [parentOnQueryChange, parentQuery]
    )

    const [query, setQuery] = useState('')
    const isEmpty =
        threadConnection !== LOADING &&
        !isErrorLike(threadConnection) &&
        threadConnection.filters[filterKey].length === 0
    const itemsFiltered: typeof LOADING | ThreadListFilterItem[] | ErrorLike =
        threadConnection !== LOADING && !isErrorLike(threadConnection)
            ? // TODO!(sqs): this type error is erroneous
              threadConnection.filters[filterKey]
                  .map(itemFunc)
                  .filter(({ text }) => text.toLowerCase().includes(query.toLowerCase()))
            : threadConnection

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
                {itemsFiltered === LOADING ? (
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
                    itemsFiltered.slice(0, 20 /* TODO!(sqs) hack */).map(item => (
                        <DropdownItem
                            key={item.queryPart}
                            // tslint:disable-next-line: jsx-no-lambda
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
