import React from 'react'

import { mdiCheckboxBlankCircleOutline, mdiCheckCircleOutline } from '@mdi/js'
import classNames from 'classnames'

import { H3, Icon, Text } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import type { BatchChangesCodeHostFields } from '../../../graphql-operations'

import { GitHubAppControls } from './GitHubAppControls'

import styles from './CommitSigningIntegrationNode.module.scss'

interface CommitSigningIntegrationNodeProps {
    readOnly: boolean
    node: BatchChangesCodeHostFields
    refetch: () => void
}

export const CommitSigningIntegrationNode: React.FunctionComponent<
    React.PropsWithChildren<CommitSigningIntegrationNodeProps>
> = ({ node, readOnly, refetch }) => {
    const ExternalServiceIcon = defaultExternalServices[node.externalServiceKind].icon
    return (
        <li className={classNames(styles.node, 'list-group-item')}>
            <div
                className={classNames(
                    styles.wrapper,
                    'd-flex justify-content-between align-items-center flex-wrap mb-0'
                )}
            >
                <H3 className="mb-0 mr-2">
                    {node.commitSigningConfiguration ? (
                        <Icon
                            aria-label="This code host has commit signing enabled with a GitHub App."
                            className="text-success"
                            svgPath={mdiCheckCircleOutline}
                        />
                    ) : (
                        <Icon
                            aria-label="This code host does not have a GitHub App connected for commit signing."
                            className="text-danger"
                            svgPath={mdiCheckboxBlankCircleOutline}
                        />
                    )}

                    <Icon className="mx-2" aria-hidden={true} as={ExternalServiceIcon} />
                    {node.externalServiceURL}
                </H3>
                {readOnly ? (
                    <ReadOnlyAppDetails config={node.commitSigningConfiguration} />
                ) : (
                    <GitHubAppControls
                        baseURL={node.externalServiceURL}
                        config={node.commitSigningConfiguration}
                        refetch={refetch}
                    />
                )}
            </div>
        </li>
    )
}

interface ReadOnlyAppDetailsProps {
    config: BatchChangesCodeHostFields['commitSigningConfiguration']
}

const ReadOnlyAppDetails: React.FunctionComponent<ReadOnlyAppDetailsProps> = ({ config }) =>
    config ? (
        <div className={styles.readonlyAppDetails}>
            <img className={styles.appLogo} src={config.logo} alt="app logo" aria-hidden={true} />
            <Text size="small" className="font-weight-bold m-0">
                {config.name}
            </Text>
        </div>
    ) : (
        <div className={styles.readonlyAppDetails}>
            <Text size="small" className="m-0">
                No App configured
            </Text>
        </div>
    )
