import { InsightType } from '../../../../core/types'
import { CaptureInsightUrlValues } from '../../../insights/creation/capture-group'
import { DATA_SERIES_COLORS, SearchInsightURLValues } from '../../../insights/creation/search-insight'

export interface TemplateSection {
    title: string
    templates: Template[]
}

export type Template = SearchTemplate | CaptureGroupTemplate

interface SearchTemplate {
    type: InsightType.SearchBased
    title: string
    description: string
    templateValues: Partial<SearchInsightURLValues>
}

interface CaptureGroupTemplate {
    type: InsightType.CaptureGroup
    title: string
    description: string
    templateValues: Partial<CaptureInsightUrlValues>
}

export const TEMPLATE_SECTIONS: TemplateSection[] = [
    {
        title: 'Popular',
        templates: [
            {
                type: InsightType.CaptureGroup,
                title: 'Java versions',
                description: 'Detect and track which Java versions are present or most popular in your code base.',
                templateValues: {
                    title: 'Java versions',
                    groupSearchQuery: 'file:pom\\.xml$ <java\\.version>(.*)</java\\.version> archived:no fork:no',
                },
            },
            {
                type: InsightType.CaptureGroup,
                title: 'Terraform versions',
                description: 'Detect and track which Terraform versions are present or most popular in your code base.',
                templateValues: {
                    title: 'Terraform versions',
                    groupSearchQuery:
                        'repoapp.terraform.io/(.*)\\n version =(.*)([0-9].[0-9].[0-9]) lang:Terraform archived:no fork:no',
                },
            },
            {
                type: InsightType.SearchBased,
                title: 'Yarn adoption',
                description: 'Are more groups or teams using yarn? Track yarn adoption.',
                templateValues: {
                    title: 'Yarn adoption',
                    allRepos: true,
                    series: [
                        {
                            name: 'Yarn',
                            query: 'select:repo file:yarn.lock archived:no fork:no',
                            stroke: DATA_SERIES_COLORS.BLUE,
                        },
                    ],
                },
            },
            {
                type: InsightType.SearchBased,
                title: 'Linter override rules',
                description: 'Code hygiene and health: How many linter override rules exist',
                templateValues: {
                    title: 'Rule overrides',
                    allRepos: true,
                    series: [
                        {
                            name: 'Rule overrides',
                            query: 'file:^\\.eslintignore .\\n patternType:regexp archived:no fork:no',
                            stroke: DATA_SERIES_COLORS.ORANGE,
                        },
                    ],
                },
            },
        ],
    },
    {
        title: 'Migration',
        templates: [
            {
                type: InsightType.SearchBased,
                title: 'React function component migration',
                description: 'Track your migration from class based component to function based',
                templateValues: {
                    title: 'React function component migration',
                    series: [
                        {
                            name: 'Function components',
                            query: 'patternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent',
                            stroke: DATA_SERIES_COLORS.ORANGE,
                        },
                        {
                            name: 'Class components',
                            query: 'patternType:regexp extends\\s(React\\.)?(Pure)?Component',
                            stroke: DATA_SERIES_COLORS.BLUE,
                        },
                    ],
                },
            },
            {
                type: InsightType.SearchBased,
                title: 'CSS module migration',
                description: 'Track how many CSS modules files you have compared to global CSS files',
                templateValues: {
                    title: 'CSS module migration',
                    series: [
                        {
                            name: 'Global CSS',
                            query:
                                'context:global type:file lang:scss -file:module.scss -file:global-styles patterntype:regexp',
                            stroke: DATA_SERIES_COLORS.RED,
                        },
                        {
                            name: 'CSS Modules',
                            query: 'context:global type:file lang:scss file:module.scss patterntype:regexp',
                            stroke: DATA_SERIES_COLORS.GREEN,
                        },
                    ],
                },
            },
            {
                type: InsightType.SearchBased,
                title: 'Migration to new GraphQL TS types Diff Search',
                description: 'Track migration from global gql types to GQL generated types',
                templateValues: {
                    title: 'Migration to new GraphQL TS types Diff Search',
                    series: [
                        {
                            name: 'Imports of old GQL.* types',
                            query: 'patternType:regex case:yes \\*\\sas\\sGQL type:diff',
                            stroke: DATA_SERIES_COLORS.RED,
                        },
                        {
                            name: 'Imports of new graphql-operations types',
                            query: "patternType:regexp case:yes /graphql-operations' type:diff",
                            stroke: DATA_SERIES_COLORS.GREEN,
                        },
                    ],
                },
            },
        ],
    },
]
