import {
    type InsightViewNode,
    SeriesSortDirection,
    SeriesSortMode,
    TimeIntervalStepUnit,
} from '../../../graphql-operations'

const DEFAULT_SERIES_DISPLAY_OPTIONS = {
    limit: null,
    numSamples: null,
    sortOptions: {
        direction: SeriesSortDirection.DESC,
        mode: SeriesSortMode.RESULT_COUNT,
    },
}

interface InsightOptions {
    id?: string
    type: 'calculated' | 'just-in-time'
}

export const createJITMigrationToGQLInsightMetadataFixture = (options: InsightOptions): InsightViewNode => ({
    __typename: 'InsightView',
    id: options.id ?? '001',
    defaultSeriesDisplayOptions: DEFAULT_SERIES_DISPLAY_OPTIONS,
    dashboardReferenceCount: 0,
    isFrozen: false,
    defaultFilters: {
        __typename: 'InsightViewFilters',
        searchContexts: [],
        includeRepoRegex: '',
        excludeRepoRegex: '',
    },
    dashboards: { nodes: [] },
    repositoryDefinition: {
        __typename: 'InsightRepositoryScope',
        repositories: ['github.com/sourcegraph/sourcegraph'],
    },
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
            timeScope: {
                __typename: 'InsightIntervalTimeScope',
                unit: TimeIntervalStepUnit.WEEK,
                value: 6,
            },
            groupBy: null,
        },
        {
            __typename: 'SearchInsightDataSeriesDefinition',
            seriesId: '002',
            query: "patternType:regexp case:yes /graphql-operations'",
            isCalculated: options.type === 'calculated',
            generatedFromCaptureGroups: false,
            timeScope: {
                __typename: 'InsightIntervalTimeScope',
                unit: TimeIntervalStepUnit.WEEK,
                value: 6,
            },
            groupBy: null,
        },
    ],
})
