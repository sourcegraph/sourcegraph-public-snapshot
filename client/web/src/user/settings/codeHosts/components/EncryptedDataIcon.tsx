import { mdiShieldCheck } from '@mdi/js'
import classNames from 'classnames'

import { Icon, AccessibleIcon } from '@sourcegraph/wildcard'

import styles from './EncryptedDataIcon.module.scss'

export const EncryptedDataIcon: AccessibleIcon = ({ className, ...rest }) => (
    <Icon
        svgPath={mdiShieldCheck}
        className={classNames('text-muted', styles.iconInside, className)}
        aria-hidden="true"
        {...rest}
    />
)
