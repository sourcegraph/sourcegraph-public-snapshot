import React from 'react'

import classNames from 'classnames'

import { WebviewPageProps } from '../../platform/context'

import styles from '../index.module.scss'

export const BrandHeader: React.FunctionComponent<Pick<WebviewPageProps, 'theme'>> = ({ theme }) => (
    <>
        <img
            className={classNames(styles.logo)}
            src={`https://sourcegraph.com/.assets/img/sourcegraph-logo-${
                theme === 'theme-light' ? 'light' : 'dark'
            }.svg`}
            alt="Sourcegraph logo"
        />
        <div className={classNames(styles.logoText)}>Search your code and 2M+ open source repositories</div>
    </>
)
