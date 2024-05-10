import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import {
    type GetDashboardInsightsResult,
    type GetInsightViewResult,
    GroupByField,
    type InsightsDashboardNode,
    type InsightsDashboardsResult,
    type InsightViewNode,
    SeriesSortDirection,
    SeriesSortMode,
    TimeIntervalStepUnit,
} from '../../../graphql-operations'

export const EMPTY_DASHBOARD: InsightsDashboardNode = {
    __typename: 'InsightsDashboard',
    id: 'EMPTY_DASHBOARD',
    title: 'Empty Dashboard',
    grants: {
        __typename: 'InsightsPermissionGrants',
        users: [testUserID],
        organizations: [],
        global: false,
    },
}

export const GET_DASHBOARD_INSIGHTS_EMPTY: GetDashboardInsightsResult = {
    insightsDashboards: {
        nodes: [
            {
                __typename: 'InsightsDashboard',
                id: EMPTY_DASHBOARD.id,
                views: { nodes: [] },
            },
        ],
    },
}

export const CAPTURE_GROUP_INSIGHT: InsightViewNode = {
    id: 'aW5zaWdodF92aWV3OiIyQnF6ZnBQQzFYUVJTeFpkUnhOWk5jYW1jQ1ki',
    defaultSeriesDisplayOptions: {
        limit: null,
        numSamples: null,
        sortOptions: {
            mode: SeriesSortMode.RESULT_COUNT,
            direction: SeriesSortDirection.DESC,
            __typename: 'SeriesSortOptions',
        },
        __typename: 'SeriesDisplayOptions',
    },
    isFrozen: false,
    defaultFilters: {
        includeRepoRegex: '',
        excludeRepoRegex: '',
        searchContexts: [],
        __typename: 'InsightViewFilters',
    },
    dashboardReferenceCount: 2,
    dashboards: {
        nodes: [
            {
                id: 'ZGFzaGJvYXJkOnsiSWRUeXBlIjoiY3VzdG9tIiwiQXJnIjoyfQ==',
                title: 'Cloud cost optimization',
                __typename: 'InsightsDashboard',
            },
            {
                id: 'ZGFzaGJvYXJkOnsiSWRUeXBlIjoiY3VzdG9tIiwiQXJnIjoxN30=',
                title: 'Each Type of Insight',
                __typename: 'InsightsDashboard',
            },
        ],
        __typename: 'InsightsDashboardConnection',
    },
    presentation: {
        __typename: 'LineChartInsightViewPresentation',
        title: 'Capture Group',
        seriesPresentation: [
            {
                seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8',
                label: '',
                color: '',
                __typename: 'LineChartDataSeriesPresentation',
            },
        ],
    },
    repositoryDefinition: {
        __typename: 'InsightRepositoryScope',
        repositories: [],
    },
    dataSeriesDefinitions: [
        {
            seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8',
            query: 'machine_type \\"([\\w]+\\-[\\w]+[\\-[\\w]+]?)\\" lang:Terraform  patterntype:regexp',
            timeScope: {
                unit: TimeIntervalStepUnit.MONTH,
                value: 1,
                __typename: 'InsightIntervalTimeScope',
            },
            isCalculated: true,
            generatedFromCaptureGroups: true,
            groupBy: null,
            __typename: 'SearchInsightDataSeriesDefinition',
        },
    ],
    __typename: 'InsightView',
}

