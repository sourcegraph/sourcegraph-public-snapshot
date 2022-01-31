import classNames from 'classnames'
import React from 'react'

import { WebviewPageProps } from '../../platform/context'
import styles from '../index.module.scss'

export const BrandHeader: React.FunctionComponent<Pick<WebviewPageProps, 'theme' | 'extensionCoreAPI'>> = ({
    theme,
    extensionCoreAPI,
}) => (
    <>
        <div className={classNames('d-flex justify-content-end w-100 p-3')}>
            <button
                type="button"
                className={classNames('btn btn-primary text border-0 text-decoration-none px-3', styles.feedbackButton)}
                onClick={() =>
                    extensionCoreAPI
                        .openLink('https://github.com/sourcegraph/sourcegraph/discussions/categories/feedback')
                        .catch(error => {
                            console.error('Error opening feedback link', error)
                        })
                }
            >
                Give us Feedback
            </button>
        </div>
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
