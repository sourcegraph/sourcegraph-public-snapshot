import React, { useCallback } from 'react'
import Dialog from '@reach/dialog'

import { Form } from '../../../../../branded/src/components/Form'

const getWarningMessage = (codeHostName: string, servicesCount: number): string => {
    const config = {
        multiple: {
            verb: 'are',
            adjective: 'these',
            repoNoun: 'repositories',
        },
        single: {
            verb: 'is',
            adjective: 'it',
            repoNoun: 'repository',
        },
    }

    const { verb, adjective, repoNoun } = servicesCount > 1 ? config.multiple : config.single

    return `There ${verb} ${servicesCount} ${repoNoun} synced to Sourcegraph from ${codeHostName}. If the connection with ${codeHostName} is removed, ${adjective} will no longer be synced with Sourcegraph.`
}

export const RemoveCodeHostConnectionModal: React.FunctionComponent<{
    name: string
    kind: string
    servicesCount: number
    onDidRemove: () => void
    onDidCancel: () => void
}> = ({ onDidRemove, onDidCancel, name, kind, servicesCount }) => {
    const onConnectionRemove = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event.preventDefault()
            onDidRemove()
        },
        [onDidRemove]
    )

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            aria-labelledby={`label--remove-${kind}-token`}
            onDismiss={onDidCancel}
        >
            <div className="web-content">
                <h3 className="text-danger mb-4">Remove connection with {name}?</h3>
                <Form onSubmit={onConnectionRemove}>
                    <div className="form-group mb-4">{getWarningMessage(name, servicesCount)}</div>
                    <div className="d-flex justify-content-end">
                        <button
                            type="button"
                            disabled={false}
                            className="btn btn-outline-secondary mr-2"
                            onClick={onDidCancel}
                        >
                            Cancel
                        </button>
                        <button type="submit" disabled={false} className="btn btn-danger">
                            Yes, remove connection
                        </button>
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
