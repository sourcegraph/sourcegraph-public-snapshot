import React, { useCallback, useState } from 'react'
import Dialog from '@reach/dialog'

import { Form } from '../../../../../branded/src/components/Form'
import { LoaderButton } from '../../../components/LoaderButton'
import { deleteExternalService } from '../../../components/externalServices/backend'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { Scalars, ExternalServiceKind } from '../../../graphql-operations'

const getWarningMessage = (codeHostName: string, repoCount: number | undefined): string => {
    if (!repoCount) {
        return `If the connection with ${codeHostName} is removed, all associated repositories  will no longer be synced with Sourcegraph.`
    }

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

    const { verb, adjective, repoNoun } = repoCount > 1 ? config.multiple : config.single

    return `There ${verb} ${repoCount} ${repoNoun} synced to Sourcegraph from ${codeHostName}. If the connection with ${codeHostName} is removed, ${adjective} will no longer be synced with Sourcegraph.`
}

export const RemoveCodeHostConnectionModal: React.FunctionComponent<{
    id: Scalars['ID']
    name: string
    kind: ExternalServiceKind
    repoCount: number | undefined

    onDidRemove: () => void
    onDidCancel: () => void
    onDidError: (error: ErrorLike) => void
}> = ({ id, name, repoCount, onDidRemove, onDidCancel, onDidError }) => {
    const [isLoading, setIsLoading] = useState(false)

    const onConnectionRemove = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            setIsLoading(true)

            try {
                await deleteExternalService(id)
                onDidRemove()
            } catch (error) {
                setIsLoading(false)
                onDidError(asError(error))
                onDidCancel()
            }
        },
        [id, onDidRemove, onDidError, onDidCancel]
    )

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            aria-labelledby={`heading--remove-${name}-code-host`}
            aria-describedby={`description--remove-${name}-code-host`}
            onDismiss={onDidCancel}
        >
            <div className="web-content">
                <h3 id={`heading--remove-${name}-code-host`} className="text-danger mb-4">
                    Remove connection with {name}?
                </h3>
                <Form onSubmit={onConnectionRemove}>
                    <div id={`description--remove-${name}-code-host`} className="form-group mb-4">
                        {getWarningMessage(name, repoCount)}
                    </div>
                    <div className="d-flex justify-content-end">
                        <button
                            type="button"
                            disabled={isLoading}
                            className="btn btn-outline-secondary mr-2"
                            onClick={onDidCancel}
                        >
                            Cancel
                        </button>
                        <LoaderButton
                            type="submit"
                            className="btn btn-danger"
                            loading={isLoading}
                            disabled={isLoading}
                            label="Yes, remove connection"
                            alwaysShowLabel={true}
                        />
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
