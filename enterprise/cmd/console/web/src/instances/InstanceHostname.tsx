import classNames from 'classnames'
import React from 'react'
import styles from './InstanceHostname.module.scss'

export const InstanceHostname: React.FunctionComponent<{ url: string }> = ({ url }) => (
    <>
        <span>{url.replace('https://', '').replace('.sourcegraph.com', '')}</span>
        <span className={classNames(styles.dotSourcegraphDotCom, 'font-weight-normal')}>.sourcegraph.com</span>
    </>
)
