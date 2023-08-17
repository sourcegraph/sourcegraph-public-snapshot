import { type FC, useState } from 'react'

import { useDebounce } from 'use-debounce'

import { useMutation, useQuery } from '@sourcegraph/http-client'
import {
    Container,
    ErrorAlert,
    Text,
    Label,
    Form,
    Alert,
    H2,
    Combobox,
    ComboboxInput,
    ComboboxPopover,
    ComboboxList,
    ComboboxOption,
} from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import type {
    AddRepoMetadataResult,
    AddRepoMetadataVariables,
    SearchRepoMetaKeysResult,
    SearchRepoMetaKeysVariables,
    SearchRepoMetaValuesResult,
    SearchRepoMetaValuesVariables,
} from '../../graphql-operations'

import { SEARCH_REPO_META_KEYS_GQL, ADD_REPO_METADATA_GQL, SEARCH_REPO_META_VALUES_GQL } from './query'

function useThrottle<T>(value: T, delay: number): T {
    const [throttledValue] = useDebounce(value, delay, { leading: true, maxWait: delay })

    return throttledValue
}

function useKeySuggestions(query: string, delay = 300): { suggestions: string[]; loading: boolean } {
    const throttledQuery = useThrottle(query, delay)
    const { data, loading } = useQuery<SearchRepoMetaKeysResult, SearchRepoMetaKeysVariables>(
        SEARCH_REPO_META_KEYS_GQL,
        {
            variables: { query: throttledQuery },
            skip: !throttledQuery,
            fetchPolicy: 'network-only',
        }
    )

    return {
        suggestions: data?.repoMeta?.keys?.nodes ?? [],
        loading,
    }
}

function useValueSuggestions(key: string, query: string, delay = 300): { suggestions: string[]; loading: boolean } {
    const [throttledKey, throttledQuery] = useThrottle([key, query], delay)
    const { data, loading } = useQuery<SearchRepoMetaValuesResult, SearchRepoMetaValuesVariables>(
        SEARCH_REPO_META_VALUES_GQL,
        {
            variables: { key: throttledKey, query: throttledQuery || null },
            skip: !throttledKey,
            fetchPolicy: 'network-only',
        }
    )

    return {
        suggestions: data?.repoMeta?.key?.values?.nodes ?? [],
        loading,
    }
}

export const AddMetadataForm: FC<{ onDidAdd: () => void; repoID: string }> = ({ onDidAdd, repoID }) => {
    const [key, setKey] = useState<string>('')
    const [value, setValue] = useState<string>('')

    const { suggestions: keys, loading: keysLoading } = useKeySuggestions(key)
    const { suggestions: values, loading: valuesLoading } = useValueSuggestions(key, value)

    const [addMeta, { called: addCalled, loading: addLoading, error: addError }] = useMutation<
        AddRepoMetadataResult,
        AddRepoMetadataVariables
    >(ADD_REPO_METADATA_GQL)

    const onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        async function add(): Promise<void> {
            await addMeta({
                variables: {
                    key,
                    value: value || null,
                    repo: repoID,
                },
            }).then(() => onDidAdd())

            setKey('')
            setValue('')
        }
        // eslint-disable-next-line no-console
        add().catch(console.error)
    }

    return (
        <>
            {!addLoading && !addError && addCalled && (
                <Alert className="flex-grow-1 mt-3 mb-3" variant="success">
                    Metadata added
                </Alert>
            )}

            <Container className="repo-metadata-page" as="section">
                <H2>Add metadata</H2>
                <Text>Add an additional key, or key-value pair, to this repository.</Text>
                <Form onSubmit={onSubmit}>
                    {!addLoading && addError && <ErrorAlert className="flex-grow-1 m-0 mb-3" error={addError} />}
                    <div className="d-flex justify-content-between align-items-center">
                        <div className="form-group flex-grow-1 mb-0 mr-4">
                            <Label htmlFor="metadata-key">Key</Label>
                            <Combobox openOnFocus={true} onSelect={setKey}>
                                <ComboboxInput
                                    id="metadata-key"
                                    required={true}
                                    disabled={addLoading}
                                    autoComplete="off"
                                    value={key}
                                    onChange={event => setKey(event.target.value)}
                                    message="e.g. 'status', 'license', 'language'"
                                />
                                <ComboboxPopover>
                                    {!keysLoading && (
                                        <ComboboxList>
                                            {keys.map(suggestion => (
                                                <ComboboxOption key={suggestion} value={suggestion} />
                                            ))}
                                        </ComboboxList>
                                    )}
                                </ComboboxPopover>
                            </Combobox>
                        </div>
                        <div className="form-group flex-grow-1 mb-0 mr-4">
                            <Label htmlFor="metadata-value">Value (optional)</Label>
                            <Combobox openOnFocus={true} onSelect={setValue}>
                                <ComboboxInput
                                    id="metadata-value"
                                    disabled={addLoading}
                                    autoComplete="off"
                                    value={value}
                                    onChange={event => setValue(event.target.value)}
                                    message="e.g. 'deprecated', 'MIT', 'Go'"
                                />
                                <ComboboxPopover>
                                    {!valuesLoading && (
                                        <ComboboxList>
                                            {values.map(suggestion => (
                                                <ComboboxOption key={suggestion} value={suggestion} />
                                            ))}
                                        </ComboboxList>
                                    )}
                                </ComboboxPopover>
                            </Combobox>
                        </div>
                        <div className="d-flex justify-content-end mt-1">
                            <LoaderButton variant="primary" type="submit" loading={addLoading} label="Add" />
                        </div>
                    </div>
                </Form>
            </Container>
        </>
    )
}
