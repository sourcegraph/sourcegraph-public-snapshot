import classNames from 'classnames'
import React from 'react'

import { WebviewPageProps } from '../platform/context'

import styles from './SearchSidebar.module.scss'

interface OpenSearchPanelCtaProps extends Pick<WebviewPageProps, 'sourcegraphVSCodeExtensionAPI'> {
    onDesktop: boolean
}

export const OpenSearchPanelCta: React.FunctionComponent<OpenSearchPanelCtaProps> = ({ onDesktop }) => (
    <div className={classNames('d-flex flex-column align-items-left justify-content-center')}>
        <p className={classNames(styles.ctaTitle)}>Welcome!</p>
        <div className={classNames(styles.ctaContainer)}>
            <p className={classNames(styles.ctaParagraph)}>
                The Sourcegraph extension allows you to search millions of open source repositories without cloning them
                to your local machine.
            </p>
            <p className={classNames(styles.ctaParagraph)}>
                Developers at some of the world's best software companies use Sourcegraph to onboard to new code bases,
                find examples, research errors, and resolve incidents.
            </p>
            <div className={classNames(styles.ctaParagraph)}>
                <p className={classNames('my-0', styles.text)}>Learn more:</p>
                <p>
                    <a href="http://sourcegraph.com/" className={classNames('my-0', styles.text)}>
                        Sourcegraph.com
                    </a>
                    <br />
                    <a
                        href="https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph"
                        className={classNames('my-0', styles.text)}
                    >
                        Sourcegraph VS Code extension
                    </a>
                </p>
            </div>
        </div>
        {!onDesktop && (
            <div className={classNames(styles.ctaParagraph)}>
                <p className={classNames(styles.ctaWarningText)}>
                    IMPORTANT: Please setup Access Token and CORS to use Sourcegraph Private Instance on VS Code Web.
                </p>
            </div>
        )}
    </div>
)
