import React from 'react'

import { mdiCheckCircleOutline, mdiCheckboxBlankCircleOutline, mdiDelete, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'
import { noop } from 'lodash'

import { AnchorLink, Button, ButtonLink, H3, Icon, Link, Text, Tooltip } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { BatchChangesCodeHostFields } from '../../../graphql-operations'

import styles from './CommitSigningIntegrationNode.module.scss'

interface CommitSigningIntegrationNodeProps {
    readOnly: boolean
    node: BatchChangesCodeHostFields
}

export const CommitSigningIntegrationNode: React.FunctionComponent<
    React.PropsWithChildren<CommitSigningIntegrationNodeProps>
> = ({ node, readOnly }) => {
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
                    <AppDetailsControls baseURL={node.externalServiceURL} config={node.commitSigningConfiguration} />
                )}
            </div>
        </li>
    )
}

interface AppDetailsControlsProps {
    baseURL: string
    config: BatchChangesCodeHostFields['commitSigningConfiguration']
}

const AppDetailsControls: React.FunctionComponent<AppDetailsControlsProps> = ({ baseURL, config }) => {
    const createURL = `/site-admin/batch-changes/new-github-app?baseURL=${encodeURIComponent(baseURL)}`
    return config ? (
        <>
            <div className="d-flex align-items-center">
                <img className={styles.appLogoLarge} src={config.logo} alt="app logo" aria-hidden={true} />
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
                {/* TODO: Hook up delete button */}
                <Tooltip content="Remove GitHub App">
                    <Button aria-label="Remove" onClick={noop} disabled={false} variant="danger" size="sm">
                        <Icon aria-hidden={true} svgPath={mdiDelete} /> Remove
                    </Button>
                </Tooltip>
            </div>
        </>
    ) : (
        // TODO: Hook up create button
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
