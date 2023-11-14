import React, { useCallback, useMemo, useEffect } from 'react'

import classNames from 'classnames'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Modal, Button, Checkbox, H3 } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'

import { NotebookShareOptionsDropdown, type ShareOption } from './NotebookShareOptionsDropdown'

import styles from './ShareNotebookModal.module.scss'

interface ShareNotebookModalProps extends TelemetryProps, TelemetryV2Props {
    isSourcegraphDotCom: boolean
    selectedShareOption: ShareOption
    setSelectedShareOption: (option: ShareOption) => void
    isOpen: boolean
    toggleModal: () => void
    authenticatedUser: AuthenticatedUser
    onUpdateVisibility: (isPublic: boolean, namespace: string) => void
}

function getSelectedShareOptionDescription(shareOption: ShareOption, isSourcegraphDotCom: boolean): string {
    if (shareOption.namespaceType === 'User') {
        const withAccess = isSourcegraphDotCom ? 'on Sourcegraph.com' : 'with access to the Sourcegraph instance'
        return shareOption.isPublic
            ? `Everyone ${withAccess} can view the notebook, but only you can edit it`
            : 'Only you can view and edit the notebook'
    }
    return `Only members of the ${shareOption.namespaceName} organization can edit the notebook`
}

export const ShareNotebookModal: React.FunctionComponent<React.PropsWithChildren<ShareNotebookModalProps>> = ({
    isOpen,
    isSourcegraphDotCom,
    selectedShareOption,
    setSelectedShareOption,
    toggleModal,
    authenticatedUser,
    telemetryService,
    telemetryRecorder,
    onUpdateVisibility,
}) => {
    useEffect(() => {
        if (isOpen) {
            telemetryService.log('SearchNotebookShareModalOpened')
            telemetryRecorder.recordEvent('SearchNotebookShareModal', 'opened')
        }
    }, [isOpen, telemetryService, telemetryRecorder])

    const shareLabelId = 'shareNotebookId'

    const description = useMemo(
        () => getSelectedShareOptionDescription(selectedShareOption, isSourcegraphDotCom),
        [selectedShareOption, isSourcegraphDotCom]
    )

    const onDoneClick = useCallback((): void => {
        onUpdateVisibility(selectedShareOption.isPublic, selectedShareOption.namespaceId)
        toggleModal()
    }, [toggleModal, onUpdateVisibility, selectedShareOption])

    return (
        <Modal isOpen={isOpen} position="top-third" onDismiss={toggleModal} aria-labelledby={shareLabelId}>
            <H3 id={shareLabelId}>Share Notebook</H3>
            <div className={classNames('mb-2', styles.body)}>
                <NotebookShareOptionsDropdown
                    isSourcegraphDotCom={isSourcegraphDotCom}
                    telemetryService={telemetryService}
                    authenticatedUser={authenticatedUser}
                    selectedShareOption={selectedShareOption}
                    onSelectShareOption={setSelectedShareOption}
                />
                <div className="text-muted mt-1">
                    <small>{description}</small>
                </div>
                {selectedShareOption.namespaceType === 'Org' && (
                    <Checkbox
                        id="org-namespace-visibility"
                        checked={selectedShareOption.isPublic}
                        wrapperClassName="mt-2"
                        onChange={event =>
                            setSelectedShareOption({
                                ...selectedShareOption,
                                isPublic: event.target.checked,
                            })
                        }
                        label={`Everyone ${
                            isSourcegraphDotCom ? 'on Sourcegraph.com' : 'with access to the Sourcegraph instance'
                        } can view the notebook`}
                    />
                )}
            </div>
            <div className="text-right">
                <Button className="mr-1" variant="secondary" outline={true} size="sm" onClick={toggleModal}>
                    Cancel
                </Button>
                <Button variant="primary" size="sm" onClick={onDoneClick} data-testid="share-notebook-done-button">
                    Done
                </Button>
            </div>
        </Modal>
    )
}
