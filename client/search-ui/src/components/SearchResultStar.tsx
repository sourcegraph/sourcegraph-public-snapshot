import { mdiStar } from '@mdi/js'
import classNames from 'classnames'

import { Icon, IconProps } from '@sourcegraph/wildcard'

import styles from './SearchResultStar.module.scss'

export const SearchResultStar: React.FunctionComponent<React.PropsWithoutRef<IconProps>> = ({
    className,
    ...props
}) => (
    <Icon
        className={classNames(styles.star, className)}
        {...props}
        svgPath={mdiStar}
        inline={false}
        aria-hidden={true}
    />
)
