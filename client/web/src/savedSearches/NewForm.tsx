import { useCallback, useEffect, useState, type FunctionComponent } from 'react'

import { useNavigate } from 'react-router-dom'

import { useMutation } from '@sourcegraph/http-client'
import { useSettingsCascade } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, ErrorAlert, Link, screenReaderAnnounce } from '@sourcegraph/wildcard'

import {
    SavedSearchVisibility,
    type CreateSavedSearchResult,
    type CreateSavedSearchVariables,
} from '../graphql-operations'
import { NamespaceSelector } from '../namespaces/NamespaceSelector'
import { namespaceTelemetryMetadata } from '../namespaces/telemetry'
import { useAffiliatedNamespaces } from '../namespaces/useAffiliatedNamespaces'
import { PageRoutes } from '../routes.constants'
import { defaultPatternTypeFromSettings } from '../util/settings'

import { SavedSearchForm, type SavedSearchFormValue } from './Form'
import { createSavedSearchMutation } from './graphql'

/**
 * Form to create a new saved search.
 */
export const NewForm: FunctionComponent<
    TelemetryV2Props & {
        isSourcegraphDotCom: boolean
    }
> = ({ isSourcegraphDotCom, telemetryRecorder }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('savedSearches.new', 'view')
    }, [telemetryRecorder])

    const {
        namespaces,
        initialNamespace,
        loading: namespacesLoading,
        error: namespacesError,
    } = useAffiliatedNamespaces()
    const [selectedNamespace, setSelectedNamespace] = useState<string | undefined>()
    const selectedNamespaceOrInitial = selectedNamespace ?? initialNamespace?.id

    const [createSavedSearch, { loading, error }] = useMutation<CreateSavedSearchResult, CreateSavedSearchVariables>(
        createSavedSearchMutation,
        {}
    )

    const navigate = useNavigate()
    const onSubmit = useCallback(
        async (fields: SavedSearchFormValue): Promise<void> => {
            try {
                const { data } = await createSavedSearch({
                    variables: {
                        input: {
                            owner: selectedNamespaceOrInitial!,
                            description: fields.description,
                            query: fields.query,
                            draft: fields.draft,
                            visibility: SavedSearchVisibility.SECRET,
                        },
                    },
                })
                const created = data?.createSavedSearch
                if (!created) {
                    return
                }
                telemetryRecorder.recordEvent('savedSearches', 'create', {
                    metadata: namespaceTelemetryMetadata(created.owner),
                })
                screenReaderAnnounce(`Created new saved search: ${created.description}`)
                navigate(created.url)
            } catch {
                // Mutation error is read in useMutation call.
            }
        },
        [createSavedSearch, selectedNamespaceOrInitial, telemetryRecorder, navigate]
    )

    const searchParameters = new URLSearchParams(location.search)
    const query = searchParameters.get('query')
    const settingsCascade = useSettingsCascade()
    const patternType = searchParameters.get('patternType') ?? defaultPatternTypeFromSettings(settingsCascade)
    const defaultValue: Partial<SavedSearchFormValue> = {
        query: [patternType ? `patternType:${patternType} ` : null, query].filter(Boolean).join(''),
    }

    return namespacesError ? (
        <ErrorAlert error={namespacesError} />
    ) : (
        <SavedSearchForm
            isSourcegraphDotCom={isSourcegraphDotCom}
            submitLabel="Create saved search"
            onSubmit={onSubmit}
            otherButtons={
                <Button as={Link} variant="secondary" outline={true} to={PageRoutes.SavedSearches}>
                    Cancel
                </Button>
            }
            initialValue={defaultValue}
            loading={loading}
            error={error}
            telemetryRecorder={telemetryRecorder}
            beforeFields={
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
        />
    )
}
