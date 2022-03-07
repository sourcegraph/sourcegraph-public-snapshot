import classNames from 'classnames'
import React, { FunctionComponent, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { of } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Container, Link, LoadingSpinner, PageHeader, useObservable } from '@sourcegraph/wildcard'

import { Collapsible } from '../components/Collapsible'
import { FeatureFlagFields } from '../graphql-operations'

import { fetchFeatureFlags as defaultFetchFeatureFlags } from './backend'
import styles from './SiteAdminFeatureFlagConfigurationPage.module.scss'
import { getFeatureFlagReferences, parseProductReference, Reference } from './SiteAdminFeatureFlagsPage'

export interface SiteAdminFeatureFlagConfigurationProps extends RouteComponentProps<{ name: string }>, TelemetryProps {
    fetchFeatureFlags?: typeof defaultFetchFeatureFlags
    productVersion?: string
}

export const SiteAdminFeatureFlagConfigurationPage: FunctionComponent<SiteAdminFeatureFlagConfigurationProps> = ({
    match: {
        params: { name },
    },
    fetchFeatureFlags = defaultFetchFeatureFlags,
    productVersion = window.context.version,
}) => {
    const productGitVersion = parseProductReference(productVersion)

    // Fetch feature flags
    const featureFlagOrError = useObservable(
        useMemo(
            () =>
                name
                    ? fetchFeatureFlags().pipe(
                          map(flags => flags.find(flag => flag.name === name)),
                          catchError((error): [ErrorLike] => [asError(error)])
                      )
                    : of(undefined),
            [name, fetchFeatureFlags]
        )
    )

    const references = useObservable(
        useMemo(
            () =>
                featureFlagOrError?.name
                    ? getFeatureFlagReferences(featureFlagOrError.name, productGitVersion)
                    : of([]),
            [featureFlagOrError, productGitVersion]
        )
    )

    if (name && featureFlagOrError === undefined) {
        return <LoadingSpinner className="mt-2" />
    }
    if (isErrorLike(featureFlagOrError)) {
        return <ErrorAlert prefix="Error fetching feature flag policy" error={featureFlagOrError} />
    }

    const verb = name === '' ? 'Create' : 'Manage'
    return (
        <>
            <PageTitle title={`${verb} feature flag`} />
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>{verb} feature flag</>,
                    },
                ]}
                className="mb-3"
            />

            {featureFlagOrError?.name && <ManageFeatureFlag flag={featureFlagOrError} references={references} />}
        </>
    )
}

const ManageFeatureFlag: FunctionComponent<{ flag: FeatureFlagFields; references?: Reference[] }> = ({
    flag,
    references,
}) => (
    <Container>
        <h3>Name</h3>
        <p>{flag.name}</p>

        <h3>Type</h3>
        <p>{flag.__typename.slice('FeatureFlag'.length)}</p>

        <h3>Value</h3>
        <p>
            {flag.__typename === 'FeatureFlagBoolean' && <code>{JSON.stringify(flag.value)}</code>}
            {flag.__typename === 'FeatureFlagRollout' && flag.rolloutBasisPoints}
        </p>

        <Collapsible
            title={<h3>Overrides</h3>}
            detail={`${flag.overrides.length} ${flag.overrides.length !== 1 ? 'overrides' : 'override'}`}
            className="p-0 font-weight-normal"
            buttonClassName="mb-0"
            titleAtStart={true}
            defaultExpanded={false}
        >
            <div className={classNames('pt-2', styles.nodeGrid)}>
                {flag.overrides.map(override => (
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

        {references && references.length > 0 && (
            <>
                <br />
                <Collapsible
                    title={<h3>References</h3>}
                    detail={`${references.length} ${references.length > 1 ? 'references' : 'reference'}`}
                    className="p-0 font-weight-normal"
                    buttonClassName="mb-0"
                    titleAtStart={true}
                    defaultExpanded={false}
                >
                    <div className="pt-2">
                        {references.map(reference => (
                            <div key={flag.name + reference.file}>
                                <Link target="_blank" rel="noopener noreferrer" to={reference.searchURL}>
                                    <code>{reference.file}</code>
                                </Link>
                            </div>
                        ))}
                    </div>
                </Collapsible>
            </>
        )}
    </Container>
)
