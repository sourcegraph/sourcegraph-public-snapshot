import { FC } from 'react'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { Button, Link, Tooltip } from '@sourcegraph/wildcard'

import { CodeHostAppLimit } from '../../../CodeHostExternalServiceAlert'
import { CodeHostIcon, getCodeHostName, getCodeHostURLParam } from '../../helpers'

import styles from './CodeHostsPicker.module.scss'

const SUPPORTED_CODE_HOSTS = [
    ExternalServiceKind.GITHUB,
    ExternalServiceKind.GITLAB,
    ExternalServiceKind.GERRIT,
    ExternalServiceKind.BITBUCKETCLOUD,
    ExternalServiceKind.BITBUCKETSERVER,
    ExternalServiceKind.AWSCODECOMMIT,
    ExternalServiceKind.GITOLITE,
    ExternalServiceKind.AZUREDEVOPS,
]

interface CodeHostsPickerProps {
    /**
     * Turns on/off code host picker buttons, originally it's used to disable
     * code host connection UI buttons when user already has 1 remote code host
     * in Sourcegraph App mode.
     */
    isLimitReached: boolean
}

export const CodeHostsPicker: FC<CodeHostsPickerProps> = props => (
    <section>
        <header className={styles.header}>
            <span>Add another remote code host</span>
            <small className="text-muted">Choose a provider from the list below</small>
        </header>

        <CodeHostAppLimit className="mb-2" />

        <ul className={styles.list}>
            {SUPPORTED_CODE_HOSTS.map(codeHostType => (
                <li key={codeHostType}>
                    <Tooltip
                        content={props.isLimitReached ? 'You have reached remote code host limit' : ''}
                        placement="left"
                    >
                        <Button
                            as={Link}
                            to={`/setup/remote-repositories/${getCodeHostURLParam(codeHostType)}/create`}
                            variant="secondary"
                            outline={true}
                            disabled={props.isLimitReached}
                            className={styles.item}
                        >
                            <CodeHostIcon codeHostType={codeHostType} aria-hidden={true} />
                            <span>{getCodeHostName(codeHostType)}</span>
                        </Button>
                    </Tooltip>
                </li>
            ))}
        </ul>
    </section>
)
