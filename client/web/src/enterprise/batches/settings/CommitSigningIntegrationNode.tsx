import React, { useState } from 'react'

import { mdiCheckCircleOutline, mdiCheckboxBlankCircleOutline, mdiCogOutline, mdiDelete, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { AnchorLink, Button, ButtonLink, H3, Icon, Link, Text } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { AppLogo } from '../../../components/gitHubApps/AppLogo'
import { RemoveGitHubAppModal } from '../../../components/gitHubApps/RemoveGitHubAppModal'
import type { BatchChangesCodeHostFields } from '../../../graphql-operations'

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
                    <AppDetailsControls
                        baseURL={node.externalServiceURL}
                        config={node.commitSigningConfiguration}
                        refetch={refetch}
                    />
                )}
            </div>
        </li>
    )
}

interface AppDetailsControlsProps {
    baseURL: string
    config: BatchChangesCodeHostFields['commitSigningConfiguration']
    refetch: () => void
}

const AppDetailsControls: React.FunctionComponent<AppDetailsControlsProps> = ({ baseURL, config, refetch }) => {
    const [removeModalOpen, setRemoveModalOpen] = useState<boolean>(false)

    const createURL = `/site-admin/batch-changes/github-apps/new?baseURL=${encodeURIComponent(baseURL)}`
    return config ? (
        <>
            {removeModalOpen && (
                <RemoveGitHubAppModal onCancel={() => setRemoveModalOpen(false)} afterDelete={refetch} app={config} />
            )}
            <div className="d-flex align-items-center">
                <AppLogo src={config.logo} name={config.name} className={classNames(styles.appLogoLarge, 'mr-2')} />

                <div className={styles.appDetailsColumn}>
                    <Text size="small" className="font-weight-bold mb-0">
                        {config.name}
                    </Text>
                    <Text size="small" className="text-muted mb-0">
                        AppID: {config.appID}
                    </Text>
                </div>
            </div>
            <div className="ml-auto">
                <AnchorLink to={config.appURL} target="_blank" className="mr-3">
                    <small>
                        View In GitHub <Icon inline={true} svgPath={mdiOpenInNew} aria-hidden={true} />
                    </small>
                </AnchorLink>
                <ButtonLink
                    className="mr-2"
                    aria-label="Edit"
                    to={`github-apps/${config.id}`}
                    variant="secondary"
                    size="sm"
                >
                    <Icon aria-hidden={true} svgPath={mdiCogOutline} /> Edit
                </ButtonLink>
                <Button
                    aria-label="Remove GitHub App"
                    onClick={() => setRemoveModalOpen(true)}
                    variant="danger"
                    size="sm"
                >
                    <Icon aria-hidden={true} svgPath={mdiDelete} /> Remove
                </Button>
            </div>
        </>
    ) : (
        <ButtonLink to={createURL} className="ml-auto text-nowrap" variant="success" as={Link} size="sm">
            Create GitHub App
        </ButtonLink>
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
