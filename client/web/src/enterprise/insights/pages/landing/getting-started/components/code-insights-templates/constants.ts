import { InsightType } from '../../../../../core/types'
import { CaptureInsightUrlValues } from '../../../../insights/creation/capture-group'
import { DATA_SERIES_COLORS, SearchInsightURLValues } from '../../../../insights/creation/search-insight'

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

const TERRAFORM_VERSIONS: Template = {
    type: InsightType.SearchBased,
    title: 'Terraform versions',
    description: 'Detect and track which Terraform versions are present or most popular in your codebase',
    templateValues: {
        title: 'Terraform versions',
        allRepos: true,
        series: [
            {
                name: '1.1.0',
                query:
                    'app.terraform.io/(.*)\\n version =(.*)1.1.0 patternType:regexp lang:Terraform archived:no fork:no',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: '1.2.0',
                query:
                    'app.terraform.io/(.*)\\n version =(.*)1.2.0 patternType:regexp lang:Terraform archived:no fork:no',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const CSS_MODULES_MIGRATION: Template = {
    type: InsightType.SearchBased,
    title: 'Global CSS to CSS modules',
    description: 'Tracking migration from global CSS to CSS modules',
    templateValues: {
        title: 'Global CSS to CSS modules',
        allRepos: true,
        series: [
            {
                name: 'Global CSS',
                query: 'select:file lang:SCSS -file:module patterntype:regexp archived:no fork:no',
                stroke: DATA_SERIES_COLORS.RED,
            },
            {
                name: 'CSS Modules',
                query: 'select:file lang:SCSS file:module patterntype:regexp archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const LOG4J_FIXED_VERSIONS: Template = {
    type: InsightType.SearchBased,
    title: 'Vulnerable and fixed Log4j versions',
    description: 'Confirm that vulnerable versions of log4j are removed and only fixed versions appear',
    templateValues: {
        title: 'Vulnerable and fixed Log4j versions',
        allRepos: true,
        series: [
            {
                name: 'Vulnerable',
                query:
                    'lang:gradle org\\.apache\\.logging\\.log4j[\'"] 2\\.(0|1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16)(\\.[0-9]+) patterntype:regexp archived:no fork:no',
                stroke: DATA_SERIES_COLORS.RED,
            },
            {
                name: 'Fixed',
                query:
                    'lang:gradle org\\.apache\\.logging\\.log4j[\'"] 2\\.(17)(\\.[0-9]+) patterntype:regexp archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const YARN_ADOPTION: Template = {
    type: InsightType.SearchBased,
    title: 'Yarn adoption',
    description:
        'Are more repos increasingly using yarn? Track yarn adoption across teams and groups in your organization',
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
}

const JAVA_VERSIONS: Template = {
    type: InsightType.CaptureGroup,
    title: 'Java versions',
    description: 'Detect and track which Java versions are most popular in your codebase',
    templateValues: {
        title: 'Java versions',
        allRepos: true,
        groupSearchQuery: 'file:pom\\.xml$ <java\\.version>(.*)</java\\.version> archived:no fork:no',
    },
}

const LINTER_OVERRIDE_RULES: Template = {
    type: InsightType.SearchBased,
    title: 'Linter override rules',
    description: 'A code health indicator for how many linter override rules exist',
    templateValues: {
        title: 'Linter override rules',
        allRepos: true,
        series: [
            {
                name: 'Rule overrides',
                query: 'file:^\\.eslintignore ^[^#].*.\\n patternType:regexp archived:no fork:no',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const TS_JS_USAGE: Template = {
    type: InsightType.SearchBased,
    title: 'Language use over time',
    description: 'Track the growth of certain languages by file count',
    templateValues: {
        title: 'Language use over time',
        allRepos: true,
        series: [
            {
                name: 'TypeScript',
                query: 'select:file lang:TypeScript',
                stroke: DATA_SERIES_COLORS.INDIGO,
            },
            {
                name: 'JavaScript',
                query: 'select:file lang:JavaScript',
                stroke: DATA_SERIES_COLORS.YELLOW,
            },
        ],
    },
}

const CONFIG_OR_DOC_FILE: Template = {
    type: InsightType.SearchBased,
    title: 'Config or docs file',
    description: 'How many repos contain a config or docs file in a specific directory',
    templateValues: {
        title: 'Config or docs file',
        allRepos: true,
        series: [
            {
                name: 'Repositories with doc',
                query: 'select:repo file:docs/*/new_config_filename archived:no fork:no',
                stroke: DATA_SERIES_COLORS.PINK,
            },
        ],
    },
}

const ALLOW_DENY_LIST_TRACKING: Template = {
    type: InsightType.SearchBased,
    title: '“blacklist/whitelist” to “denylist/allowlist”',
    description: 'How the switch from files containing “blacklist/whitelist” to “denylist/allowlist” is progressing',
    templateValues: {
        title: '“blacklist/whitelist” to “denylist/allowlist”',
        allRepos: true,
        series: [
            {
                name: 'blacklist/whitelist',
                query: 'select:file blacklist OR whitelist archived:no fork:no',
                stroke: DATA_SERIES_COLORS.RED,
            },
            {
                name: 'denylist/allowlist',
                query: 'select:file denylist OR allowlist archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const PYTHON_2_3: Template = {
    type: InsightType.SearchBased,
    title: 'Python 2 to Python 3',
    description: 'How far along is the Python major version migration',
    templateValues: {
        title: 'Python 2 to Python 3',
        allRepos: true,
        series: [
            {
                name: 'Python 3',
                query: '#!/usr/bin/env python3 archived:no fork:no',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'Python 2',
                query: '#!/usr/bin/env python2 archived:no fork:no',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const REACT_FUNCTION_CLASS: Template = {
    type: InsightType.SearchBased,
    title: 'React Class to Function Components Migration',
    description: "What's the status of migrating to React function components from class components",
    templateValues: {
        title: 'React function component migration',
        series: [
            {
                name: 'Function components',
                query: 'patternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'Class components',
                query: 'patternType:regexp extends\\s(React\\.)?(Pure)?Component',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const NEW_API_USAGE: Template = {
    type: InsightType.SearchBased,
    title: 'New API usage',
    description: 'How many repos or teams are using a new API your team built',
    templateValues: {
        title: 'New API usage',
        series: [
            {
                name: 'New API',
                query: 'select:repo ourApiLibraryName.load archived:no fork:no',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
        ],
    },
}

const FREQUENTLY_USED_DATABASE: Template = {
    type: InsightType.SearchBased,
    title: 'Frequently used databases',
    description: 'Which databases we are calling or writing to most often',
    templateValues: {
        title: 'Frequently used databases',
        allRepos: true,
        series: [
            {
                name: 'Redis',
                query: 'redis\\.set patternType:regexp archived:no fork:no',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'GraphQL',
                query: 'graphql\\( patternType:regexp archived:no fork:no',
                stroke: DATA_SERIES_COLORS.PINK,
            },
        ],
    },
}

const LARGE_PACKAGE_USAGE: Template = {
    type: InsightType.SearchBased,
    title: 'Large or expensive package usage',
    description: 'Understand if a growing number of repos import a large/expensive package',
    templateValues: {
        title: 'Large or expensive package usage',
        allRepos: true,
        series: [
            {
                name: 'Repositories with large package usage',
                query: 'select:repo import\\slargePkg patternType:regexp archived:no fork:no',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const REACT_COMPONENT_LIB_USAGE: Template = {
    type: InsightType.SearchBased,
    title: 'React Component use',
    description: 'How many places are importing components from a library',
    templateValues: {
        title: 'React Component use',
        allRepos: true,
        series: [
            {
                name: 'Library imports',
                query: "from '@sourceLibrary/component' patternType:literal archived:no fork:no",
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
        ],
    },
}

const CI_TOOLING: Template = {
    type: InsightType.SearchBased,
    title: 'CI tooling adoption',
    description: 'How many repos are using our CI system',
    templateValues: {
        title: 'CI tooling adoption',
        allRepos: true,
        series: [
            {
                name: 'Repo with CircleCI config',
                query: 'file:\\.circleci/config.yml select:repo fork:no archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
        ],
    },
}

const CSS_CLASS: Template = {
    type: InsightType.SearchBased,
    title: 'CSS class',
    description: 'The removal of all deprecated CSS class',
    templateValues: {
        title: 'CSS class',
        allRepos: true,
        series: [
            {
                name: 'Deprecated CSS class',
                query: 'deprecated-class archived:no fork:no',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const ICON_OR_IMAGE: Template = {
    type: InsightType.SearchBased,
    title: 'Icon or image',
    description: 'The removal of all deprecated icon or image instances',
    templateValues: {
        title: 'Icon or image',
        allRepos: true,
        series: [
            {
                name: 'Deprecated logo',
                query: '2018logo.png archived:no fork:no',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const STRUCTURAL_CODE_PATTERN: Template = {
    type: InsightType.SearchBased,
    title: 'Structural code pattern',
    description:
        "Deprecating a structural code pattern in favor of a safer pattern, like how many tries don't have catches",
    templateValues: {
        title: 'Structural code pattern',
        allRepos: true,
        series: [
            {
                name: 'Try catch',
                query:
                    'try {:[_]} catch (:[e]) { } finally {:[_]} lang:java patternType:structural archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
        ],
    },
}

const TOOLING_MIGRATION: Template = {
    type: InsightType.SearchBased,
    title: 'Tooling',
    description: 'The progress of deprecating tooling you’re moving off of',
    templateValues: {
        title: 'Tooling',
        allRepos: true,
        series: [
            {
                name: 'Deprecated logger',
                query: 'deprecatedEventLogger.log archived:no fork:no',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const VAR_KEYWORDS: Template = {
    type: InsightType.SearchBased,
    title: 'Var keywords',
    description: 'Number of var keywords in the code basee (ES5 depreciation)',
    templateValues: {
        title: 'Var keywords',
        allRepos: true,
        series: [
            {
                name: 'var statements',
                query: '(lang:TypeScript OR lang:JavaScript) var ... = archived:no fork:no patterntype:structural',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const TESTING_LIBRARIES: Template = {
    type: InsightType.SearchBased,
    title: 'Consolidation of Testing Libraries',
    description: 'Which React test libraries are being consolidated',
    templateValues: {
        title: 'Consolidation of Testing Libraries',
        allRepos: true,
        series: [
            {
                name: '@testing-library',
                query: "from '@testing-library/react' archived:no fork:no",
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'enzume',
                query: "from 'enzyme' archived:no fork:no",
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const LICENSE_TYPES: Template = {
    type: InsightType.CaptureGroup,
    title: 'License types in the codebase',
    description: 'See the breakdown of licenses from package.json files',
    templateValues: {
        title: 'License types in the codebase',
        allRepos: true,
        groupSearchQuery: 'file:package.json "license":\\s"(.*)" archived:no fork:no',
    },
}

const ALL_LOG4J_VERSIONS: Template = {
    type: InsightType.CaptureGroup,
    title: 'All log4j versions',
    description: 'Which log4j versions are present, including vulnerable versions',
    templateValues: {
        title: 'All log4j versions',
        allRepos: true,
        groupSearchQuery: 'lang:gradle org\\.apache\\.logging\\.log4j[\'"] 2\\.([0-9]+)\\. archived:no fork:no',
    },
}

const PYTHON_VERSIONS: Template = {
    type: InsightType.CaptureGroup,
    title: 'Python versions',
    description: 'Which python versions are in use or haven’t been updated',
    templateValues: {
        title: 'Python versions',
        allRepos: true,
        groupSearchQuery: '#!/usr/bin/env python([0-9]\\.[0-9]+) archived:no fork:no',
    },
}

const NODEJS_VERSIONS: Template = {
    type: InsightType.CaptureGroup,
    title: 'Node.js versions',
    description: 'Which node.js versions are present based on nvm files',
    templateValues: {
        title: 'Node.js versions',
        allRepos: true,
        groupSearchQuery: 'nvm\\suse\\s([0-9]+\\.[0-9]+) archived:no fork:no',
    },
}

const CSS_COLORS: Template = {
    type: InsightType.CaptureGroup,
    title: 'CSS Colors',
    description: 'What CSS colors are present or most popular',
    templateValues: {
        title: 'CSS Colors',
        allRepos: true,
        groupSearchQuery: 'color:#([0-9a-fA-f]{3,6}) archived:no fork:no',
    },
}

const CHECKOV_SKIP_TYPES: Template = {
    type: InsightType.CaptureGroup,
    title: 'Types of checkov skips',
    description: 'See the most common reasons for why secuirty checks in checkov are skipped',
    templateValues: {
        title: 'Types of checkov skips',
        allRepos: true,
        groupSearchQuery: 'patterntype:regexp file:.tf #checkov:skip=(.*) archived:no fork:no',
    },
}

const TODOS: Template = {
    type: InsightType.SearchBased,
    title: 'TODOs',
    description: 'How many TODOs are in a specific part of the codebase (or all of it)',
    templateValues: {
        title: 'TODOs',
        allRepos: true,
        series: [
            {
                name: 'TODOs',
                query: 'TODO archived:no fork:no',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
        ],
    },
}

const REVERT_COMMITS: Template = {
    type: InsightType.SearchBased,
    title: 'Commits with “revert”',
    description: 'How frequently there are commits with “revert” in the commit message',
    templateValues: {
        title: 'Commits with “revert”',
        allRepos: true,
        series: [
            {
                name: 'Reverts',
                query: 'type:commit revert archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
        ],
    },
}

const DEPRECATED_CALLS: Template = {
    type: InsightType.SearchBased,
    title: 'Deprecated calls',
    description: 'How many times deprecated calls are used',
    templateValues: {
        title: 'Deprecated calls',
        allRepos: true,
        series: [
            {
                name: '@deprecated',
                query: 'lang:java @deprecated archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
        ],
    },
}

const STORYBOOK_TESTS: Template = {
    type: InsightType.SearchBased,
    title: 'Storybook tests',
    description: 'How many tests for Storybook exist',
    templateValues: {
        title: 'Storybook tests',
        allRepos: true,
        series: [
            {
                name: 'Stories',
                query: 'patternType:regexp f:\\.story\\.tsx$ \\badd\\( archived:no fork:no',
                stroke: DATA_SERIES_COLORS.PINK,
            },
        ],
    },
}

const REPOS_WITH_README: Template = {
    type: InsightType.SearchBased,
    title: 'Repos with Documentation',
    description: "How many repos do or don't have READMEs",
    templateValues: {
        title: 'Repos with Documentation',
        allRepos: true,
        series: [
            {
                name: 'with readme',
                query: 'repohasfile:readme.md select:repo archived:no fork:no',
                stroke: DATA_SERIES_COLORS.LIME,
            },
            {
                name: 'without readme',
                query: '-repohasfile:readme.md select:repo archived:no fork:no',
                stroke: DATA_SERIES_COLORS.YELLOW,
            },
        ],
    },
}

const OWNERSHIP_TRACKING: Template = {
    type: InsightType.SearchBased,
    title: 'Ownership via CODEOWNERS files',
    description: "How many repos do or don't have CODEOWNERS files",
    templateValues: {
        title: 'Ownership via CODEOWNERS files',
        allRepos: true,
        series: [
            {
                name: 'with readme',
                query: 'repohasfile:CODEOWNERS select:repo archived:no fork:no',
                stroke: DATA_SERIES_COLORS.LIME,
            },
            {
                name: 'without readme',
                query: '-repohasfile:CODEOWNERS select:repo archived:no fork:no',
                stroke: DATA_SERIES_COLORS.YELLOW,
            },
        ],
    },
}

const VULNERABLE_OPEN_SOURCE: Template = {
    type: InsightType.SearchBased,
    title: 'Vulnerable open source library',
    description:
        'Confirm that a vulnerable open source library has been fully removed, or see the speed of the deprecation',
    templateValues: {
        title: 'Vulnerable open source library',
        allRepos: true,
        series: [
            {
                name: 'vulnerableLibrary@14.3.9',
                query: 'vulnerableLibrary@14.3.9 archived:no fork:no',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const API_KEYS_DETECTION: Template = {
    type: InsightType.SearchBased,
    title: 'API keys',
    description: 'How quickly we notice and remove API keys when they are committed',
    templateValues: {
        title: 'API keys',
        allRepos: true,
        series: [
            {
                name: 'API key',
                query: 'regexMatchingAPIKey patternType:regexp archived:no fork:no',
                stroke: DATA_SERIES_COLORS.RED,
            },
        ],
    },
}

const SKIPPED_TESTS: Template = {
    type: InsightType.SearchBased,
    title: 'How many tests are skipped',
    description: 'See how many tests have skip conditions',
    templateValues: {
        title: 'How many tests are skipped',
        allRepos: true,
        series: [
            {
                name: 'Skipped tests',
                query: '(this.skip() OR it.skip) lang:TypeScript archived:no fork:no',
                stroke: DATA_SERIES_COLORS.RED,
            },
        ],
    },
}

const TEST_AMOUNT_AND_TYPES: Template = {
    type: InsightType.SearchBased,
    title: 'Tests amount and types',
    description: 'See what types of tests are most common and total counts',
    templateValues: {
        title: 'Tests amount and types',
        allRepos: true,
        series: [
            {
                name: 'e2e tests',
                query: 'patternType:regexp case:yes \\b(it|test)\\( f:/end-to-end/.*\\.test\\.ts$ archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'regression tests',
                query: 'patternType:regexp case:yes \\b(it|test)\\( f:/regression/.*\\.test\\.ts$ archived:no fork:no',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'integration tests',
                query: 'patternType:regexp case:yes \\b(it|test)\\( f:/integration/.*\\.test\\.ts$ archived:no fork:no',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const TS_VS_GO: Template = {
    type: InsightType.SearchBased,
    title: 'Typescript vs. Go',
    description: 'Are there more Typescript or more Go files',
    templateValues: {
        title: 'Typescript vs. Go',
        allRepos: true,
        series: [
            {
                name: 'TypeScript',
                query: 'select:file lang:TypeScript archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'GO lang',
                query: 'select:file lang:Go archived:no fork:no',
                stroke: DATA_SERIES_COLORS.INDIGO,
            },
        ],
    },
}

const IOS_APP_SCREENS: Template = {
    type: InsightType.SearchBased,
    title: 'iOS app screens',
    description: 'What number of iOS app screens are in the entire app',
    templateValues: {
        title: 'iOS app screens',
        allRepos: true,
        series: [
            {
                name: 'Screens',
                query: 'struct\\s(.*):\\sview$ patternType:regexp lang:swift archived:no fork:no',
                stroke: DATA_SERIES_COLORS.YELLOW,
            },
        ],
    },
}

const ADOPTING_NEW_API: Template = {
    type: InsightType.SearchBased,
    title: 'Adopting new API by Team',
    description: 'Which teams or repos have adopted a new API so far',
    templateValues: {
        title: 'Adopting new API by Team',
        allRepos: true,
        series: [
            {
                name: 'Mobile team',
                query: 'file:mobileTeam newAPI.call archived:no fork:no',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'Web team',
                query: 'file:webappTeam newAPI.call archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
        ],
    },
}

const PROBLEMATIC_API_BY_TEAM: Template = {
    type: InsightType.SearchBased,
    title: 'Problematic API by Team',
    description: 'Which teams have the most usage of a problematic API',
    templateValues: {
        title: 'Problematic API by Team',
        allRepos: true,
        series: [
            {
                name: 'Mobile team',
                query: 'problemAPI file:teamOneDirectory archived:no fork:no',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'Web team',
                query: 'problemAPI file:teamTwoDirectory archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
        ],
    },
}

const DATA_FETCHING_GQL: Template = {
    type: InsightType.SearchBased,
    title: 'Data fetching from GraphQL',
    description: 'What GraphQL operations are being called often',
    templateValues: {
        title: 'Data fetching from GraphQL',
        allRepos: true,
        series: [
            {
                name: 'requestGraphQL',
                query: 'patternType:regexp requestGraphQL(\\(|<[^>]*>\\() archived:no fork:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'Direct query/mutate calls',
                query: 'patternType:regexp (query|mutate)GraphQL(\\(|<[^>]*>\\() archived:no fork:no',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'Hooks',
                query:
                    'patternType:regexp use(Query|Mutation|Connection|LazyQuery)(\\(|<[^>]*>\\() archived:no fork:no',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

export const TEMPLATE_SECTIONS: TemplateSection[] = [
    {
        title: 'Popular',
        templates: [
            TERRAFORM_VERSIONS,
            CSS_MODULES_MIGRATION,
            LOG4J_FIXED_VERSIONS,
            YARN_ADOPTION,
            JAVA_VERSIONS,
            LINTER_OVERRIDE_RULES,
            TS_JS_USAGE,
        ],
    },
    {
        title: 'Migration',
        templates: [
            CONFIG_OR_DOC_FILE,
            ALLOW_DENY_LIST_TRACKING,
            CSS_MODULES_MIGRATION,
            PYTHON_2_3,
            REACT_FUNCTION_CLASS,
        ],
    },
    {
        title: 'Adoption',
        templates: [
            NEW_API_USAGE,
            YARN_ADOPTION,
            FREQUENTLY_USED_DATABASE,
            LARGE_PACKAGE_USAGE,
            REACT_COMPONENT_LIB_USAGE,
            CI_TOOLING,
        ],
    },
    {
        title: 'Deprecation',
        templates: [
            CSS_CLASS,
            ICON_OR_IMAGE,
            STRUCTURAL_CODE_PATTERN,
            TOOLING_MIGRATION,
            VAR_KEYWORDS,
            TESTING_LIBRARIES,
        ],
    },
    {
        title: 'Versions and patterns',
        templates: [
            JAVA_VERSIONS,
            LICENSE_TYPES,
            ALL_LOG4J_VERSIONS,
            PYTHON_VERSIONS,
            NODEJS_VERSIONS,
            CSS_COLORS,
            CHECKOV_SKIP_TYPES,
        ],
    },
    {
        title: 'Code health',
        templates: [
            TODOS,
            LINTER_OVERRIDE_RULES,
            REVERT_COMMITS,
            DEPRECATED_CALLS,
            STORYBOOK_TESTS,
            REPOS_WITH_README,
            OWNERSHIP_TRACKING,
            CI_TOOLING,
        ],
    },
    {
        title: 'Security',
        templates: [
            VULNERABLE_OPEN_SOURCE,
            API_KEYS_DETECTION,
            LOG4J_FIXED_VERSIONS,
            SKIPPED_TESTS,
            TEST_AMOUNT_AND_TYPES,
        ],
    },
    {
        title: 'Other',
        templates: [TS_VS_GO, IOS_APP_SCREENS, ADOPTING_NEW_API, PROBLEMATIC_API_BY_TEAM, DATA_FETCHING_GQL],
    },
]
