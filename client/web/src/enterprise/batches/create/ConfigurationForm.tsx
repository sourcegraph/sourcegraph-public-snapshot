import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import LockIcon from 'mdi-react/LockIcon'
import { useHistory } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { useMutation } from '@sourcegraph/http-client'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import {
    SettingsCascadeProps,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { Button, Container, Input, Icon, RadioButton, Typography } from '@sourcegraph/wildcard'

import {
    BatchChangeFields,
    CreateBatchSpecFromRawResult,
    CreateBatchSpecFromRawVariables,
    CreateEmptyBatchChangeResult,
    CreateEmptyBatchChangeVariables,
    Scalars,
} from '../../../graphql-operations'

import { CREATE_BATCH_SPEC_FROM_RAW, CREATE_EMPTY_BATCH_CHANGE } from './backend'
import { NamespaceSelector } from './NamespaceSelector'
import { useNamespaces } from './useNamespaces'

import styles from './ConfigurationForm.module.scss'

/* Regex pattern for a valid batch change name. Needs to match what's defined in the BatchSpec JSON schema. */
const NAME_PATTERN = /^[\w.-]+$/

interface ConfigurationFormProps extends SettingsCascadeProps<Settings> {
    /**
     * Whether or not to display the configuration form in read-only mode, i.e. to view
     * for an existing batch change.
     */
    isReadOnly?: boolean
    /** The existing batch change to use to pre-populate the form. */
    batchChange?: Pick<BatchChangeFields, 'name' | 'namespace'>

    /**
     * When set, apply a template to the batch spec before redirecting to the edit page.
     */
    renderTemplate?: (title: string) => string
    /** The title of the insight this was created from, if any. */
    insightTitle?: string
    /**
     * When set, will pre-select the namespace with the given ID from the dropdown
     * selector, if an existing batch change is not available.
     */
    initialNamespaceID?: Scalars['ID']
}

export const ConfigurationForm: React.FunctionComponent<React.PropsWithChildren<ConfigurationFormProps>> = ({
    settingsCascade,
    isReadOnly,
    batchChange,
    renderTemplate,
    insightTitle,
    initialNamespaceID,
}) => {
    const [createEmptyBatchChange, { loading: batchChangeLoading, error: batchChangeError }] = useMutation<
        CreateEmptyBatchChangeResult,
        CreateEmptyBatchChangeVariables
    >(CREATE_EMPTY_BATCH_CHANGE)
    const [createBatchSpecFromRaw, { loading: batchSpecLoading, error: batchSpecError }] = useMutation<
        CreateBatchSpecFromRawResult,
        CreateBatchSpecFromRawVariables
    >(CREATE_BATCH_SPEC_FROM_RAW)

    const loading = batchChangeLoading || batchSpecLoading
    const error = batchChangeError || batchSpecError

    const { namespaces, defaultSelectedNamespace } = useNamespaces(
        settingsCascade,
        batchChange?.namespace.id || initialNamespaceID
    )

    // The namespace selected for creating the new batch change under.
    const [selectedNamespace, setSelectedNamespace] = useState<SettingsUserSubject | SettingsOrgSubject>(
        defaultSelectedNamespace
    )

    const [nameInput, setNameInput] = useState(batchChange?.name || '')
    const [isNameValid, setIsNameValid] = useState<boolean>()

    const onNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setNameInput(event.target.value)
        setIsNameValid(NAME_PATTERN.test(event.target.value))
    }, [])

    const history = useHistory()
    const handleCancel = (): void => history.goBack()
    const handleCreate = (): void => {
        const redirectSearchParameters = new URLSearchParams()
        if (insightTitle) {
            redirectSearchParameters.set('title', insightTitle)
        }
        let serializedRedirectSearchParameters = redirectSearchParameters.toString()
        if (serializedRedirectSearchParameters.length > 0) {
            serializedRedirectSearchParameters = '?' + serializedRedirectSearchParameters
        }
        createEmptyBatchChange({
            variables: { namespace: selectedNamespace.id, name: nameInput },
        })
            .then(args => {
                if (!renderTemplate) {
                    return Promise.resolve(args)
                }

                const template = renderTemplate(nameInput)

                return args.data?.createEmptyBatchChange.id && template
                    ? createBatchSpecFromRaw({
                          variables: { namespace: selectedNamespace.id, spec: template, noCache: false },
                      }).then(() => Promise.resolve(args))
                    : Promise.resolve(args)
            })
            .then(({ data }) =>
                data
                    ? history.push(`${data.createEmptyBatchChange.url}/edit${serializedRedirectSearchParameters}`)
                    : noop()
            )
            // We destructure and surface the error from `useMutation` instead.
            .catch(noop)
    }

    return (
        <Form className={styles.form} onSubmit={handleCreate}>
            <Container className="mb-4">
                {error && <ErrorAlert error={error} />}
                <NamespaceSelector
                    namespaces={namespaces}
                    selectedNamespace={selectedNamespace.id}
                    onSelect={setSelectedNamespace}
                    disabled={isReadOnly}
                />
                <Input
                    label="Batch change name"
                    value={nameInput}
                    onChange={onNameChange}
                    pattern={String(NAME_PATTERN)}
                    required={true}
                    status={isNameValid === undefined ? undefined : isNameValid ? 'valid' : 'error'}
                    placeholder="My batch change name"
                    disabled={isReadOnly}
                />
                {!isReadOnly && (
                    <small className="text-muted">
                        Give it a short, descriptive name to reference the batch change on Sourcegraph. Do not include
                        confidential information.{' '}
                        <span className={classNames(isNameValid === false && 'text-danger')}>
                            Only letters, numbers, _, and - are allowed.
                        </span>
                    </small>
                )}
                <hr className="my-3" />
                <Typography.H3 className="text-muted">
                    Visibility <Icon data-tooltip="Coming soon" as={InfoCircleOutlineIcon} />
                </Typography.H3>
                <div className="form-group mb-1">
                    <RadioButton
                        name="visibility"
                        value="public"
                        className="mr-2"
                        checked={true}
                        disabled={true}
                        label="Public"
                        aria-label="Public"
                    />
                </div>
                <div className="form-group mb-0">
                    <RadioButton
                        name="visibility"
                        value="private"
                        className="mr-2 mb-0"
                        disabled={true}
                        label={
                            <>
                                Private <Icon className="text-warning" aria-hidden={true} as={LockIcon} />
                            </>
                        }
                        aria-label="Private"
                    />
                </div>
            </Container>

            {!isReadOnly && (
                <div className={styles.ctaGroup}>
                    <Button variant="secondary" type="button" outline={true} onClick={handleCancel}>
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        type="submit"
                        onClick={handleCreate}
                        disabled={loading || nameInput === '' || !isNameValid}
                    >
                        Create
                    </Button>
                </div>
            )}
        </Form>
    )
}
