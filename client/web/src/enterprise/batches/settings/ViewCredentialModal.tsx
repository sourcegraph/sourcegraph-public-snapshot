import React from 'react'

import { Button, Checkbox, Modal, H4, Input } from '@sourcegraph/wildcard'

import {
    BatchChangesCodeHostFields,
    BatchChangesCredentialFields,
    ExternalServiceKind,
} from '../../../graphql-operations'

import { CodeHostSshPublicKey } from './CodeHostSshPublicKey'
import { ModalHeader } from './ModalHeader'

interface ViewCredentialModalProps {
    codeHost: BatchChangesCodeHostFields
    credential: BatchChangesCredentialFields
    supportsSignedCommits: boolean

    onClose: () => void
}

export const ViewCredentialModal: React.FunctionComponent<React.PropsWithChildren<ViewCredentialModalProps>> = ({
    codeHost,
    credential,
    supportsSignedCommits,
    onClose,
}) => {
    const labelId = 'viewCredential'
    return (
        <Modal onDismiss={onClose} aria-labelledby={labelId}>
            <ModalHeader
                id={labelId}
                externalServiceKind={codeHost.externalServiceKind}
                externalServiceURL={codeHost.externalServiceURL}
            />

            <H4>Personal access token</H4>
            {supportsSignedCommits && (
                <Input className="form-group" value="PATs cannot be viewed after entering." disabled={true} />
            )}

            {(codeHost.externalServiceKind === ExternalServiceKind.GITHUB ||
                codeHost.externalServiceKind === ExternalServiceKind.GITLAB) && (
                <Checkbox
                    name="enable-sign-commits"
                    id="enable-sign-commits"
                    checked={credential.commitSigningOptedIn}
                    disabled={true}
                    label="Sign commits on this code host"
                    message={`This property cannot be modified after creation. To ${
                        credential.commitSigningOptedIn ? 'disable' : 'enable'
                    } commit signing, remove this credential and create a new one.`}
                />
            )}

            <hr className="my-3" />

            <CodeHostSshPublicKey
                externalServiceKind={codeHost.externalServiceKind}
                sshPublicKey={credential.sshPublicKey!}
            />

            <div className="d-flex justify-content-end">
                <Button onClick={onClose} outline={true} variant="secondary">
                    Close
                </Button>
            </div>
        </Modal>
    )
}
