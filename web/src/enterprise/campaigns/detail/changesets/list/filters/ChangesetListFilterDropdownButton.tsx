/* eslint-disable react/jsx-no-bind */
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { QueryLink } from '../../../../../../components/listHeaderQueryLinks/ListHeaderQueryLinks'
import { QueryParameterProps } from '../../../../../../util/useQueryParameter'
import { Link } from 'react-router-dom'
import { threadsQueryWithValues, threadsQueryMatches } from '../../../../../../components/listHeaderQueryLinks/url'
import { CheckableDropdownItem } from '../../../../../../components/CheckableDropdownItem'

export interface ChangesetListFilterItem
    extends Pick<QueryLink, 'label' | 'queryField' | 'queryValues' | 'removeQueryFields'> {}

// tslint:disable-next-line: no-any
interface Props extends QueryParameterProps {
    items: ChangesetListFilterItem[]

    buttonText: string
    headerText: string

    className?: string
}

export const ChangesetListFilterDropdownButton: React.FunctionComponent<Props> = ({
    items,
    buttonText,
    headerText,
    query,
    locationWithQuery,
    className = '',
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className} direction="down">
            <DropdownToggle className="btn bg-transparent border-0" caret={true}>
                {buttonText}
            </DropdownToggle>
            <DropdownMenu>
                {headerText && (
                    <>
                        <DropdownItem header={true}>{headerText}</DropdownItem>
                        <DropdownItem divider={true} />
                    </>
                )}
                {items.map(item => (
                    <CheckableDropdownItem
                        key={item.label}
                        checked={item.queryValues.every(queryValue =>
                            threadsQueryMatches(query, { [item.queryField]: queryValue })
                        )}
                        tag={Link}
                        // TODO(sqs): look at ListHeaderQueryLink, handle removeQueryFields
                        to={locationWithQuery(threadsQueryWithValues(query, { [item.queryField]: item.queryValues }))}
                        className="d-flex align-items-center"
                    >
                        <span className="text-truncate">{item.label}</span>
                    </CheckableDropdownItem>
                ))}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