export const SEARCH_BASED_INSIGHT: InsightViewNode = {
    id: 'aW5zaWdodF92aWV3OiIyQmRnV2VFYktwWGF2UjlGcXpuVDA1cld0c2si',
    defaultSeriesDisplayOptions: {
        limit: null,
        numSamples: null,
        sortOptions: {
            mode: null,
            direction: null,
            __typename: 'SeriesSortOptions',
        },
        __typename: 'SeriesDisplayOptions',
    },
    isFrozen: false,
    defaultFilters: {
        includeRepoRegex: '',
        excludeRepoRegex: '',
        searchContexts: [],
        __typename: 'InsightViewFilters',
    },
    dashboardReferenceCount: 3,
    dashboards: {
        nodes: [
            {
                id: 'ZGFzaGJvYXJkOnsiSWRUeXBlIjoiY3VzdG9tIiwiQXJnIjoxfQ==',
                title: 'Dev Experience',
                __typename: 'InsightsDashboard',
            },
            {
                id: 'ZGFzaGJvYXJkOnsiSWRUeXBlIjoiY3VzdG9tIiwiQXJnIjo1fQ==',
                title: 'asd',
                __typename: 'InsightsDashboard',
            },
            {
                id: 'ZGFzaGJvYXJkOnsiSWRUeXBlIjoiY3VzdG9tIiwiQXJnIjoxN30=',
                title: 'Each Type of Insight',
                __typename: 'InsightsDashboard',
            },
        ],
        __typename: 'InsightsDashboardConnection',
    },
    presentation: {
        __typename: 'LineChartInsightViewPresentation',
        title: 'Search Based',
        seriesPresentation: [
            {
                seriesId: '2D2MUtp6DzHhwhjUo9mIlBbhqoO',
                label: 'exec',
                color: 'var(--oc-grape-7)',
                __typename: 'LineChartDataSeriesPresentation',
            },
            {
                seriesId: '2D2MUHBTNUe5v18Je4o9woc8zrH',
                label: 'sourcegraph/run',
                color: 'var(--oc-teal-7)',
                __typename: 'LineChartDataSeriesPresentation',
            },
        ],
    },
    repositoryDefinition: {
        __typename: 'InsightRepositoryScope',
        repositories: [
            'github.com/sourcegraph/sourcegraph',
            'github.com/sourcegraph/deploy-sourcegraph-managed',
            'github.com/sourcegraph/infrastructure',
            'github.com/sourcegraph/deploy-sourcegraph-cloud',
        ],
    },
    dataSeriesDefinitions: [
        {
            seriesId: '2D2MUtp6DzHhwhjUo9mIlBbhqoO',
            query: 'lang:go exec.Cmd OR exec.CommandContext',
            timeScope: {
                unit: TimeIntervalStepUnit.WEEK,
                value: 2,
                __typename: 'InsightIntervalTimeScope',
            },
            isCalculated: true,
            generatedFromCaptureGroups: false,
            groupBy: null,
            __typename: 'SearchInsightDataSeriesDefinition',
        },
        {
            seriesId: '2D2MUHBTNUe5v18Je4o9woc8zrH',
            query: 'lang:go content:"github.com/sourcegraph/run" AND (run.Cmd OR run.Bash)',
            timeScope: {
                unit: TimeIntervalStepUnit.WEEK,
                value: 2,
                __typename: 'InsightIntervalTimeScope',
            },
            isCalculated: true,
            generatedFromCaptureGroups: false,
            groupBy: null,
            __typename: 'SearchInsightDataSeriesDefinition',
        },
    ],
    __typename: 'InsightView',
}

