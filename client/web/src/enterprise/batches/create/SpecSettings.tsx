import { ApolloError } from '@apollo/client'
import React, { useState } from 'react'

import { SettingsOrgSubject, SettingsUserSubject } from '@sourcegraph/shared/src/settings/settings'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { Button, Container, Input } from '@sourcegraph/wildcard'

import { Scalars } from '../../../graphql-operations'

import { NamespaceSelector } from './NamespaceSelector'
import styles from './SpecSettings.module.scss'

interface SpecSettingsProps {
    /** The set of all namespaces available for a user to create batch changes in. */
    namespaces: (SettingsOrgSubject | SettingsUserSubject)[]
    /** The current name of the batch change, if it has been given one. */
    currentName?: string
    /** The namespace which should currently be selected from the dropdown. */
    selectedNamespace: SettingsOrgSubject | SettingsUserSubject
    /** Handler to set a new selected namespace. */
    setSelectedNamespace: (namespace: SettingsOrgSubject | SettingsUserSubject) => void
    /** Callback to cancel changes. */
    onCancel: () => void
    /** Callback to confirm changes and close settings. */
    onConfirm: (name: string, namespace: Scalars['ID']) => void
    confirmButtonText: string
    confirmLoading: boolean
    confirmError?: ApolloError
}

export const SpecSettings: React.FunctionComponent<SpecSettingsProps> = ({
    namespaces,
    currentName = '',
    selectedNamespace,
    setSelectedNamespace,
    onCancel,
    onConfirm,
    confirmLoading,
    confirmError,
    confirmButtonText,
}) => {
    const [nameInput, setNameInput] = useState(currentName)

    return (
        <>
            <h4>Batch spec settings</h4>
            <Container>
                {confirmError && <ErrorAlert error={confirmError} />}
                <NamespaceSelector
                    namespaces={namespaces}
                    selectedNamespace={selectedNamespace.id}
                    onSelect={setSelectedNamespace}
                />
                <Input
                    className={styles.nameInput}
                    label="Batch change name"
                    value={nameInput}
                    onChange={event => setNameInput(event.target.value)}
                    onKeyPress={event => {
                        if (event.key === 'Enter') {
                            onConfirm(nameInput, selectedNamespace.id)
                        }
                    }}
                />
            </Container>
            <div className="mt-3 align-self-end">
                <Button variant="secondary" outline={true} className="mr-2" onClick={onCancel}>
                    Cancel
                </Button>
                <Button
                    variant="primary"
                    onClick={() => onConfirm(nameInput, selectedNamespace.id)}
                    disabled={confirmLoading}
                >
                    {confirmButtonText}
                </Button>
            </div>
        </>
    )
}
