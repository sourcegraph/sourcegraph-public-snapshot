import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import * as H from 'history'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps, useHistory } from 'react-router'
import { Subject } from 'rxjs'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { GitObjectType } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '@sourcegraph/web/src/components/FilteredConnection'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { CodeIntelligenceConfigurationPolicyFields } from '../../../../graphql-operations'
import { CodeIntelConfigurationPageHeader } from '../components/CodeIntelConfigurationPageHeader'
import { EmptyPoliciesList } from '../components/EmptyPoliciesList'
import { FlashMessage } from '../components/FlashMessage'
import { IndexingPolicyDescription } from '../components/IndexingPolicyDescription'
import { PolicyListActions } from '../components/PolicyListActions'
import { RetentionPolicyDescription } from '../components/RetentionPolicyDescription'
import { queryPolicies as defaultQueryPolicies } from '../hooks/queryPolicies'

import styles from './CodeIntelConfigurationPage.module.scss'

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'For',
        type: 'select',
        values: [
            {
                label: 'Anything',
                value: 'anything',
                tooltip: 'Anything',
                args: {},
            },
            {
                label: 'Data retention',
                value: 'data-retention',
                tooltip: 'Data retention',
                args: { forDataRetention: true },
            },
            {
                label: 'Indexing',
                value: 'indexing',
                tooltip: 'Indexing',
                args: { forIndexing: true },
            },
        ],
    },
]

export interface CodeIntelConfigurationPageProps extends RouteComponentProps<{}>, ThemeProps, TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    queryPolicies?: typeof defaultQueryPolicies
    repo?: { id: string }
    indexingEnabled?: boolean
    isLightTheme: boolean
    telemetryService: TelemetryService
}

export const CodeIntelConfigurationPage: FunctionComponent<CodeIntelConfigurationPageProps> = ({
    authenticatedUser,
    queryPolicies = defaultQueryPolicies,
    repo,
    indexingEnabled = window.context?.codeIntelAutoIndexingEnabled,
    telemetryService,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelConfigurationPage'), [telemetryService])

    const history = useHistory()

    const apolloClient = useApolloClient()
    const queryPoliciesCallback = useCallback(
        (args: FilteredConnectionQueryArguments) => queryPolicies({ ...args, repository: repo?.id }, apolloClient),
        [queryPolicies, repo?.id, apolloClient]
    )

    const querySubject = useMemo(() => new Subject<string>(), [])

    return (
        <>
            <PageTitle title="Precise code intelligence configuration" />
            <CodeIntelConfigurationPageHeader>
                <PageHeader
                    headingElement="h2"
                    path={[
                        {
                            text: <>Precise code intelligence configuration</>,
                        },
                    ]}
                    description={`Rules that control data retention${
                        indexingEnabled ? ' and auto-indexing' : ''
                    } behavior for precise code intelligence.`}
                    className="mb-3"
                />
                {authenticatedUser?.siteAdmin && <PolicyListActions history={history} />}
            </CodeIntelConfigurationPageHeader>

            {history.location.state && (
                <FlashMessage state={history.location.state.modal} message={history.location.state.message} />
            )}
            <Container>
                <FilteredConnection<CodeIntelligenceConfigurationPolicyFields, {}>
                    listComponent="div"
                    listClassName={classNames(styles.grid, 'mb-3')}
                    showMoreClassName="mb-0"
                    noun="configuration policy"
                    pluralNoun="configuration policies"
                    querySubject={querySubject}
                    nodeComponent={PoliciesNode}
                    nodeComponentProps={{
                        indexingEnabled,
                        history,
                    }}
                    queryConnection={queryPoliciesCallback}
                    history={history}
                    location={props.location}
                    cursorPaging={true}
                    filters={filters}
                    emptyElement={<EmptyPoliciesList />}
                />
            </Container>
        </>
    )
}

export interface PoliciesNodeProps {
    node: CodeIntelligenceConfigurationPolicyFields
    indexingEnabled?: boolean
    history?: H.History
}

export const PoliciesNode: FunctionComponent<PoliciesNodeProps> = ({
    node: policy,
    indexingEnabled = false,
    history,
}) => (
    <>
        <span className={styles.separator} />

        <div className={classNames(styles.name, 'd-flex flex-column')}>
            <div className="m-0">
                <h3 className="m-0 d-block d-md-inline">{policy.name}</h3>
            </div>

            <div>
                <div className="mr-2 d-block d-mdinline-block">
                    Applied to{' '}
                    {policy.type === GitObjectType.GIT_COMMIT
                        ? 'commits'
                        : policy.type === GitObjectType.GIT_TAG
                        ? 'tags'
                        : policy.type === GitObjectType.GIT_TREE
                        ? 'branches'
                        : ''}{' '}
                    matching <span className="text-monospace">{policy.pattern}</span>
                    {policy.repository ? (
                        ` of ${policy.repository.name}`
                    ) : policy.repositoryPatterns ? (
                        <>
                            {' '}
                            in repositories matching{' '}
                            {policy.repositoryPatterns.map((pattern, index) => (
                                <React.Fragment key={pattern}>
                                    {index !== 0 &&
                                        (index === (policy.repositoryPatterns || []).length - 1 ? <>, or </> : <>, </>)}
                                    <span key={pattern} className="text-monospace">
                                        {pattern}
                                    </span>
                                </React.Fragment>
                            ))}
                        </>
                    ) : (
                        <> in any repository.</>
                    )}
                </div>

                <div>
                    {indexingEnabled && !policy.retentionEnabled && !policy.indexingEnabled ? (
                        <p className="text-muted mt-2">Data retention and auto-indexing disabled.</p>
                    ) : (
                        <>
                            <p className="mt-2">
                                <RetentionPolicyDescription policy={policy} />
                            </p>
                            {indexingEnabled && (
                                <p className="mt-2">
                                    <IndexingPolicyDescription policy={policy} />
                                </p>
                            )}
                        </>
                    )}
                </div>
            </div>
        </div>

        <span className={classNames(styles.button, 'd-none d-md-inline')}>
            <Link to={`./configuration/${policy.id}`} className="p-0">
                <ChevronRightIcon />
            </Link>
        </span>
    </>
)
