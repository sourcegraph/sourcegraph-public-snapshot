import React from 'react'

import StarIcon from 'mdi-react/StarIcon'

import styles from './SearchResultStar.module.scss'

export const SearchResultStar: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <StarIcon className={styles.star} />
)
