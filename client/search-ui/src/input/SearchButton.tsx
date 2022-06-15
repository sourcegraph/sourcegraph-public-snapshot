import React from 'react'

import classNames from 'classnames'
import SearchIcon from 'mdi-react/SearchIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon } from '@sourcegraph/wildcard'

import { SearchHelpDropdownButton } from './SearchHelpDropdownButton'

import styles from './SearchButton.module.scss'

interface Props extends TelemetryProps {
    /** Hide the "help" icon and dropdown. */
    hideHelpButton?: boolean
    className?: string
    isSourcegraphDotCom?: boolean
}

/**
 * A search button with a dropdown with related links. It must be wrapped in a form whose onSubmit
 * handler performs the search.
 */
export const SearchButton: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    hideHelpButton,
    className,
    isSourcegraphDotCom,
    telemetryService,
}) => (
    <div className={className}>
        <Button
            data-search-button={true}
            className={classNames('test-search-button', styles.btn)}
            type="submit"
            aria-label="Search"
            variant="primary"
        >
            <Icon aria-hidden="true" as={SearchIcon} />
        </Button>
        {!hideHelpButton && (
            <SearchHelpDropdownButton isSourcegraphDotCom={isSourcegraphDotCom} telemetryService={telemetryService} />
        )}
    </div>
)
