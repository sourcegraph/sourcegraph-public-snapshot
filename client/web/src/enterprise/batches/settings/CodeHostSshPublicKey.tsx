import React, { useCallback, useState } from 'react'

import { mdiContentCopy } from '@mdi/js'
import copy from 'copy-to-clipboard'
import { noop } from 'lodash'

import { Button, TextArea, Link, Icon, Label, Text } from '@sourcegraph/wildcard'

import { ExternalServiceKind } from '../../../graphql-operations'

const configInstructionLinks: Record<ExternalServiceKind, string> = {
    [ExternalServiceKind.GITHUB]:
        'https://docs.github.com/en/github/authenticating-to-github/adding-a-new-ssh-key-to-your-github-account',
    [ExternalServiceKind.GITLAB]: 'https://docs.gitlab.com/ee/ssh/#add-an-ssh-key-to-your-gitlab-account',
    [ExternalServiceKind.BITBUCKETSERVER]:
        'https://confluence.atlassian.com/bitbucketserver/ssh-user-keys-for-personal-use-776639793.html',
    [ExternalServiceKind.AWSCODECOMMIT]: 'unsupported',
    [ExternalServiceKind.AZUREDEVOPS]: 'unsupported',
    [ExternalServiceKind.BITBUCKETCLOUD]: 'unsupported',
    [ExternalServiceKind.GERRIT]: 'unsupported',
    [ExternalServiceKind.GITOLITE]: 'unsupported',
    [ExternalServiceKind.GOMODULES]: 'unsupported',
    [ExternalServiceKind.JVMPACKAGES]: 'unsupported',
    [ExternalServiceKind.NPMPACKAGES]: 'unsupported',
    [ExternalServiceKind.OTHER]: 'unsupported',
    [ExternalServiceKind.PERFORCE]: 'unsupported',
    [ExternalServiceKind.PAGURE]: 'unsupported',
    [ExternalServiceKind.PHABRICATOR]: 'unsupported',
    [ExternalServiceKind.PYTHONPACKAGES]: 'unsupported',
    [ExternalServiceKind.RUSTPACKAGES]: 'unsupported',
    [ExternalServiceKind.RUBYPACKAGES]: 'unsupported',
}

export interface CodeHostSshPublicKeyProps {
    externalServiceKind: ExternalServiceKind
    sshPublicKey: string
    label?: string
    showInstructionsLink?: boolean
    showCopyButton?: boolean
}

export const CodeHostSshPublicKey: React.FunctionComponent<React.PropsWithChildren<CodeHostSshPublicKeyProps>> = ({
    externalServiceKind,
    sshPublicKey,
    showInstructionsLink = true,
    showCopyButton = true,
    label = 'Public SSH key',
}) => {
    const [copied, setCopied] = useState<boolean>(false)
    const onCopy = useCallback(() => {
        copy(sshPublicKey)
        setCopied(true)
    }, [sshPublicKey])
    return (
        <>
            <div className="d-flex justify-content-between align-items-end mb-2">
                <Label htmlFor={LABEL_ID}>{label}</Label>
                {showCopyButton && (
                    <Button onClick={onCopy} variant="secondary">
                        <Icon aria-hidden={true} svgPath={mdiContentCopy} />
                        {copied ? 'Copied!' : 'Copy'}
                    </Button>
                )}
            </div>
            <TextArea
                id={LABEL_ID}
                className="text-monospace mb-3"
                rows={5}
                spellCheck="false"
                value={sshPublicKey}
                onChange={noop}
            />
            {showInstructionsLink && (
                <Text>
                    <Link to={configInstructionLinks[externalServiceKind]} target="_blank" rel="noopener">
                        Configuration instructions
                    </Link>
                </Text>
            )}
        </>
    )
}

const LABEL_ID = 'code-host-ssh-public-key-textarea'
