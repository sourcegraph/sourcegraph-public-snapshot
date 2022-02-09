import React, { useState, useCallback, useMemo } from 'react'
import { Dropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { NotebookFields } from '../../graphql-operations'

import { NotebookShareOptionsDropdown, ShareOption } from './NotebookVisibilitySettingsDropdown'
import styles from './ShareNotebookDropdown.module.scss'

interface ShareNotebookDropdownProps extends TelemetryProps {
    isPublic: boolean
    authenticatedUser: AuthenticatedUser
    namespace: NonNullable<NotebookFields['namespace']>
    onUpdateVisibility: (isPublic: boolean, namespace: string) => void
}

function getSelectedShareOptionDescription(shareOption: ShareOption): string {
    if (shareOption.namespaceType === 'User') {
        return shareOption.isPublic
            ? 'Everyone can view the notebook, but only you can edit it'
            : 'Only you can view and edit the notebook'
    }
    return `Only members of the ${shareOption.namespaceName} organization can edit the notebook`
}

export const ShareNotebookDropdown: React.FunctionComponent<ShareNotebookDropdownProps> = ({
    isPublic: initialIsPublic,
    authenticatedUser,
    telemetryService,
    namespace,
    onUpdateVisibility,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const [selectedShareOption, setSelectedShareOption] = useState<ShareOption>({
        namespaceType: namespace.__typename,
        namespaceId: namespace.id,
        namespaceName: namespace.namespaceName,
        isPublic: initialIsPublic,
    })

    const toggleOpen = useCallback(() => {
        telemetryService.log('ShareNotebookDropdownToggled')
        setIsOpen(value => !value)
    }, [telemetryService])

    const description = useMemo(() => getSelectedShareOptionDescription(selectedShareOption), [selectedShareOption])

    const onDoneClick = useCallback((): void => {
        onUpdateVisibility(selectedShareOption.isPublic, selectedShareOption.namespaceId)
        setIsOpen(false)
    }, [setIsOpen, onUpdateVisibility, selectedShareOption])

    return (
        <Dropdown isOpen={isOpen} toggle={toggleOpen}>
            <DropdownToggle color="primary">Share</DropdownToggle>
            <DropdownMenu right={true}>
                <div className={styles.wrapper}>
                    <div className="mb-2">
                        <strong>Share Notebook</strong>
                    </div>
                    <div className="mb-2">
                        <NotebookShareOptionsDropdown
                            telemetryService={telemetryService}
                            authenticatedUser={authenticatedUser}
                            selectedShareOption={selectedShareOption}
                            onSelectShareOption={setSelectedShareOption}
                        />
                        <div className="text-muted mt-1">
                            <small>{description}</small>
                        </div>
                        {selectedShareOption.namespaceType === 'Org' && (
                            <div className="form-check mt-1">
                                <input
                                    className="form-check-input"
                                    type="checkbox"
                                    id="org-namespace-visibility"
                                    checked={selectedShareOption.isPublic}
                                    onChange={event =>
                                        setSelectedShareOption({
                                            ...selectedShareOption,
                                            isPublic: event.target.checked,
                                        })
                                    }
                                />
                                <label className="form-check-label" htmlFor="org-namespace-visibility">
                                    Everyone can view the notebook
                                </label>
                            </div>
                        )}
                    </div>
                    <div className="text-right">
                        <Button
                            className="mr-1"
                            variant="secondary"
                            outline={true}
                            size="sm"
                            onClick={() => setIsOpen(false)}
                        >
                            Cancel
                        </Button>
                        <Button variant="primary" size="sm" onClick={onDoneClick}>
                            Done
                        </Button>
                    </div>
                </div>
            </DropdownMenu>
        </Dropdown>
    )
}
