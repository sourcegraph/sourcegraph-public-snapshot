import * as React from 'react'

import classNames from 'classnames'

import { SourcegraphIcon } from '@sourcegraph/wildcard'

import styles from './HoverOverlayLogo.module.scss'

export const HoverOverlayLogo: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <span className={classNames(styles.container, className)}>
        <SourcegraphIcon width="1rem" height="1rem" />
    </span>
)
