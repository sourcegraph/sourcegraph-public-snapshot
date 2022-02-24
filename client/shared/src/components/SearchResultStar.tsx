import StarIcon from 'mdi-react/StarIcon'
import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

import styles from './SearchResultStar.module.scss'

export const SearchResultStar: React.FunctionComponent = () => (
    <Icon inline={false} as={StarIcon} className={styles.star} />
)
