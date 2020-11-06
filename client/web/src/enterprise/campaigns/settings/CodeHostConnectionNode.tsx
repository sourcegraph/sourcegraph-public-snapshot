import React, { useCallback, useState } from 'react'
import * as H from 'history'
import CheckboxBlankCircleOutlineIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import CheckCircleOutlineIcon from 'mdi-react/CheckCircleOutlineIcon'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { CampaignsCodeHostFields } from '../../../graphql-operations'
import { AddCredentialModal } from './AddCredentialModal'
import { RemoveCredentialModal } from './RemoveCredentialModal'
import { Subject } from 'rxjs'

export interface CodeHostConnectionNodeProps {
    node: CampaignsCodeHostFields
    history: H.History
    updateList: Subject<void>
}

export const CodeHostConnectionNode: React.FunctionComponent<CodeHostConnectionNodeProps> = ({
    node,
    history,
    updateList,
}) => {
    const Icon = defaultExternalServices[node.externalServiceKind].icon
    const [showDeleteModal, setShowDeleteModal] = useState<boolean>(false)

    const [showAddModal, setShowAddModal] = useState<boolean>(false)
    const onClickAdd = useCallback(() => {
        setShowAddModal(true)
    }, [])
    const onCancelAdd = useCallback(() => {
        setShowAddModal(false)
    }, [])
    const afterCreate = useCallback(() => {
        setShowAddModal(false)
        updateList.next()
    }, [updateList])

    const onRemove = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setShowDeleteModal(true)
    }, [])
    const onCancelDelete = useCallback(() => {
        setShowDeleteModal(false)
    }, [])
    const afterDelete = useCallback(() => {
        setShowDeleteModal(false)
        updateList.next()
    }, [updateList])
    return (
        <>
            <li className="list-group-item p-3">
                <div className="d-flex justify-content-between align-content-center mb-0">
                    <h3 className="mb-0">
                        <CodeHostConnectionState enabled={node.credential !== null} />
                        <Icon className="icon-inline mx-2" /> {node.externalServiceURL}
                    </h3>
                    <div className="mb-0">
                        {node.credential !== null && (
                            <a href="" className="btn btn-link text-danger" onClick={onRemove}>
                                Remove
                            </a>
                        )}
                        {node.credential === null && (
                            <button type="button" className="btn btn-success" onClick={onClickAdd}>
                                Add token
                            </button>
                        )}
                    </div>
                </div>
            </li>
            {showDeleteModal && (
                <RemoveCredentialModal
                    afterDelete={afterDelete}
                    credentialID={node.credential!.id}
                    history={history}
                    onCancel={onCancelDelete}
                />
            )}
            {showAddModal && (
                <AddCredentialModal
                    onCancel={onCancelAdd}
                    afterCreate={afterCreate}
                    externalServiceKind={node.externalServiceKind}
                    externalServiceURL={node.externalServiceURL}
                    history={history}
                />
            )}
        </>
    )
}

interface CodeHostConnectionStateProps {
    enabled: boolean
}

const CodeHostConnectionState: React.FunctionComponent<CodeHostConnectionStateProps> = ({ enabled }) => (
    <>
        {enabled && <CheckCircleOutlineIcon className="text-success icon-inline" />}
        {!enabled && <CheckboxBlankCircleOutlineIcon className="text-danger icon-inline" />}
    </>
)
