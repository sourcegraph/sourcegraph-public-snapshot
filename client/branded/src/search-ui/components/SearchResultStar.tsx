import classNames from 'classnames'
import type { MdiReactIconProps } from 'mdi-react'
import StarIcon from 'mdi-react/StarIcon'

import styles from './SearchResultStar.module.scss'

export const SearchResultStar: React.FunctionComponent<MdiReactIconProps> = ({ className, ...props }) => (
    <StarIcon className={classNames(styles.star, className)} {...props} />
)
