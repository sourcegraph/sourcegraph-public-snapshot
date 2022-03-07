import classNames from 'classnames'
import React, { useCallback, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { catchError, map } from 'rxjs/operators'

import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, PageHeader, Container } from '@sourcegraph/wildcard'

import { Collapsible } from '../components/Collapsible'
import { FilteredConnection, FilteredConnectionFilter } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { FeatureFlagFields } from '../graphql-operations'

import { fetchFeatureFlags as defaultFetchFeatureFlags } from './backend'
import styles from './SiteAdminFeatureFlagsPage.module.scss'

export interface SiteAdminFeatureFlagsPageProps extends RouteComponentProps<{}>, TelemetryProps {
    fetchFeatureFlags?: typeof defaultFetchFeatureFlags
}

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'Type',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all feature flags',
                args: {},
            },
            {
                label: 'Boolean',
                value: 'boolean',
                tooltip: 'Show boolean feature flags',
                args: { type: 'FeatureFlagBoolean' },
            },
            {
                label: 'Rollout',
                value: 'rollout',
                tooltip: 'Show rollout feature flags',
                args: { type: 'FeatureFlagRollout' },
            },
        ],
    },
]

export const SiteAdminFeatureFlagsPage: React.FunctionComponent<SiteAdminFeatureFlagsPageProps> = ({
    fetchFeatureFlags = defaultFetchFeatureFlags,
    telemetryService,
    ...props
}) => {
    const featureFlagsOrErrorsObservable = useMemo(
        () => fetchFeatureFlags().pipe(catchError((error): [ErrorLike] => [asError(error)])),
        [fetchFeatureFlags]
    )

    const queryFeatureFlags = useCallback(
        (args: { query?: string; type?: string }) =>
            featureFlagsOrErrorsObservable.pipe(
                map(featureFlagsOrErrors => {
                    if (isErrorLike(featureFlagsOrErrors)) {
                        return { nodes: [] }
                    }
                    return {
                        nodes: featureFlagsOrErrors.filter(
                            node =>
                                (args.type === undefined || node.__typename === args.type) &&
                                (!args.query || node.name.toLowerCase().includes(args.query.toLowerCase()))
                        ),
                        totalCount: featureFlagsOrErrors.length,
                        pageInfo: { hasNextPage: false },
                    }
                })
            ),
        [featureFlagsOrErrorsObservable]
    )

    return (
        <>
            <PageTitle title="Feature flags - Admin" />
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Feature flags</>,
                    },
                ]}
                description={
                    <>
                        <p>
                            Feature flags, as opposed to experimental features, are intended to be strictly short-lived.
                            They are designed to be useful for A/B testing, and the values of all active feature flags
                            are added to every event log for the purpose of analytics.
                        </p>
                        <p>
                            To learn more, refer to{' '}
                            <Link target="_blank" rel="noopener noreferrer" to="/help/dev/how-to/use_feature_flags">
                                How to use feature flags
                            </Link>
                            .
                        </p>
                    </>
                }
                className="mb-3"
            />

            <Container>
                <FilteredConnection<FeatureFlagFields, {}>
                    listComponent="div"
                    listClassName={classNames('mb-3', styles.flagsGrid)}
                    noun="feature flag"
                    pluralNoun="feature flags"
                    queryConnection={queryFeatureFlags}
                    nodeComponent={FeatureFlagNode}
                    history={props.history}
                    location={props.location}
                    filters={filters}
                />
            </Container>
        </>
    )
}

interface FeatureFlagNodeProps {
    node: FeatureFlagFields
}

const FeatureFlagNode: React.FunctionComponent<FeatureFlagNodeProps> = ({ node }) => (
    <React.Fragment key={node.name}>
        <span className={styles.separator} />

        <div className={classNames('d-flex flex-column', styles.information)}>
            <div>
                <h3>{node.name}</h3>

                <p className="m-0">
                    <span className="text-muted">{node.__typename}</span>
                </p>
            </div>
        </div>

        <span className={classNames('d-none d-md-inline', styles.progress)}>
            <div className="m-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
                <div>
                    {node.__typename === 'FeatureFlagBoolean' && <code>{JSON.stringify(node.value)}</code>}
                    {node.__typename === 'FeatureFlagRollout' && node.rolloutBasisPoints}
                </div>

                {node.__typename === 'FeatureFlagRollout' && (
                    <div>
                        <meter
                            min={0}
                            max={1}
                            optimum={1}
                            value={node.rolloutBasisPoints / (100 * 100)}
                            data-tooltip={`${Math.floor(node.rolloutBasisPoints / 100)}%`}
                            aria-label="rollout progress"
                            data-placement="bottom"
                        />
                    </div>
                )}
            </div>
        </span>

        {/*
            TODO: move into individual feature flag page as part of
            https://github.com/sourcegraph/sourcegraph/issues/32232
        */}
        {node.overrides.length > 0 && (
            <Collapsible
                title={
                    <strong>
                        {node.overrides.length} {node.overrides.length > 1 ? 'overrides' : 'override'}
                    </strong>
                }
                className="p-0 font-weight-normal"
                titleClassName="flex-grow-1"
                buttonClassName="mb-0"
                defaultExpanded={false}
            >
                <div className={classNames('pt-2', styles.nodeGrid)}>
                    {node.overrides.map(override => (
                        <React.Fragment key={override.id}>
                            <div className="py-1 pr-2">
                                <code>{JSON.stringify(override.value)}</code>
                            </div>

                            <span className={classNames('py-1 pl-2', styles.nodeGridCode)}>
                                {/*
                                        TODO: querying for namespace connection seems to
                                        error out often, so just present the ID for now.
                                        https://github.com/sourcegraph/sourcegraph/issues/32238
                                    */}
                                <code>{override.id}</code>
                            </span>
                        </React.Fragment>
                    ))}
                </div>
            </Collapsible>
        )}
    </React.Fragment>
)
