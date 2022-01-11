import classNames from 'classnames'
import SearchIcon from 'mdi-react/SearchIcon'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import styles from './SearchButton.module.scss'
import { SearchHelpDropdownButton } from './SearchHelpDropdownButton'

interface Props {
    /** Hide the "help" icon and dropdown. */
    hideHelpButton?: boolean
    className?: string
}

/**
 * A search button with a dropdown with related links. It must be wrapped in a form whose onSubmit
 * handler performs the search.
 */
export const SearchButton: React.FunctionComponent<Props> = ({ hideHelpButton, className }) => (
    <div className={className}>
        <Button
            data-search-button={true}
            className={classNames('test-search-button', styles.btn)}
            type="submit"
            aria-label="Search"
            variant="primary"
        >
            <SearchIcon className="icon-inline" aria-hidden="true" />
        </Button>
        {!hideHelpButton && <SearchHelpDropdownButton />}
    </div>
)
