import { ApolloClient, gql } from '@apollo/client'
import { Observable, throwError, of, from } from 'rxjs'
import { map, mapTo } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'
import {
    CreateDashboardResult,
    CreateInsightsDashboardInput,
    DeleteDashboardResult,
    InsightsDashboardsResult,
    InsightsPermissionGrantsInput,
    UpdateDashboardResult,
    UpdateInsightsDashboardInput,
} from '@sourcegraph/web/src/graphql-operations'

import { Insight, InsightDashboard, InsightsDashboardType } from '../types'
import { SupportedInsightSubject } from '../types/subjects'

import { CodeInsightsBackend } from './code-insights-backend'
import {
    DashboardCreateInput,
    DashboardDeleteInput,
    DashboardUpdateInput,
    RepositorySuggestionData,
} from './code-insights-backend-types'

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

export class CodeInsightsGqlBackend implements CodeInsightsBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    // Insights
    public getInsights = (ids?: string[]): Observable<Insight[]> => {
        console.warn('TODO: Implement getInsights for GraphQL API')
        return of([])
    }
    public getInsightById = errorMockMethod('getInsightById')
    public findInsightByName = errorMockMethod('findInsightByName')
    public getReachableInsights = errorMockMethod('getReachableInsights')
    public getBackendInsightData = errorMockMethod('getBackendInsightData')
    public getBuiltInInsightData = errorMockMethod('getBuiltInInsightData')
    public getInsightSubjects = (): Observable<SupportedInsightSubject[]> => {
        console.warn('TODO: Get insight subjects')
        return of([])
    }
    public getSubjectSettingsById = errorMockMethod('getSubjectSettingsById')
    public createInsight = errorMockMethod('createInsight')
    public createInsightWithNewFilters = errorMockMethod('createInsightWithNewFilters')
    public updateInsight = errorMockMethod('updateInsight')
    public deleteInsight = errorMockMethod('deleteInsight')

    // Dashboards
    public getDashboards = (): Observable<InsightDashboard[]> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<InsightsDashboardsResult>({
                query: gql`
                    query InsightsDashboards {
                        insightsDashboards {
                            nodes {
                                id
                                title
                                views {
                                    nodes {
                                        id
                                    }
                                }
                                grants {
                                    users
                                    organizations
                                    global
                                }
                            }
                        }
                    }
                `,
            })
        ).pipe(
            map(({ data }) =>
                data.insightsDashboards.nodes.map(
                    (dashboard): InsightDashboard => ({
                        id: dashboard.id,
                        title: dashboard.title,
                        insightIds: dashboard.views?.nodes.map(view => view.id),
                        grants: dashboard.grants,
                        type: this.parseType(dashboard.grants),
                    })
                )
            )
        )
    public getDashboardById = (dashboardId?: string): Observable<InsightDashboard | undefined> =>
        this.getDashboards().pipe(map(dashboards => dashboards.find(({ id }) => id === dashboardId)))

    public findDashboardByName = errorMockMethod('findDashboardByName')

    /**
     * Helper function to parse the dashboard type from the grants object.
     * TODO: Remove this function when settings api is deprecated
     *
     * @param grants {object} - A grants object from an insight dashboard
     * @param grants.global {boolean}
     * @param grants.users {string[]}
     * @param grants.organizations {string[]}
     * @returns - The type of the dashboard
     */
    private parseType(grants?: {
        global?: boolean
        users?: string[]
        organizations?: string[]
    }): InsightsDashboardType.Personal | InsightsDashboardType.Organization | InsightsDashboardType.Global {
        if (grants?.global) {
            return InsightsDashboardType.Global
        }
        if (grants?.organizations?.length) {
            return InsightsDashboardType.Organization
        }
        return InsightsDashboardType.Personal
    }

    /**
     * Helper function to parse a grants object from a given type and visibility.
     * TODO: Remove this function when settings api is deprecated
     *
     * @param type {('personal'|'organization'|'global')} - The type of the dashboard
     * @param visibility {string} - Usually the user or organization id
     * @returns - A properly formatted grants object
     */
    private parseGrants = (type: string, visibility: string): InsightsPermissionGrantsInput => {
        const grants: InsightsPermissionGrantsInput = {}
        if (type === 'personal') {
            grants.users = [visibility]
        }
        if (type === 'organization') {
            grants.organizations = [visibility]
        }
        if (type === 'global') {
            grants.global = true
        }

        return grants
    }

    public createDashboard = (input: DashboardCreateInput): Observable<void> => {
        if (!input.type) {
            throw new Error('`grants` are required to create a new dashboard')
        }

        const mappedInput: CreateInsightsDashboardInput = {
            title: input.name,
            grants: this.parseGrants(input.type, input.visibility),
        }

        return from(
            this.apolloClient.mutate<CreateDashboardResult>({
                mutation: gql`
                    mutation CreateDashboard($input: CreateInsightsDashboardInput!) {
                        createInsightsDashboard(input: $input) {
                            dashboard {
                                id
                            }
                        }
                    }
                `,
                variables: { input: mappedInput },
            })
        ).pipe(mapTo(undefined))
    }

    public deleteDashboard = ({ id }: DashboardDeleteInput): Observable<void> => {
        if (!id) {
            throw new Error('`id` is required to delete a dashboard')
        }

        return from(
            this.apolloClient.mutate<DeleteDashboardResult>({
                mutation: gql`
                    mutation DeleteDashboard($id: ID!) {
                        deleteInsightsDashboard(id: $id) {
                            alwaysNil
                        }
                    }
                `,
                variables: { id },
            })
        ).pipe(mapTo(undefined))
    }

    public updateDashboard = ({ id, nextDashboardInput }: DashboardUpdateInput): Observable<void> => {
        if (!id) {
            throw new Error('`id` is required to update a dashboard')
        }

        if (!nextDashboardInput.type) {
            throw new Error('`grants` are required to update a dashboard')
        }

        const input: UpdateInsightsDashboardInput = {
            title: nextDashboardInput.name,
            grants: this.parseGrants(nextDashboardInput.type, nextDashboardInput.visibility),
        }

        return from(
            this.apolloClient.mutate<UpdateDashboardResult>({
                mutation: gql`
                    mutation UpdateDashboard($id: ID!, $input: UpdateInsightsDashboardInput!) {
                        updateInsightsDashboard(id: $id, input: $input) {
                            dashboard {
                                id
                            }
                        }
                    }
                `,
                variables: {
                    id,
                    input,
                },
            })
        ).pipe(mapTo(undefined))
    }

    // Live preview fetchers
    public getSearchInsightContent = (): Promise<LineChartContent<any, string>> =>
        errorMockMethod('getSearchInsightContent')().toPromise()
    public getLangStatsInsightContent = (): Promise<PieChartContent<any>> =>
        errorMockMethod('getLangStatsInsightContent')().toPromise()

    // Repositories API
    public getRepositorySuggestions = (): Promise<RepositorySuggestionData[]> =>
        errorMockMethod('getRepositorySuggestions')().toPromise()
    public getResolvedSearchRepositories = (): Promise<string[]> =>
        errorMockMethod('getResolvedSearchRepositories')().toPromise()
}
