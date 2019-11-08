import * as H from 'history'
import React, { useCallback } from 'react'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { Form } from '../../components/Form'
import { submitSearch, QueryValue } from '../helpers'
import { QueryInput } from './QueryInput'
import { SearchButton } from './SearchButton'
import { PatternTypeProps } from '..'

interface Props extends ActivationProps, PatternTypeProps {
    location: H.Location
    history: H.History
    navbarSearchValue: QueryValue
    onChange: (newValue: QueryValue) => void
}

/**
 * The search item in the navbar
 */
export const SearchNavbarItem: React.FunctionComponent<Props> = ({
    navbarSearchValue,
    onChange,
    activation,
    location,
    history,
    patternType,
    togglePatternType,
}) => {
    // Only autofocus the query input on search result pages (otherwise we
    // capture down-arrow keypresses that the user probably intends to scroll down
    // in the page).
    const autoFocus = location.pathname === '/search'

    const onSubmit = useCallback(
        (e: React.FormEvent<HTMLFormElement>): void => {
            e.preventDefault()
            submitSearch(history, navbarSearchValue.query, 'nav', patternType, activation)
        },
        [history, navbarSearchValue.query, patternType, activation]
    )

    return (
        <Form className="search search--navbar-item d-flex align-items-start" onSubmit={onSubmit}>
            <QueryInput
                value={navbarSearchValue}
                onChange={onChange}
                autoFocus={autoFocus ? 'cursor-at-end' : undefined}
                hasGlobalQueryBehavior={true}
                location={location}
                history={history}
                patternType={patternType}
                togglePatternType={togglePatternType}
            />
            <SearchButton />
        </Form>
    )
}
