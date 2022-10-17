import React from 'react'

import { mdiMagnify } from '@mdi/js'
import classNames from 'classnames'

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
            <Icon aria-hidden="true" svgPath={mdiMagnify} />
        </Button>
        {!hideHelpButton && (
            <SearchHelpDropdownButton
                isSourcegraphDotCom={isSourcegraphDotCom}
                className={styles.helpButton}
                telemetryService={telemetryService}
            />
        )}
    </div>
)
