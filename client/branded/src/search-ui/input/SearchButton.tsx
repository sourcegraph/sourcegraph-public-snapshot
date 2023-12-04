import type { FC } from 'react'

import { mdiMagnify } from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon } from '@sourcegraph/wildcard'

import styles from './SearchButton.module.scss'

interface Props {
    className?: string
}

/**
 * A search button with a dropdown with related links. It must be wrapped in a form whose onSubmit
 * handler performs the search.
 */
export const SearchButton: FC<Props> = ({ className }) => (
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
    </div>
)
