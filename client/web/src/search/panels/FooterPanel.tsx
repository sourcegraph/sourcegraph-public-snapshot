import classNames from 'classnames'
import * as React from 'react'

import styles from './FooterPanel.module.scss'

interface FooterPanelProps {
    className?: string
}

export const FooterPanel: React.FunctionComponent<FooterPanelProps> = ({ children, className }) => (
    <div className={classNames(className, styles.footer)}>{children}</div>
)
