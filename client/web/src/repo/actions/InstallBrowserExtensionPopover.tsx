import React from 'react'

import classNames from 'classnames'
import ExportIcon from 'mdi-react/ExportIcon'
import PlusThickIcon from 'mdi-react/PlusThickIcon'
import FocusLock from 'react-focus-lock'

import { ExternalServiceKind } from '@sourcegraph/shared/src/schema'
import { ButtonLink, PopoverContent, Position, Typography } from '@sourcegraph/wildcard'

import { SourcegraphIcon } from '../../auth/icons'

import { serviceKindDisplayNameAndIcon } from './GoToCodeHostAction'

import styles from './InstallBrowserExtensionPopover.module.scss'

interface Props {
    url: string
    serviceKind: ExternalServiceKind | null
    onClose: () => void
    onReject: () => void
    onInstall: () => void
}

export const InstallBrowserExtensionPopover: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    url,
    serviceKind,
    onClose,
    onReject,
    onInstall,
}) => {
    const { displayName, icon } = serviceKindDisplayNameAndIcon(serviceKind)
    const Icon = icon || ExportIcon

    // Open all external links in new tab
    const linkProps = { rel: 'noopener noreferrer', target: '_blank' }

    return (
        <PopoverContent
            position={Position.bottom}
            className={classNames('shadow border', styles.installBrowserExtensionPopover)}
            tail={true}
        >
            <FocusLock returnFocus={true}>
                <div className="p-3 text-wrap  test-install-browser-extension-popover">
                    <Typography.H3 className="mb-0 test-install-browser-extension-popover-header">
                        Take Sourcegraph's code intelligence to {displayName}!
                    </Typography.H3>
                    <p className="py-3">
                        Install Sourcegraph browser extension to add code intelligence{' '}
                        {serviceKind === ExternalServiceKind.PHABRICATOR
                            ? 'while browsing and reviewing code'
                            : `to ${serviceKind === ExternalServiceKind.GITLAB ? 'MR' : 'PR'}s and file views`}{' '}
                        on {displayName} or any other connected code host.
                    </p>

                    <div
                        className={classNames(
                            'mx-auto d-flex justify-content-between align-items-center',
                            styles.graphicContainer
                        )}
                    >
                        <SourcegraphIcon className={classNames('p-1', styles.logo)} />
                        <PlusThickIcon className={styles.plusIcon} />
                        <Icon role="img" className={styles.logo} aria-hidden={true} />
                    </div>

                    <div className="d-flex justify-content-end">
                        <ButtonLink
                            className="mr-2"
                            onSelect={onReject}
                            to={url}
                            {...linkProps}
                            variant="secondary"
                            outline={true}
                        >
                            No, thanks
                        </ButtonLink>

                        <ButtonLink
                            className="mr-2"
                            onSelect={onClose}
                            to={url}
                            {...linkProps}
                            variant="secondary"
                            outline={true}
                        >
                            Remind me later
                        </ButtonLink>

                        <ButtonLink
                            className="mr-2"
                            onSelect={onInstall}
                            to="/help/integration/browser_extension"
                            {...linkProps}
                            variant="primary"
                        >
                            Install browser extension
                        </ButtonLink>
                    </div>
                </div>
            </FocusLock>
        </PopoverContent>
    )
}
