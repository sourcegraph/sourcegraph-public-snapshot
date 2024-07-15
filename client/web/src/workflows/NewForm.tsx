import { useCallback, useEffect, useMemo, useState, type FunctionComponent } from 'react'

import { useNavigate } from 'react-router-dom'

import { useMutation } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, ErrorAlert, Link, screenReaderAnnounce } from '@sourcegraph/wildcard'

import type { CreateWorkflowResult, CreateWorkflowVariables } from '../graphql-operations'
import { NamespaceSelector } from '../namespaces/NamespaceSelector'
import { namespaceTelemetryMetadata } from '../namespaces/telemetry'
import { useAffiliatedNamespaces } from '../namespaces/useAffiliatedNamespaces'
import { PageRoutes } from '../routes.constants'

import { WorkflowForm, type WorkflowFormValue } from './Form'
import { createWorkflowMutation } from './graphql'

/**
 * Form to create a new workflow.
 */
export const NewForm: FunctionComponent<
    TelemetryV2Props & {
        isSourcegraphDotCom: boolean
    }
> = ({ isSourcegraphDotCom, telemetryRecorder }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('workflows.new', 'view')
    }, [telemetryRecorder])

    const {
        namespaces,
        initialNamespace,
        loading: namespacesLoading,
        error: namespacesError,
    } = useAffiliatedNamespaces()
    const [selectedNamespace, setSelectedNamespace] = useState<string | undefined>()
    const selectedNamespaceOrInitial = selectedNamespace ?? initialNamespace?.id

    const [createWorkflow, { loading, error }] = useMutation<CreateWorkflowResult, CreateWorkflowVariables>(
        createWorkflowMutation,
        {}
    )

    const navigate = useNavigate()
    const onSubmit = useCallback(
        async (fields: WorkflowFormValue): Promise<void> => {
            try {
                const { data } = await createWorkflow({
                    variables: {
                        input: {
                            name: fields.name,
                            description: fields.description,
                            templateText: fields.templateText,
                            draft: fields.draft,
                            owner: selectedNamespaceOrInitial!,
                        },
                    },
                })
                const created = data?.createWorkflow
                if (!created) {
                    return
                }
                telemetryRecorder.recordEvent('workflows', 'create', {
                    metadata: namespaceTelemetryMetadata(created.owner),
                })
                screenReaderAnnounce(`Created new workflow: ${created.description}`)
                navigate(created.url)
            } catch {
                // Mutation error is read in useMutation call.
            }
        },
        [createWorkflow, selectedNamespaceOrInitial, telemetryRecorder, navigate]
    )

    const initialValue = useMemo<Partial<WorkflowFormValue>>(() => ({}), [])

    return namespacesError ? (
        <ErrorAlert error={namespacesError} />
    ) : (
        <WorkflowForm
            submitLabel="Create workflow"
            onSubmit={onSubmit}
            otherButtons={
                <Button as={Link} variant="secondary" outline={true} to={PageRoutes.Workflows}>
                    Cancel
                </Button>
            }
            initialValue={initialValue}
            namespaceField={
                <NamespaceSelector
                    namespaces={namespaces}
                    value={selectedNamespaceOrInitial}
                    onSelect={namespace => setSelectedNamespace(namespace)}
                    disabled={loading || namespacesLoading}
                    loading={namespacesLoading}
                    label="Owner"
                    className="w-fit-content"
                />
            }
            loading={loading}
            error={error}
            telemetryRecorder={telemetryRecorder}
        />
    )
}
