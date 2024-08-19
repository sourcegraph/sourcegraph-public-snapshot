import { useCallback, useEffect, useMemo, useState, type FunctionComponent } from 'react'

import { useNavigate } from 'react-router-dom'

import { useMutation } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, ErrorAlert, Link, screenReaderAnnounce } from '@sourcegraph/wildcard'

import { PromptVisibility, type CreatePromptResult, type CreatePromptVariables } from '../graphql-operations'
import { NamespaceSelector } from '../namespaces/NamespaceSelector'
import { namespaceTelemetryMetadata } from '../namespaces/telemetry'
import { useAffiliatedNamespaces } from '../namespaces/useAffiliatedNamespaces'
import { PageRoutes } from '../routes.constants'

import { PromptForm, type PromptFormValue } from './Form'
import { createPromptMutation } from './graphql'

/**
 * Form to create a new prompt.
 */
export const NewForm: FunctionComponent<TelemetryV2Props> = ({ telemetryRecorder }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('prompts.new', 'view')
    }, [telemetryRecorder])

    const {
        namespaces,
        initialNamespace,
        loading: namespacesLoading,
        error: namespacesError,
    } = useAffiliatedNamespaces()
    const [selectedNamespace, setSelectedNamespace] = useState<string | undefined>()
    const selectedNamespaceOrInitial = selectedNamespace ?? initialNamespace?.id

    const [createPrompt, { loading, error }] = useMutation<CreatePromptResult, CreatePromptVariables>(
        createPromptMutation,
        {}
    )

    const navigate = useNavigate()
    const onSubmit = useCallback(
        async (fields: PromptFormValue): Promise<void> => {
            try {
                const { data } = await createPrompt({
                    variables: {
                        input: {
                            name: fields.name,
                            description: fields.description,
                            definitionText: fields.definitionText,
                            draft: fields.draft,
                            owner: selectedNamespaceOrInitial!,
                            visibility: PromptVisibility.SECRET,
                        },
                    },
                })
                const created = data?.createPrompt
                if (!created) {
                    return
                }
                telemetryRecorder.recordEvent('prompts', 'create', {
                    metadata: namespaceTelemetryMetadata(created.owner),
                })
                screenReaderAnnounce(`Created new prompt: ${created.description}`)
                navigate(created.url)
            } catch {
                // Mutation error is read in useMutation call.
            }
        },
        [createPrompt, selectedNamespaceOrInitial, telemetryRecorder, navigate]
    )

    const initialValue = useMemo<Partial<PromptFormValue>>(() => ({}), [])

    return namespacesError ? (
        <ErrorAlert error={namespacesError} />
    ) : (
        <PromptForm
            submitLabel="Create prompt"
            onSubmit={onSubmit}
            otherButtons={
                <Button as={Link} variant="secondary" outline={true} to={PageRoutes.Prompts}>
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
