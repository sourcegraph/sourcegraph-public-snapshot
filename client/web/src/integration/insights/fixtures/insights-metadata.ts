import { InsightViewNode, SeriesSortDirection, SeriesSortMode, TimeIntervalStepUnit } from '../../../graphql-operations'

const DEFAULT_SERIES_DISPLAY_OPTIONS = {
    limit: 20,
    sortOptions: {
        direction: SeriesSortDirection.DESC,
        mode: SeriesSortMode.RESULT_COUNT,
    },
}

interface InsightOptions {
    type: 'calculated' | 'just-in-time'
}

export const createJITMigrationToGQLInsightMetadataFixture = (options: InsightOptions): InsightViewNode => ({
    __typename: 'InsightView',
    id: '001',
    appliedSeriesDisplayOptions: DEFAULT_SERIES_DISPLAY_OPTIONS,
    defaultSeriesDisplayOptions: DEFAULT_SERIES_DISPLAY_OPTIONS,
    dashboardReferenceCount: 0,
    isFrozen: false,
    appliedFilters: {
        __typename: 'InsightViewFilters',
        searchContexts: [],
        includeRepoRegex: '',
        excludeRepoRegex: '',
    },
    dashboards: { nodes: [] },
    presentation: {
        __typename: 'LineChartInsightViewPresentation',
        title: 'Migration to new GraphQL TS types',
        seriesPresentation: [
            {
                __typename: 'LineChartDataSeriesPresentation',
                seriesId: '001',
                label: 'Imports of old GQL.* types',
                color: 'var(--oc-red-7)',
            },
            {
                __typename: 'LineChartDataSeriesPresentation',
                seriesId: '002',
                label: 'Imports of new graphql-operations types',
                color: 'var(--oc-blue-7)',
            },
        ],
    },
    dataSeriesDefinitions: [
        {
            __typename: 'SearchInsightDataSeriesDefinition',
            seriesId: '001',
            query: 'patternType:regex case:yes \\*\\sas\\sGQL',
            isCalculated: options.type === 'calculated',
            generatedFromCaptureGroups: false,
            repositoryScope: {
                __typename: 'InsightRepositoryScope',
                repositories: ['github.com/sourcegraph/sourcegraph'],
            },
            timeScope: {
                __typename: 'InsightIntervalTimeScope',
                unit: TimeIntervalStepUnit.WEEK,
                value: 6,
            },
        },
        {
            __typename: 'SearchInsightDataSeriesDefinition',
            seriesId: '002',
            query: "patternType:regexp case:yes /graphql-operations'",
            isCalculated: options.type === 'calculated',
            generatedFromCaptureGroups: false,
            repositoryScope: {
                __typename: 'InsightRepositoryScope',
                repositories: ['github.com/sourcegraph/sourcegraph'],
            },
            timeScope: {
                __typename: 'InsightIntervalTimeScope',
                unit: TimeIntervalStepUnit.WEEK,
                value: 6,
            },
        },
    ],
})

export const STORYBOOK_GROWTH_INSIGHT_METADATA_FIXTURE: InsightViewNode = {
    __typename: 'InsightView',
    id: '002',
    appliedSeriesDisplayOptions: DEFAULT_SERIES_DISPLAY_OPTIONS,
    defaultSeriesDisplayOptions: DEFAULT_SERIES_DISPLAY_OPTIONS,
    dashboardReferenceCount: 0,
    dashboards: { nodes: [] },
    isFrozen: false,
    appliedFilters: {
        __typename: 'InsightViewFilters',
        includeRepoRegex: '',
        excludeRepoRegex: '',
        searchContexts: [],
    },
    presentation: {
        __typename: 'LineChartInsightViewPresentation',
        title: 'Team head count',
        seriesPresentation: [
            {
                __typename: 'LineChartDataSeriesPresentation',
                seriesId: '001',
                label: 'Client storybook tests',
                color: 'var(--oc-blue-7)',
            },
        ],
    },
    dataSeriesDefinitions: [
        {
            __typename: 'SearchInsightDataSeriesDefinition',
            seriesId: '001',
            query: 'patternType:regexp f:\\.story\\.tsx$ \\badd\\(',
            isCalculated: false,
            generatedFromCaptureGroups: false,
            repositoryScope: {
                __typename: 'InsightRepositoryScope',
                repositories: ['github.com/sourcegraph/sourcegraph'],
            },
            timeScope: {
                __typename: 'InsightIntervalTimeScope',
                unit: TimeIntervalStepUnit.WEEK,
                value: 6,
            },
        },
    ],
}

export const SOURCEGRAPH_LANG_STATS_INSIGHT_METADATA_FIXTURE: InsightViewNode = {
    __typename: 'InsightView',
    id: '003',
    appliedSeriesDisplayOptions: DEFAULT_SERIES_DISPLAY_OPTIONS,
    defaultSeriesDisplayOptions: DEFAULT_SERIES_DISPLAY_OPTIONS,
    dashboardReferenceCount: 0,
    dashboards: { nodes: [] },
    isFrozen: false,
    appliedFilters: {
        __typename: 'InsightViewFilters',
        includeRepoRegex: '',
        excludeRepoRegex: '',
        searchContexts: [],
    },
    presentation: {
        __typename: 'PieChartInsightViewPresentation',
        title: 'Sourcegraph languages',
        otherThreshold: 0.03,
    },
    dataSeriesDefinitions: [
        {
            seriesId: '001',
            query: '',
            repositoryScope: {
                repositories: ['github.com/sourcegraph/sourcegraph'],
                __typename: 'InsightRepositoryScope',
            },
            timeScope: {
                unit: TimeIntervalStepUnit.MONTH,
                value: 0,
                __typename: 'InsightIntervalTimeScope',
            },
            isCalculated: false,
            generatedFromCaptureGroups: false,
            __typename: 'SearchInsightDataSeriesDefinition',
        },
    ],
}
