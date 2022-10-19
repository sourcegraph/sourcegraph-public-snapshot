import React, { useCallback, useRef, useState } from 'react'

import classNames from 'classnames'
import LockIcon from 'mdi-react/LockIcon'

import { Badge, Button, Icon, H3 } from '@sourcegraph/wildcard'

import { ExecutorSecretFields } from '../../../graphql-operations'

import { RemoveSecretModal } from './RemoveSecretModal'
import { UpdateSecretModal } from './UpdateSecretModal'

import styles from './ExecutorSecretNode.module.scss'

export interface ExecutorSecretNodeProps {
    node: ExecutorSecretFields
    refetchAll: () => void
}

type OpenModal = 'update' | 'delete'

export const ExecutorSecretNode: React.FunctionComponent<React.PropsWithChildren<ExecutorSecretNodeProps>> = ({
    node,
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
            <li className={classNames(styles.ExecutorSecretNodeContainer, 'list-group-item')}>
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
                            >
                                Global secret
                            </Badge>
                        )}
                    </H3>
                    <div className="mb-0 d-flex justify-content-end flex-grow-1 align-items-baseline">
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
                    </div>
                </div>
            </li>
            {openModal === 'delete' && (
                <RemoveSecretModal onCancel={closeModal} afterDelete={afterAction} secret={node} />
            )}
            {openModal === 'update' && (
                <UpdateSecretModal onCancel={closeModal} afterUpdate={afterAction} secret={node} />
            )}
        </>
    )
}
