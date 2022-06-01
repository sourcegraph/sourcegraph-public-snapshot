import React, { useCallback, useState } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, ErrorLike } from '@sourcegraph/common'
import { Button, Modal, Typography } from '@sourcegraph/wildcard'

import { deleteExternalService } from '../../../components/externalServices/backend'
import { LoaderButton } from '../../../components/LoaderButton'
import { Scalars, ExternalServiceKind } from '../../../graphql-operations'

const getWarningMessage = (serviceName: string, orgName: string, repoCount: number | undefined): string => {
    const membersWillNoLongerSearchAcross = `will no longer be synced and members of ${orgName} will no longer be able to search across`
    const willNotBeSynced = `${membersWillNoLongerSearchAcross} these repositories on Sourcegraph.`

    if (!repoCount) {
        return `If the connection with ${serviceName} is removed, all associated repositories ${willNotBeSynced}`
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

    return `There ${verb} ${repoCount} ${repoNoun} synced to Sourcegraph by ${orgName} from ${serviceName}. If the connection with ${serviceName} is removed, ${
        repoCount > 1 ? adjective + ' repositories' : adjective
    } ${membersWillNoLongerSearchAcross} ${repoCount === 1 ? 'this' : adjective} ${repoNoun} on Sourcegraph.`
}

export const RemoveCodeHostConnectionModal: React.FunctionComponent<
    React.PropsWithChildren<{
        serviceID: Scalars['ID']
        serviceName: string
        orgName: string
        serviceKind: ExternalServiceKind
        repoCount: number | undefined

        onDidRemove: () => void
        onDidCancel: () => void
        onDidError: (error: ErrorLike) => void
    }>
> = ({ serviceID, serviceName, orgName, repoCount, onDidRemove, onDidCancel, onDidError }) => {
    const [isLoading, setIsLoading] = useState(false)

    const onConnectionRemove = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            setIsLoading(true)

            try {
                await deleteExternalService(serviceID)
                onDidRemove()
            } catch (error) {
                setIsLoading(false)
                onDidError(asError(error))
                onDidCancel()
            }
        },
        [serviceID, onDidRemove, onDidError, onDidCancel]
    )

    return (
        <Modal
            aria-labelledby={`heading--remove-${serviceName}-code-host`}
            aria-describedby={`description--remove-${serviceName}-code-host`}
            onDismiss={onDidCancel}
        >
            <Typography.H3 id={`heading--remove-${serviceName}-code-host`} className="text-danger mb-4">
                Remove connection with {serviceName}?
            </Typography.H3>
            <Form onSubmit={onConnectionRemove}>
                <div id={`description--remove-${serviceName}-code-host`} className="form-group mb-4">
                    {getWarningMessage(serviceName, orgName, repoCount)}
                </div>
                <div className="d-flex justify-content-end">
                    <Button
                        disabled={isLoading}
                        className="mr-2"
                        onClick={onDidCancel}
                        outline={true}
                        variant="secondary"
                    >
                        Cancel
                    </Button>
                    <LoaderButton
                        type="submit"
                        loading={isLoading}
                        disabled={isLoading}
                        label="Yes, remove connection"
                        alwaysShowLabel={true}
                        variant="danger"
                    />
                </div>
            </Form>
        </Modal>
    )
}
