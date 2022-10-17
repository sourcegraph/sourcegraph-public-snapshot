import * as React from 'react'

import classNames from 'classnames'

import styles from './FooterPanel.module.scss'

interface FooterPanelProps {
    className?: string
}

export const FooterPanel: React.FunctionComponent<React.PropsWithChildren<FooterPanelProps>> = ({
    children,
    className,
}) => <div className={classNames(className, styles.footer)}>{children}</div>
