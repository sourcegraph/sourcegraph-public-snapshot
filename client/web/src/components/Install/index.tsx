import { FunctionComponent, useEffect, useState } from 'react'
import classNames from 'classnames'
import copyToClipboard from 'copy-to-clipboard'

import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import { ButtonLink } from '@sourcegraph/wildcard'

import styles from './install.module.scss'

const installText =
    'docker run --publish 7080:7080 --publish 127.0.0.1:3370:3370 --rm --volume ~/.sourcegraph/config:/etc/sourcegraph --volume ~/.sourcegraph/data:/var/opt/sourcegraph sourcegraph/server:3.40.1'

const Install: FunctionComponent = () => {
    const [copied, setCopied] = useState(false)

    const copy = async (): Promise<void> => {
        copyToClipboard(installText)
        setCopied(true)
    }

    useEffect(() => {
        const resetCopied = setTimeout(() => setCopied(false), 1000)
        return () => clearTimeout(resetCopied)
    }, [copied])

    return (
        <div className={classNames(styles.windowUI, 'bg-white overflow-hidden')}>
            <div className={classNames(styles.windowActions, 'd-flex align-items-center px-4')}>
                {['close', 'minimize', 'fullscreen'].map(action => (
                    <span key={action} className={classNames(styles.windowAction, 'bg-white rounded-circle mr-2')} />
                ))}
            </div>

            <div className="p-5">
                <h3 className="font-weight-bold text-3xl">
                    Install Sourcegraph locally
                    <span
                        onClick={copy}
                        onKeyDown={copy}
                        role="button"
                        className={classNames(styles.icon, 'icon-inline ml-4 align-text-top')}
                        tabIndex={0}
                    >
                        <ContentCopyIcon />
                    </span>
                </h3>

                <code className="d-block my-4 pr-5 text-lg">
                    <small className={copied ? classNames(styles.flashBackground, 'text-break') : 'text-break'}>
                        {installText}
                    </small>
                </code>

                <a
                    className={classNames('d-inline-block', styles.deployToServer)}
                    href="https://docs.sourcegraph.com"
                    target="_blank"
                >
                    Deploy to a server or cluster
                </a>
            </div>
        </div>
    )
}

export default Install
