import classNames from 'classnames'
import React from 'react'

import { WebviewPageProps } from '../platform/context'

import styles from './SearchSidebar.module.scss'

interface OpenSearchPanelCtaProps extends Pick<WebviewPageProps, 'sourcegraphVSCodeExtensionAPI'> {
    className?: string
    onDesktop: boolean
}

export const OpenSearchPanelCta: React.FunctionComponent<OpenSearchPanelCtaProps> = ({
    className,
    sourcegraphVSCodeExtensionAPI,
    onDesktop,
}) => (
    <div className={classNames('d-flex flex-column align-items-left justify-content-center', className)}>
        <p className={classNames('mt-3', styles.title)}>Welcome!</p>
        <p className={classNames('my-3', styles.text)}>
            The Sourcegraph extension allows you to search millions of open source repositories without cloning them to
            your local machine.
        </p>
        <p className={classNames('my-3', styles.text)}>
            Developers at some of the world's best software companies use Sourcegraph to onboard to new code bases, find
            examples, research errors, and resolve incidents.
        </p>
        <div className={classNames('my-3', styles.text)}>
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
        <button
            type="button"
            onClick={() => sourcegraphVSCodeExtensionAPI.openSearchPanel()}
            className={classNames('mb-3 btn btn-sm w-100 border-0 font-weight-normal', styles.button)}
        >
            Open Search Panel
        </button>
        {/* Display warning if user is using VS Code Web */}
        {!onDesktop && <p>IMPORTANT: Please add Access Token and CORS to use Sourcegraph on VS Code Web.</p>}
    </div>
)