export const LANG_STATS_INSIGHT: InsightViewNode = {
    id: 'aW5zaWdodF92aWV3OiIyQ3VMOXlZbXNndHI3NW05NEpUY3BWWVFNMFoi',
    defaultSeriesDisplayOptions: {
        limit: null,
        numSamples: null,
        sortOptions: {
            mode: null,
            direction: null,
            __typename: 'SeriesSortOptions',
        },
        __typename: 'SeriesDisplayOptions',
    },
    isFrozen: false,
    defaultFilters: {
        includeRepoRegex: null,
        excludeRepoRegex: null,
        searchContexts: [],
        __typename: 'InsightViewFilters',
    },
    dashboardReferenceCount: 2,
    dashboards: {
        nodes: [
            {
                id: 'ZGFzaGJvYXJkOnsiSWRUeXBlIjoiY3VzdG9tIiwiQXJnIjo3fQ==',
                title: 'CPT Sourcegraph',
                __typename: 'InsightsDashboard',
            },
            {
                id: 'ZGFzaGJvYXJkOnsiSWRUeXBlIjoiY3VzdG9tIiwiQXJnIjoxN30=',
                title: 'Each Type of Insight',
                __typename: 'InsightsDashboard',
            },
        ],
        __typename: 'InsightsDashboardConnection',
    },
    presentation: {
        __typename: 'PieChartInsightViewPresentation',
        title: 'Lang Stats',
        otherThreshold: 0.03,
    },
    repositoryDefinition: {
        repositories: ['github.com/sourcegraph/about'],
        __typename: 'InsightRepositoryScope',
    },
    dataSeriesDefinitions: [
        {
            seriesId: '2CuLABWoJVNlP8KqoB49hdes8MK',
            query: '',
            timeScope: {
                unit: TimeIntervalStepUnit.MONTH,
                value: 0,
                __typename: 'InsightIntervalTimeScope',
            },
            isCalculated: false,
            generatedFromCaptureGroups: false,
            groupBy: null,
            __typename: 'SearchInsightDataSeriesDefinition',
        },
    ],
    __typename: 'InsightView',
}

export const COMPUTE_INSIGHT: InsightViewNode = {
    id: 'aW5zaWdodF92aWV3OiIyRjdlUk1Tc1ZoUHRBd0FKNzJ2TEJEOWZEQUgi',
    defaultSeriesDisplayOptions: {
        limit: null,
        numSamples: null,
        sortOptions: { mode: null, direction: null, __typename: 'SeriesSortOptions' },
        __typename: 'SeriesDisplayOptions',
    },
    isFrozen: false,
    defaultFilters: {
        includeRepoRegex: '',
        excludeRepoRegex: '',
        searchContexts: [],
        __typename: 'InsightViewFilters',
    },
    dashboardReferenceCount: 1,
    dashboards: {
        nodes: [
            {
                id: 'ZGFzaGJvYXJkOnsiSWRUeXBlIjoiY3VzdG9tIiwiQXJnIjoxNn0=',
                title: 'help',
                __typename: 'InsightsDashboard',
            },
        ],
        __typename: 'InsightsDashboardConnection',
    },
    presentation: {
        __typename: 'LineChartInsightViewPresentation',
        title: 'Compute Insight',
        seriesPresentation: [
            {
                seriesId: '2F7eRYTr4EyEblhHeoQE2lRXG2y',
                label: 'dep case:yes',
                color: 'var(--oc-orange-7)',
                __typename: 'LineChartDataSeriesPresentation',
            },
        ],
    },
    repositoryDefinition: {
        repositories: ['github.com/sourcegraph/test_DEPRECATED', 'github.com/sourcegraph/deploy-k8s-helper'],
        __typename: 'InsightRepositoryScope',
    },
    dataSeriesDefinitions: [
        {
            seriesId: '2F7eRYTr4EyEblhHeoQE2lRXG2y',
            query: 'DEP case:yes',
            timeScope: { unit: TimeIntervalStepUnit.WEEK, value: 2, __typename: 'InsightIntervalTimeScope' },
            isCalculated: true,
            generatedFromCaptureGroups: true,
            groupBy: GroupByField.REPO,
            __typename: 'SearchInsightDataSeriesDefinition',
        },
    ],
    __typename: 'InsightView',
}

export const GET_DASHBOARD_INSIGHTS_POPULATED: GetDashboardInsightsResult = {
    insightsDashboards: {
        nodes: [
            {
                id: 'EACH_TYPE_OF_INSIGHT',
                views: {
                    nodes: [CAPTURE_GROUP_INSIGHT, null, LANG_STATS_INSIGHT, SEARCH_BASED_INSIGHT, COMPUTE_INSIGHT],
                    __typename: 'InsightViewConnection',
                },
                __typename: 'InsightsDashboard',
            },
        ],
        __typename: 'InsightsDashboardConnection',
    },
}

