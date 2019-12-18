import * as H from 'history'
import React, { useCallback } from 'react'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { Form } from '../../components/Form'
import { submitSearch, QueryState } from '../helpers'
import { QueryInput } from './QueryInput'
import { SearchButton } from './SearchButton'
import { PatternTypeProps } from '..'

interface Props extends ActivationProps, PatternTypeProps {
    location: H.Location
    history: H.History
    navbarSearchState: QueryState
    onChange: (newValue: QueryState) => void
}

/**
 * The search item in the navbar
 */
export const SearchNavbarItem: React.FunctionComponent<Props> = ({
    navbarSearchState,
    onChange,
    activation,
    location,
    history,
    patternType,
    setPatternType,
}) => {
    // Only autofocus the query input on search result pages (otherwise we
    // capture down-arrow keypresses that the user probably intends to scroll down
    // in the page).
    const autoFocus = location.pathname === '/search'

    const onSubmit = useCallback(
        (e: React.FormEvent<HTMLFormElement>): void => {
            e.preventDefault()
            submitSearch(history, navbarSearchState.query, 'nav', patternType, activation)
        },
        [history, navbarSearchState.query, patternType, activation]
    )

    return (
        <Form className="search search--navbar-item d-flex align-items-start flex-grow-1" onSubmit={onSubmit}>
            <QueryInput
                value={navbarSearchState}
                onChange={onChange}
                autoFocus={autoFocus ? 'cursor-at-end' : undefined}
                hasGlobalQueryBehavior={true}
                location={location}
                history={history}
                patternType={patternType}
                setPatternType={setPatternType}
            />
            <SearchButton />
        </Form>
    )
}
