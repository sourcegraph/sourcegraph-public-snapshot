import React, { useCallback, useRef, useState } from 'react'

import classNames from 'classnames'
import CheckboxBlankCircleOutlineIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import CheckCircleOutlineIcon from 'mdi-react/CheckCircleOutlineIcon'

import { Badge, Button, Icon, Typography } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { BatchChangesCodeHostFields, Scalars } from '../../../graphql-operations'

import { AddCredentialModal } from './AddCredentialModal'
import { RemoveCredentialModal } from './RemoveCredentialModal'
import { ViewCredentialModal } from './ViewCredentialModal'

import styles from './CodeHostConnectionNode.module.scss'

export interface CodeHostConnectionNodeProps {
    node: BatchChangesCodeHostFields
    refetchAll: () => void
    userID: Scalars['ID'] | null
}

type OpenModal = 'add' | 'view' | 'delete'

export const CodeHostConnectionNode: React.FunctionComponent<React.PropsWithChildren<CodeHostConnectionNodeProps>> = ({
    node,
    refetchAll,
    userID,
}) => {
    const ExternalServiceIcon = defaultExternalServices[node.externalServiceKind].icon
    const codeHostDisplayName = defaultExternalServices[node.externalServiceKind].defaultDisplayName

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

    const isEnabled = node.credential !== null && (userID === null || !node.credential.isSiteCredential)

    const headingAriaLabel = `Sourcegraph ${
        isEnabled ? 'has credentials configured' : 'does not have credentials configured'
    } for ${codeHostDisplayName} (${node.externalServiceURL}).${
        !isEnabled && node.credential?.isSiteCredential
            ? ' Changesets on this code host will be created with a global token until a personal access token is added.'
            : ''
    }`

    return (
        <>
            <li
                className={classNames(
                    styles.codeHostConnectionNodeContainer,
                    'list-group-item test-code-host-connection-node'
                )}
            >
                <div
                    className={classNames(
                        styles.wrapper,
                        'd-flex justify-content-between align-items-center flex-wrap mb-0'
                    )}
                >
                    <Typography.H3 className="text-nowrap mb-0" aria-label={headingAriaLabel}>
                        {isEnabled && (
                            <Icon
                                className="text-success test-code-host-connection-node-enabled"
                                data-tooltip="This code host has credentials connected."
                                aria-label="This code host has credentials connected."
                                as={CheckCircleOutlineIcon}
                                role="img"
                            />
                        )}
                        {!isEnabled && (
                            <Icon
                                className="text-danger test-code-host-connection-node-disabled"
                                data-tooltip="This code host does not have credentials configured."
                                aria-label="This code host does not have credentials configured."
                                as={CheckboxBlankCircleOutlineIcon}
                                role="img"
                            />
                        )}
                        <Icon className="mx-2" role="img" aria-hidden={true} as={ExternalServiceIcon} />{' '}
                        {node.externalServiceURL}{' '}
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
                    </Typography.H3>
                    <div className="mb-0 d-flex justify-content-end flex-grow-1">
                        {isEnabled ? (
                            <>
                                <Button
                                    className="text-danger text-nowrap test-code-host-connection-node-btn-remove"
                                    onClick={onClickRemove}
                                    variant="link"
                                    aria-label={`Remove credentials for ${codeHostDisplayName}`}
                                    ref={buttonReference}
                                >
                                    Remove
                                </Button>
                                {node.requiresSSH && (
                                    <Button onClick={onClickView} className="text-nowrap ml-2" variant="secondary">
                                        View public key
                                    </Button>
                                )}
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
                                aria-label={`Add credentials for ${codeHostDisplayName}`}
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
