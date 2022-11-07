import React, { useCallback, useRef, useState } from 'react'

import classNames from 'classnames'
import LockIcon from 'mdi-react/LockIcon'

import { Badge, Button, Icon, H3, Link } from '@sourcegraph/wildcard'

import { ExecutorSecretFields, Scalars } from '../../../graphql-operations'

import { RemoveSecretModal } from './RemoveSecretModal'
import { SecretAccessLogsModal } from './SecretAccessLogsModal'
import { UpdateSecretModal } from './UpdateSecretModal'

import styles from './ExecutorSecretNode.module.scss'

export interface ExecutorSecretNodeProps {
    node: ExecutorSecretFields
    namespaceID: Scalars['ID'] | null
    refetchAll: () => void
}

type OpenModal = 'update' | 'delete' | 'accessLogs'

export const ExecutorSecretNode: React.FunctionComponent<React.PropsWithChildren<ExecutorSecretNodeProps>> = ({
    node,
    namespaceID,
    refetchAll,
}) => {
    const buttonReference = useRef<HTMLButtonElement | null>(null)

    const [openModal, setOpenModal] = useState<OpenModal | undefined>()

    const onClickRemove = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('delete')
    }, [])
    const onClickUpdate = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('update')
    }, [])
    const onClickAccessLogs = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setOpenModal('accessLogs')
    }, [])
    const closeModal = useCallback(() => {
        setOpenModal(undefined)
    }, [])
    const afterAction = useCallback(() => {
        setOpenModal(undefined)
        buttonReference.current?.focus()
        refetchAll()
    }, [refetchAll, buttonReference])

    const headingAriaLabel = 'Secret value'

    return (
        <>
            <li className={classNames(styles.executorSecretNodeContainer, 'list-group-item')}>
                <div
                    className={classNames(
                        styles.wrapper,
                        'd-flex justify-content-between align-items-center flex-wrap mb-0'
                    )}
                >
                    <H3 className="text-nowrap mb-0" aria-label={headingAriaLabel}>
                        <Icon className="mx-2" aria-hidden={true} as={LockIcon} /> {node.key}{' '}
                        {node.namespace === null && (
                            <Badge
                                variant="secondary"
                                tooltip="This secret will be usable by all users of the Sourcegraph instance."
                                aria-label="This secret will be usable by all users of the Sourcegraph instance."
                                className="mr-2"
                            >
                                Global secret
                            </Badge>
                        )}
                        {node.overwritesGlobalSecret && (
                            <Badge
                                variant="secondary"
                                tooltip="This secret overwrites an existing secret set globally in this Sourcegraph instance."
                                aria-label="This secret overwrites an existing secret set globally in this Sourcegraph instance."
                                className="mr-2"
                            >
                                Overwrites global secret
                            </Badge>
                        )}
                        <small>
                            By {node.creator && <Link to={node.creator.url}>{node.creator.username}</Link>}
                            {!node.creator && <>deleted user</>}
                        </small>
                    </H3>
                    <div className="mb-0 d-flex justify-content-end flex-grow-1 align-items-baseline">
                        <Button
                            onClick={onClickAccessLogs}
                            variant="link"
                            aria-label={`View access logs for secret ${node.key}`}
                        >
                            Access logs
                        </Button>
                        {(namespaceID === null || (namespaceID !== null && node.namespace !== null)) && (
                            <>
                                <Button
                                    onClick={onClickUpdate}
                                    variant="link"
                                    aria-label={`Update secret value for ${node.key}`}
                                    ref={buttonReference}
                                >
                                    Update
                                </Button>
                                <Button
                                    className="text-danger text-nowrap"
                                    onClick={onClickRemove}
                                    variant="link"
                                    aria-label={`Remove scret ${node.key}`}
                                >
                                    Remove
                                </Button>
                            </>
                        )}
                    </div>
                </div>
            </li>
            {openModal === 'delete' && (
                <RemoveSecretModal onCancel={closeModal} afterDelete={afterAction} secret={node} />
            )}
            {openModal === 'update' && (
                <UpdateSecretModal onCancel={closeModal} afterUpdate={afterAction} secret={node} />
            )}
            {openModal === 'accessLogs' && <SecretAccessLogsModal onCancel={closeModal} secretID={node.id} />}
        </>
    )
}
