import ExportIcon from 'mdi-react/ExportIcon'
import PlusThickIcon from 'mdi-react/PlusThickIcon'
import React, { useMemo } from 'react'
import { Popover } from 'reactstrap'
import { ButtonLink } from '../../../../shared/src/components/LinkOrButton'
import { SourcegraphIcon } from '../../auth/icons'
import { serviceTypeDisplayNameAndIcon } from './GoToCodeHostAction'
import FocusLock from 'react-focus-lock'

interface CodeHostExtensionPopoverProps {
    url: string
    serviceType: string | null
    onClose: () => void
    onRejection: () => void
    onClickInstall: () => void
    targetID: string
    toggle: () => void
    isOpen: boolean
    wasAuxClick: boolean
}

export const InstallExtensionPopover: React.FunctionComponent<CodeHostExtensionPopoverProps> = ({
    url,
    serviceType,
    onClose,
    onRejection,
    onClickInstall,
    targetID,
    toggle,
    isOpen,
    wasAuxClick,
}) => {
    const { displayName, icon } = serviceTypeDisplayNameAndIcon(serviceType)
    const Icon = icon || ExportIcon

    // Store type of original click (normal vs aux). If it was an aux click, open code host link in new tab
    const linkProps = wasAuxClick ? { rel: 'noopener noreferrer', target: '_blank' } : {}

    return (
        <Popover
            toggle={toggle}
            target={targetID}
            isOpen={isOpen}
            popperClassName="shadow border install-extension-popover web-content"
            innerClassName="border-0"
            placement="bottom"
            boundariesElement="window"
            modifiers={useMemo(
                () => ({
                    offset: {
                        offset: '0, 0',
                        enabled: true,
                    },
                }),
                []
            )}
        >
            {isOpen && (
                <FocusLock returnFocus={true}>
                    <div className="p-3 text-wrap  test-install-extension-popover">
                        <h3 className="mb-0 test-install-extension-popover-header">
                            Take Sourcegraph's code intelligence to {displayName}!
                        </h3>
                        <p className="py-3">
                            Install Sourcegraph browser extension to get code intelligence while browsing files and
                            reading PRs on {displayName}.
                        </p>

                        <div className="mx-auto install-extension-popover__graphic-container d-flex justify-content-between align-items-center">
                            <SourcegraphIcon className="install-extension-popover__logo p-1" />
                            <PlusThickIcon className="install-extension-popover__plus-icon" />
                            <Icon className="install-extension-popover__logo" />
                        </div>

                        <div className="d-flex justify-content-end">
                            <ButtonLink
                                className="btn btn-outline-secondary mr-2"
                                onSelect={onRejection}
                                to={url}
                                {...linkProps}
                            >
                                No, thanks
                            </ButtonLink>

                            <ButtonLink
                                className="btn btn-outline-secondary mr-2"
                                onSelect={onClose}
                                to={url}
                                {...linkProps}
                            >
                                Remind me later
                            </ButtonLink>

                            <ButtonLink
                                className="btn btn-primary mr-2"
                                onSelect={onClickInstall}
                                to="/help/integration/browser_extension"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                Install browser extension
                            </ButtonLink>
                        </div>
                    </div>
                </FocusLock>
            )}
        </Popover>
    )
}
