import ExportIcon from 'mdi-react/ExportIcon'
import PlusThickIcon from 'mdi-react/PlusThickIcon'
import React, { useMemo } from 'react'
import { ButtonLink } from '../../../../shared/src/components/LinkOrButton'
import { SourcegraphIcon } from '../../auth/icons'
import { PopoverContainer } from '../../components/PopoverContainer'
import { serviceTypeDisplayNameAndIcon } from './GoToCodeHostAction'

interface CodeHostExtensionPopoverProps {
    url: string
    serviceType: string | null
    onClose: () => void
    onRejection: () => void
    onClickInstall: () => void
    targetID: string
}

export const InstallExtensionPopover: React.FunctionComponent<CodeHostExtensionPopoverProps> = ({
    url,
    serviceType,
    onClose,
    onRejection,
    onClickInstall,
    targetID,
}) => {
    const { displayName, icon } = serviceTypeDisplayNameAndIcon(serviceType)
    const Icon = icon || ExportIcon

    return (
        <PopoverContainer
            onClose={onClose}
            targetID={targetID}
            popperOptions={useMemo(
                () => ({
                    placement: 'bottom-start',
                    modifiers: [
                        {
                            name: 'offset',
                            options: {
                                offset: [64, 4],
                            },
                        },
                    ],
                }),
                []
            )}
        >
            {modalBodyReference => (
                <div
                    ref={modalBodyReference as React.MutableRefObject<HTMLDivElement>}
                    // TODO: move .extension-permission-modal to .modal-body
                    className="extension-permission-modal p-4 web-content text-wrap border shadow test-install-extension-popover"
                >
                    <h3 className="mb-0 test-install-extension-popover-header">
                        Take Sourcegraph's code intelligence to {displayName}!
                    </h3>
                    <p className="py-3">
                        Install Sourcegraph browser extension to get code intelligence while browsing files and reading
                        PRs on {displayName}.
                    </p>

                    <div className="mx-auto install-extension-popover__graphic-container d-flex justify-content-between align-items-center">
                        <SourcegraphIcon className="install-extension-popover__logo p-1" />
                        <PlusThickIcon className="install-extension-popover__plus-icon" />
                        <Icon className="install-extension-popover__logo" />
                    </div>

                    <div className="d-flex justify-content-end">
                        <ButtonLink className="btn btn-outline-secondary mr-2" onSelect={onRejection} to={url}>
                            No, thanks
                        </ButtonLink>

                        <ButtonLink className="btn btn-outline-secondary mr-2" onSelect={onClose} to={url}>
                            Remind me later
                        </ButtonLink>

                        <ButtonLink
                            className="btn btn-primary mr-2"
                            onSelect={onClickInstall}
                            to="/help/integration/browser_extension"
                        >
                            Install browser extension
                        </ButtonLink>
                    </div>
                </div>
            )}
        </PopoverContainer>
    )
}
