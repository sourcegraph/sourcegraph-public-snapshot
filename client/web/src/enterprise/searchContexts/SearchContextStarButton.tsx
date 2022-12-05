import React from 'react'

import { mdiStar, mdiStarOutline } from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon } from '@sourcegraph/wildcard'

import styles from './SearchContextStarButton.module.scss'

interface SearchContextStarButtonProps {
    starred: boolean
    onClick: () => void
}

export const SearchContextStarButton: React.FunctionComponent<SearchContextStarButtonProps> = ({
    starred,
    onClick,
}) => (
    <Button variant="icon" onClick={onClick} className={classNames(styles.button, starred && styles.buttonStarred)}>
        <Icon
            svgPath={starred ? mdiStar : mdiStarOutline}
            aria-label={starred ? 'Starred, click to remove star' : 'Not starred, click to add star'}
        />
    </Button>
)
