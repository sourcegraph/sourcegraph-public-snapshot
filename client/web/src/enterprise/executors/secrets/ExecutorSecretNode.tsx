import React, { useCallback, useRef, useState } from 'react'

import { mdiCheckCircleOutline, mdiCheckboxBlankCircleOutline } from '@mdi/js'
import classNames from 'classnames'

import { logger } from '@sourcegraph/common'
import { useLazyQuery } from '@sourcegraph/http-client'
import { Badge, Button, Icon, H3, Tooltip } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import {
    BatchChangesCodeHostFields,
    CheckBatchChangesCredentialResult,
    CheckBatchChangesCredentialVariables,
    ExecutorSecretFields,
    Scalars,
} from '../../../graphql-operations'

import { AddCredentialModal } from './AddSecretModal'
import { CHECK_BATCH_CHANGES_CREDENTIAL } from './backend'
import { RemoveCredentialModal } from './RemoveSecretModal'
import { ViewCredentialModal } from './ViewCredentialModal'

import styles from './ExecutorSecretNode.module.scss'
import LockIcon from 'mdi-react/LockIcon'

export interface ExecutorSecretNodeProps {
    node: ExecutorSecretFields
    refetchAll: () => void
    userID: Scalars['ID'] | null
}

type OpenModal = 'add' | 'view' | 'delete'

export const ExecutorSecretNode: React.FunctionComponent<React.PropsWithChildren<ExecutorSecretNodeProps>> = ({
    node,
    refetchAll,
    userID,
}) => {
    const buttonReference = useRef<HTMLButtonElement | null>(null)

    const [openModal, setOpenModal] = useState<OpenModal | undefined>()
    const onClickAdd = useCallback(() => {
        setOpenModal('add')
    }, [])
    const onClickRemove = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('delete')
    }, [])
    const onClickView = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('view')
    }, [])
    const closeModal = useCallback(() => {
        setOpenModal(undefined)
    }, [])
    const afterAction = useCallback(() => {
        setOpenModal(undefined)
        buttonReference.current?.focus()
        refetchAll()
    }, [refetchAll, buttonReference])

    const isEnabled = true // node.credential !== null && (userID === null || !node.credential.isSiteCredential)

    // const headingAriaLabel = `Sourcegraph ${
    //     isEnabled ? 'has credentials configured' : 'does not have credentials configured'
    // } for ${codeHostDisplayName} (${node.externalServiceURL}).${
    //     !isEnabled && node.credential?.isSiteCredential
    //         ? ' Changesets on this code host will be created with a global token until a personal access token is added.'
    //         : ''
    // }`
    const headingAriaLabel = 'Secret value'

    return (
        <>
            <li className={classNames(styles.ExecutorSecretNodeContainer, 'list-group-item')}>
                <div
                    className={classNames(
                        styles.wrapper,
                        'd-flex justify-content-between align-items-center flex-wrap mb-0'
                    )}
                >
                    <H3 className="text-nowrap mb-0" aria-label={headingAriaLabel}>
                        <Icon className="mx-2" aria-hidden={true} as={LockIcon} /> {node.key}{' '}
                        {!isEnabled && node.credential?.isSiteCredential && (
                            <Badge
                                variant="secondary"
                                tooltip="Changesets on this code host will
                            be created with a global token until a personal access token is added."
                                aria-label="Changesets on this code host will
                            be created with a global token until a personal access token is added."
                            >
                                Global token
                            </Badge>
                        )}
                    </H3>
                    <div className="mb-0 d-flex justify-content-end flex-grow-1 align-items-baseline">
                        {isEnabled ? (
                            <>
                                <Button
                                    // TODO:
                                    // onClick={onClickRemove}
                                    variant="link"
                                    aria-label={`Update secret value for ${node.key}`}
                                    // ref={buttonReference}
                                >
                                    Update
                                </Button>
                                <Button
                                    className="text-danger text-nowrap test-code-host-connection-node-btn-remove"
                                    onClick={onClickRemove}
                                    variant="link"
                                    aria-label={`Remove scret ${node.key}`}
                                    ref={buttonReference}
                                >
                                    Remove
                                </Button>
                            </>
                        ) : (
                            /*
                                a11y-ignore
                                Rule: "color-contrast" (Elements must have sufficient color contrast)
                                GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                            */
                            <Button
                                className="a11y-ignore text-nowrap test-code-host-connection-node-btn-add"
                                onClick={onClickAdd}
                                aria-label={`Add credentials for ${node.key}`}
                                variant="success"
                                ref={buttonReference}
                            >
                                Add credentials
                            </Button>
                        )}
                    </div>
                </div>
            </li>
            {openModal === 'delete' && (
                <RemoveCredentialModal
                    onCancel={closeModal}
                    afterDelete={afterAction}
                    codeHost={node}
                    credential={node.credential!}
                />
            )}
            {openModal === 'view' && (
                <ViewCredentialModal onClose={closeModal} codeHost={node} credential={node.credential!} />
            )}
            {openModal === 'add' && (
                <AddCredentialModal
                    onCancel={closeModal}
                    afterCreate={afterAction}
                    userID={userID}
                    externalServiceKind={node.externalServiceKind}
                    externalServiceURL={node.externalServiceURL}
                    requiresSSH={node.requiresSSH}
                    requiresUsername={node.requiresUsername}
                />
            )}
        </>
    )
}