export const INSIGHTS_DASHBOARDS: InsightsDashboardsResult = {
    currentUser: {
        __typename: 'User',
        id: testUserID,
        organizations: { nodes: [] },
    },
    insightsDashboards: {
        __typename: 'InsightsDashboardConnection',
        nodes: [
            EMPTY_DASHBOARD,
            {
                id: 'EACH_TYPE_OF_INSIGHT',
                title: 'Each Type of Insight',
                grants: {
                    users: [testUserID],
                    organizations: [],
                    global: false,
                    __typename: 'InsightsPermissionGrants',
                },
                __typename: 'InsightsDashboard',
            },
        ],
    },
}

export const GET_INSIGHT_VIEW_CAPTURE_GROUP_INSIGHT: GetInsightViewResult = {
    insightViews: {
        nodes: [
            {
                id: CAPTURE_GROUP_INSIGHT.id,
                dataSeries: [
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n2-highmem-4',
                        label: 'n2-highmem-4',
                        points: [
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 10,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 14,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 14,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n1-standard-1',
                        label: 'n1-standard-1',
                        points: [
                            {
                                dateTime: '2021-10-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n1-standard-2',
                        label: 'n1-standard-2',
                        points: [
                            {
                                dateTime: '2021-10-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n2-standard-8',
                        label: 'n2-standard-8',
                        points: [
                            {
                                dateTime: '2021-10-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 10,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n1-standard-32',
                        label: 'n1-standard-32',
                        points: [
                            {
                                dateTime: '2021-10-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 4,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n2-standard-16',
                        label: 'n2-standard-16',
                        points: [
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 4,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n1-standard-8',
                        label: 'n1-standard-8',
                        points: [
                            {
                                dateTime: '2021-10-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 4,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 7,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 7,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 4,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 4,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n1-highmem-4',
                        label: 'n1-highmem-4',
                        points: [
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n1-standard-16',
                        label: 'n1-standard-16',
                        points: [
                            {
                                dateTime: '2021-10-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 7,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 7,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 5,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 4,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 4,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n2-highmem-32',
                        label: 'n2-highmem-32',
                        points: [
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n2-highmem-8',
                        label: 'n2-highmem-8',
                        points: [
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n2-standard-32',
                        label: 'n2-standard-32',
                        points: [
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-c2-standard-4',
                        label: 'c2-standard-4',
                        points: [
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-e2-standard-8',
                        label: 'e2-standard-8',
                        points: [
                            {
                                dateTime: '2021-10-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n1-highmem-8',
                        label: 'n1-highmem-8',
                        points: [
                            {
                                dateTime: '2021-10-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n1-standard-64',
                        label: 'n1-standard-64',
                        points: [
                            {
                                dateTime: '2021-10-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n2-standard-48',
                        label: 'n2-standard-48',
                        points: [
                            {
                                dateTime: '2021-10-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-n2-standard-64',
                        label: 'n2-standard-64',
                        points: [
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-c2-standard-8',
                        label: 'c2-standard-8',
                        points: [
                            {
                                dateTime: '2021-11-21T00:00:00Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2021-12-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-01-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-02-21T00:00:00Z',
                                value: 3,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-21T00:00:00Z',
                                value: 7,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-21T00:00:00Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-21T00:00:00Z',
                                value: 8,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-21T00:00:00Z',
                                value: 7,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-21T00:00:00Z',
                                value: 6,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2CGKrC1dcbOpawrHQUOkiSu0NC8-c2d-highcpu-4',
                        label: 'c2d-highcpu-4',
                        points: [
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-21T18:07:48Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:12Z',
                                value: 1,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                ],
                __typename: 'InsightView',
            },
        ],
        __typename: 'InsightViewConnection',
    },
}

export const GET_INSIGHT_VIEW_SEARCH_BASED_INSIGHT: GetInsightViewResult = {
    insightViews: {
        nodes: [
            {
                id: SEARCH_BASED_INSIGHT.id,
                dataSeries: [
                    {
                        seriesId: '2D2MUHBTNUe5v18Je4o9woc8zrH',
                        label: 'sourcegraph/run',
                        points: [
                            {
                                dateTime: '2022-05-15T00:00:00Z',
                                value: 16,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-29T00:00:00Z',
                                value: 68,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-12T00:00:00Z',
                                value: 84,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-26T00:00:00Z',
                                value: 89,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-10T00:00:00Z',
                                value: 94,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-24T00:00:00Z',
                                value: 104,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-07T00:00:00Z',
                                value: 98,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:36Z',
                                value: 103,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-04T18:07:32Z',
                                value: 99,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-18T18:07:51Z',
                                value: 101,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:09Z',
                                value: 100,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                    {
                        seriesId: '2D2MUtp6DzHhwhjUo9mIlBbhqoO',
                        label: 'exec',
                        points: [
                            {
                                dateTime: '2022-03-06T00:00:00Z',
                                value: 142,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-03-20T00:00:00Z',
                                value: 142,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-03T00:00:00Z',
                                value: 158,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-04-17T00:00:00Z',
                                value: 167,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-01T00:00:00Z',
                                value: 137,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-15T00:00:00Z',
                                value: 137,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-05-29T00:00:00Z',
                                value: 132,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-12T00:00:00Z',
                                value: 130,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-06-26T00:00:00Z',
                                value: 160,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-10T00:00:00Z',
                                value: 160,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-07-24T00:00:00Z',
                                value: 159,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-07T00:00:00Z',
                                value: 160,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-08-21T17:31:37Z',
                                value: 160,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-04T18:07:31Z',
                                value: 160,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-18T18:07:51Z',
                                value: 160,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                            {
                                dateTime: '2022-09-22T00:33:08Z',
                                value: 161,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                ],
                __typename: 'InsightView',
            },
        ],
        __typename: 'InsightViewConnection',
    },
}

export const GET_INSIGHT_VIEW_COMPUTE_INSIGHT: GetInsightViewResult = {
    insightViews: {
        nodes: [
            {
                id: 'aW5zaWdodF92aWV3OiIyRjdlUk1Tc1ZoUHRBd0FKNzJ2TEJEOWZEQUgi',
                dataSeries: [
                    {
                        seriesId: '2F7eRYTr4EyEblhHeoQE2lRXG2y-github.com/sourcegraph/test_DEPRECATED',
                        label: 'github.com/sourcegraph/test_DEPRECATED',
                        points: [
                            {
                                dateTime: '2022-09-23T00:06:59Z',
                                value: 2,
                                __typename: 'InsightDataPoint',
                                pointInTimeQuery: 'type:diff',
                            },
                        ],
                        status: {
                            isLoadingData: false,
                            incompleteDatapoints: [],
                            __typename: 'InsightSeriesStatus',
                        },
                        __typename: 'InsightsSeries',
                    },
                ],
                __typename: 'InsightView',
            },
        ],
        __typename: 'InsightViewConnection',
    },
}

export const LANG_STAT_INSIGHT_CONTENT = {
    search: {
        results: { limitHit: false },
        stats: {
            languages: [
                { name: 'Markdown', totalLines: 83440 },
                { name: 'TypeScript', totalLines: 22302 },
                { name: 'SVG', totalLines: 15436 },
                { name: 'YAML', totalLines: 9483 },
                { name: 'HTML', totalLines: 5011 },
                { name: 'SCSS', totalLines: 2127 },
                { name: 'JavaScript', totalLines: 289 },
                { name: 'TOML', totalLines: 181 },
                { name: 'CSS', totalLines: 151 },
                { name: 'Go', totalLines: 147 },
                { name: 'JSON', totalLines: 146 },
                { name: 'Shell', totalLines: 84 },
                { name: 'Jsonnet', totalLines: 15 },
                { name: 'Ignore List', totalLines: 3 },
            ],
        },
    },
}
