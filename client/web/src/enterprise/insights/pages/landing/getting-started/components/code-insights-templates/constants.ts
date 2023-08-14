import { DATA_SERIES_COLORS } from '../../../../../constants'
import { InsightType } from '../../../../../core'
import type { CaptureInsightUrlValues } from '../../../../insights/creation/capture-group'
import type { SearchInsightURLValues } from '../../../../insights/creation/search-insight'

export interface TemplateSection {
    title: string
    templates: Template[]
    experimental?: boolean
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
        repoQuery: 'repo:has.file(path:\\.tf$)',
        series: [
            {
                name: '1.1.0',
                query: 'app.terraform.io/(.*)\\n version =(.*)1.1.0 patternType:regexp lang:Terraform',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: '1.2.0',
                query: 'app.terraform.io/(.*)\\n version =(.*)1.2.0 patternType:regexp lang:Terraform',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Global CSS',
                query: 'select:file lang:SCSS -file:module patterntype:regexp',
                stroke: DATA_SERIES_COLORS.RED,
            },
            {
                name: 'CSS Modules',
                query: 'select:file lang:SCSS file:module patterntype:regexp',
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
        repoQuery: 'repo:has.file(path:\\.gradle$)',
        series: [
            {
                name: 'Vulnerable',
                query: 'lang:gradle org\\.apache\\.logging\\.log4j[\'"] 2\\.(0|1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16)(\\.[0-9]+) patterntype:regexp',
                stroke: DATA_SERIES_COLORS.RED,
            },
            {
                name: 'Fixed',
                query: 'lang:gradle org\\.apache\\.logging\\.log4j[\'"] 2\\.(17)(\\.[0-9]+) patterntype:regexp',
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
        repoQuery: 'repo:has.file(path:yarn\\.lock$)',
        series: [
            {
                name: 'Yarn',
                query: 'select:repo file:yarn.lock',
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
        repoQuery: 'repo:has.file(path:pom\\.xml$)',
        groupSearchQuery: 'file:pom\\.xml$ <java\\.version>(.*)</java\\.version>',
    },
}

const LINTER_OVERRIDE_RULES: Template = {
    type: InsightType.SearchBased,
    title: 'Linter override rules',
    description: 'A code health indicator for how many linter override rules exist',
    templateValues: {
        title: 'Linter override rules',
        repoQuery: 'repo:has.file(path:\\.eslintignore$)',
        series: [
            {
                name: 'Rule overrides',
                query: 'file:^\\.eslintignore ^[^#].*.\\n patternType:regexp',
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
        repoQuery: 'repo:has.file(path:\\.ts$) or repo:has.file(path:\\.js$)',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Repositories with doc',
                query: 'select:repo file:docs/*/new_config_filename',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'blacklist/whitelist',
                query: 'select:file blacklist OR whitelist',
                stroke: DATA_SERIES_COLORS.RED,
            },
            {
                name: 'denylist/allowlist',
                query: 'select:file denylist OR allowlist',
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
        repoQuery: 'repo:has.file(path:\\.py$)',
        series: [
            {
                name: 'Python 3',
                query: '#!/usr/bin/env python3',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'Python 2',
                query: '#!/usr/bin/env python2',
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
                query: 'select:repo ourApiLibraryName.load',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Redis',
                query: 'redis\\.set patternType:regexp',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'GraphQL',
                query: 'graphql\\( patternType:regexp',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Repositories with large package usage',
                query: 'select:repo import\\slargePkg patternType:regexp',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Library imports',
                query: "from '@sourceLibrary/component' patternType:literal",
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
        repoQuery: 'repo:has.file(path:\\.circleci/config\\.yml)',
        series: [
            {
                name: 'Repo with CircleCI config',
                query: 'file:\\.circleci/config\\.yml select:repo',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Deprecated CSS class',
                query: 'deprecated-class',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Deprecated logo',
                query: '2018logo.png',
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
        repoQuery: 'repo:has.file(path:\\.java$)',
        series: [
            {
                name: 'Try catch',
                query: 'try {:[_]} catch (:[e]) { } finally {:[_]} lang:java patternType:structural',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Deprecated logger',
                query: 'deprecatedEventLogger.log',
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
        repoQuery: 'repo:has.file(path:\\.ts$) or repo:has.file(path:\\.js$)',
        series: [
            {
                name: 'var statements',
                query: '(lang:TypeScript OR lang:JavaScript) var ... = patterntype:structural',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: '@testing-library',
                query: "from '@testing-library/react'",
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'enzyme',
                query: "from 'enzyme'",
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
        repoQuery: 'repo:has.file(path:package\\.json$)',
        groupSearchQuery: 'file:package.json "license":\\s"(.*)"',
    },
}

const ALL_LOG4J_VERSIONS: Template = {
    type: InsightType.CaptureGroup,
    title: 'All log4j versions',
    description: 'Which log4j versions are present, including vulnerable versions',
    templateValues: {
        title: 'All log4j versions',
        repoQuery: 'repo:has.file(path:\\.gradle$)',
        groupSearchQuery: 'lang:gradle org\\.apache\\.logging\\.log4j[\'"] 2\\.([0-9]+)\\.',
    },
}

const PYTHON_VERSIONS: Template = {
    type: InsightType.CaptureGroup,
    title: 'Python versions',
    description: 'Which python versions are in use or haven’t been updated',
    templateValues: {
        title: 'Python versions',
        repoQuery: 'repo:.*',
        groupSearchQuery: '#!/usr/bin/env python([0-9]\\.[0-9]+)',
    },
}

const NODEJS_VERSIONS: Template = {
    type: InsightType.CaptureGroup,
    title: 'Node.js versions',
    description: 'Which node.js versions are present based on nvm files',
    templateValues: {
        title: 'Node.js versions',
        repoQuery: 'repo:.*',
        groupSearchQuery: 'nvm\\suse\\s([0-9]+\\.[0-9]+)',
    },
}

const CSS_COLORS: Template = {
    type: InsightType.CaptureGroup,
    title: 'CSS Colors',
    description: 'What CSS colors are present or most popular',
    templateValues: {
        title: 'CSS Colors',
        repoQuery: 'repo:.*',
        groupSearchQuery: 'color:#([0-9a-fA-f]{3,6})',
    },
}

const CHECKOV_SKIP_TYPES: Template = {
    type: InsightType.CaptureGroup,
    title: 'Types of checkov skips',
    description: 'See the most common reasons for why secuirty checks in checkov are skipped',
    templateValues: {
        title: 'Types of checkov skips',
        repoQuery: 'repo:.*',
        groupSearchQuery: 'patterntype:regexp file:.tf #checkov:skip=(.*)',
    },
}

const TODOS: Template = {
    type: InsightType.SearchBased,
    title: 'TODOs',
    description: 'How many TODOs are in a specific part of the codebase (or all of it)',
    templateValues: {
        title: 'TODOs',
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'TODOs',
                query: 'TODO',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Reverts',
                query: 'type:commit revert',
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
        repoQuery: 'repo:has.file(path:\\.java$)',
        series: [
            {
                name: '@deprecated',
                query: 'lang:java @deprecated',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Stories',
                query: 'patternType:regexp f:\\.story\\.tsx$ \\badd\\(',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'with readme',
                query: 'repohasfile:readme.md select:repo',
                stroke: DATA_SERIES_COLORS.LIME,
            },
            {
                name: 'without readme',
                query: '-repohasfile:readme.md select:repo',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'with CODEOWNERS',
                query: 'repohasfile:CODEOWNERS select:repo',
                stroke: DATA_SERIES_COLORS.LIME,
            },
            {
                name: 'without CODEOWNERS',
                query: '-repohasfile:CODEOWNERS select:repo',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'vulnerableLibrary@14.3.9',
                query: 'vulnerableLibrary@14.3.9',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'API key',
                query: 'regexMatchingAPIKey patternType:regexp',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Skipped tests',
                query: '(this.skip() OR it.skip) lang:TypeScript',
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
        repoQuery: 'repo:has.file(path:\\.ts$)',
        series: [
            {
                name: 'e2e tests',
                query: 'patternType:regexp case:yes \\b(it|test)\\( f:/end-to-end/.*\\.test\\.ts$',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'regression tests',
                query: 'patternType:regexp case:yes \\b(it|test)\\( f:/regression/.*\\.test\\.ts$',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'integration tests',
                query: 'patternType:regexp case:yes \\b(it|test)\\( f:/integration/.*\\.test\\.ts$',
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
        repoQuery: 'repo:has.file(path:\\.ts$) or repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'TypeScript',
                query: 'select:file lang:TypeScript',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'GO lang',
                query: 'select:file lang:Go',
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
        repoQuery: 'repo:has.file(path:\\.swift$)',
        series: [
            {
                name: 'Screens',
                query: 'struct\\s(.*):\\sview$ patternType:regexp lang:swift',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Mobile team',
                query: 'file:mobileTeam newAPI.call',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'Web team',
                query: 'file:webappTeam newAPI.call',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'Mobile team',
                query: 'problemAPI file:teamOneDirectory',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'Web team',
                query: 'problemAPI file:teamTwoDirectory',
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
        repoQuery: 'repo:.*',
        series: [
            {
                name: 'requestGraphQL',
                query: 'patternType:regexp requestGraphQL(\\(|<[^>]*>\\()',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'Direct query/mutate calls',
                query: 'patternType:regexp (query|mutate)GraphQL(\\(|<[^>]*>\\()',
                stroke: DATA_SERIES_COLORS.BLUE,
            },
            {
                name: 'Hooks',
                query: 'patternType:regexp use(Query|Mutation|Connection|LazyQuery)(\\(|<[^>]*>\\()',
                stroke: DATA_SERIES_COLORS.ORANGE,
            },
        ],
    },
}

const GO_STATIC_CHECK_SA6005: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Inefficient string comparison with strings.ToLower or strings.ToUpper',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Inefficient string comparison with strings.ToLower or strings.ToUpper',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'SA6005',
                query: 'if strings.ToLower(:[1]) == strings.ToLower(:[2]) or if strings.ToUpper(:[1]) == strings.ToUpper(:[2]) or if strings.ToLower(:[1]) != strings.ToLower(:[2]) or if strings.ToUpper(:[1]) != strings.ToUpper(:[2]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'SA6005 - fixed',
                query: 'if strings.EqualFold(:[1], :[2]) or if !strings.EqualFold(:[1], :[2]) or if strings.EqualFold(:[1], :[2]) or if !strings.EqualFold(:[1], :[2]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1002: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Omit comparison with boolean constant',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Omit comparison with boolean constant',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1002',
                query: 'if :[1:e] == false or if :[1:e] == true patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1003: Template = {
    type: InsightType.SearchBased,
    title: 'Replace call to strings.Index with strings.Contains',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Replace call to strings.Index with strings.Contains',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1003',
                query: 'strings.Index(:[1], :[2]) < 0  or strings.Index(:[1], :[2]) == -1 or strings.Index(:[1], :[2]) != -1 or strings.Index(:[1], :[2]) >= 0 or strings.Index(:[1], :[2]) > -1 or strings.IndexAny(:[1], :[2]) < 0 or strings.IndexAny(:[1], :[2]) == -1 or strings.IndexAny(:[1], :[2]) != -1 or strings.IndexAny(:[1], :[2]) >= 0 or strings.IndexAny(:[1], :[2]) > -1 or strings.IndexRune(:[1], :[2]) < 0 or strings.IndexRune(:[1], :[2]) == -1 or strings.IndexRune(:[1], :[2]) != -1 or strings.IndexRune(:[1], :[2]) >= 0 or strings.IndexRune(:[1], :[2]) > -1 patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1003 - fixed',
                query: 'strings.ContainsRune(:[1], :[2]) or strings.ContainsRune(:[1], :[2]) or strings.ContainsRune(:[1], :[2]) or !strings.ContainsRune(:[1], :[2]) or !strings.ContainsRune(:[1], :[2]) or strings.ContainsAny(:[1], :[2]) or strings.ContainsAny(:[1], :[2]) or strings.ContainsAny(:[1], :[2]) or !strings.ContainsAny(:[1], :[2]) or !strings.ContainsAny(:[1], :[2]) or strings.Contains(:[1], :[2]) or !strings.Contains(:[1], :[2]) or !strings.Contains(:[1], :[2]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1004: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Replace call to bytes.Compare with bytes.Equal',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Replace call to bytes.Compare with bytes.Equal',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1004',
                query: 'bytes.Compare(:[1], :[2]) != 0 or bytes.Compare(:[1], :[2]) == 0 patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1004 - fixed',
                query: '!bytes.Equal(:[1], :[2]) or bytes.Equal(:[1], :[2]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1005: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Drop unnecessary use of the blank identifier',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Drop unnecessary use of the blank identifier',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1005',
                query: 'for :[1:e], :[~_] := range or for :[1:e], :[~_] = range or for :[~_] = range or for :[~_], :[~_] = range patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1005 - fixed',
                query: 'for range or for :[1] := range or for :[1] = range patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1006: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Use for { ... } for infinite loops',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Use for { ... } for infinite loops',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1006',
                query: 'for true {:[x]} patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1006 - fixed',
                query: 'for {:[x]} patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1010: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Omit default slice index',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Omit default slice index',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S10010',
                query: ':[s.][:len(:[s])] patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1010 - fixed',
                query: ':[s.][:] patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1012: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Replace time.Now().Sub(x) with time.Since(x)',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Replace time.Now().Sub(x) with time.Since(x)',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S10012',
                query: 'time.Now().Sub(:[x]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1012 - fixed',
                query: 'time.Since(:[x]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

// const GO_STATIC_CHECK_S1017: skipped
// const GO_STATIC_CHECK_S1018: skipped

const GO_STATIC_CHECK_S1019: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Simplify make call by omitting redundant arguments',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Simplify make call by omitting redundant arguments',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S10019',
                query: 'make(chan int, 0) or make(map[:[[1]]]:[[1]], 0) or make(:[1], :[2], :[2]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1019 - fixed',
                query: 'make(chan int) or make(map[:[[1]]]:[[1]]) or make(:[1], :[2]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1020: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Omit redundant nil check in type assertion',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Omit redundant nil check in type assertion',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1020',
                query: 'if :[_.], ok := :[i.].(:[T]); :[i.] != nil && ok {:[body]} or if :[_.], ok := :[i.].(:[T]); ok && :[i.] != nil {:[body]} or if :[i.] != nil {  if :[_.], ok := :[i.].(:[T]); ok {:[body]}} patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1020 - fixed',
                query: 'if :[_.], ok := :[i.].(:[T]); ok {:[body]} patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1023: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Omit redundant control flow',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Omit redundant control flow',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1023',
                query: 'func() {:[body] return } or func :[fn.](:[args]) {:[body] return } patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1023 - fixed',
                query: 'func() {:[body]} or func :[fn.](:[args]) {:[body]} patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1024: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Replace x.Sub(time.Now()) with time.Until(x)',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Replace x.Sub(time.Now()) with time.Until(x)',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1024',
                query: ':[x.].Sub(time.Now()) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1024 - fixed',
                query: 'func() {:[body]} or func :[fn.](:[args]) {:[body]} patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1025: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Don’t use fmt.Sprintf("%s", x) unnecessarily',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Don’t use fmt.Sprintf("%s", x) unnecessarily',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1025',
                query: 'fmt.Println("%s", ":[s]") patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1025 - fixed',
                query: 'fmt.Println(":[s]") patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1028: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Simplify error construction with fmt.Errorf',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Simplify error construction with fmt.Errorf',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1028',
                query: 'errors.New(fmt.Sprintf(:[1])) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1028 - fixed',
                query: 'fmt.Errorf(:[1]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1029: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Range over the string directly',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Range over the string directly',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1029',
                query: 'for :[~_], :[r.] := range []rune(:[s.]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1029 - fixed',
                query: 'for _, :[r] := range :[s] patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1032: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Use sort.Ints(x), sort.Float64s(x), and sort.Strings(x)',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Use sort.Ints(x), sort.Float64s(x), and sort.Strings(x)',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1032',
                query: 'sort.Float64s(:[1]) or sort.Strings(:[1]) or sort.Ints(:[1]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1035: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1035',
                query: 'headers.Set(http.CanonicalHeaderKey(:[1])) or headers.Get(http.CanonicalHeaderKey(:[1])) or headers.Del(http.CanonicalHeaderKey(:[1])) or headers.Add(http.CanonicalHeaderKey(:[1]), :[1]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1035 - fixed',
                query: 'headers.Set(:[1]) or headers.Get(:[1]) or headers.Del(:[1]) or headers.Add(:[1], :[1]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

// const GO_STATIC_CHECK_S1036: skipped

const GO_STATIC_CHECK_S1037: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Redundant call to net/http.CanonicalHeaderKey in method call on net/http.Header',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1035',
                query: 'headers.Set(http.CanonicalHeaderKey(:[1])) or headers.Get(http.CanonicalHeaderKey(:[1])) or headers.Del(http.CanonicalHeaderKey(:[1])) or headers.Add(http.CanonicalHeaderKey(:[1]), :[1]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1037 - fixed',
                query: 'time.Sleep(:[t]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1038: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Unnecessarily complex way of printing formatted string',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] - Unnecessarily complex way of printing formatted string',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1038',
                query: 'select {	case <-time.After(:[t]):} patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1038 - fixed',
                query: 'time.Sleep(:[t]) patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

const GO_STATIC_CHECK_S1039: Template = {
    type: InsightType.SearchBased,
    title: '[quickfix] - Unnecessary use of fmt.Sprint',
    description: 'Code search turned code checker',
    templateValues: {
        title: '[quickfix] Unnecessary use of fmt.Sprint',
        repoQuery: 'repo:has.file(path:\\.go$)',
        series: [
            {
                name: 'S1039',
                query: 'fmt.Sprintf("%s", ":[s]") patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GRAPE,
            },
            {
                name: 'S1025 - fixed',
                query: '":[s]" patternType:structural archived:no',
                stroke: DATA_SERIES_COLORS.GREEN,
            },
        ],
    },
}

export const getTemplateSections = (goCodeCheckerTemplates: boolean | undefined): TemplateSection[] => {
    const allButGoChecker: TemplateSection[] = [
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

    if (!goCodeCheckerTemplates) {
        return allButGoChecker
    }

    const all = [...allButGoChecker]

    all.splice(-1, 0, {
        title: 'Go code checker',
        experimental: true,
        templates: [
            GO_STATIC_CHECK_SA6005,
            GO_STATIC_CHECK_S1002,
            GO_STATIC_CHECK_S1003,
            GO_STATIC_CHECK_S1004,
            GO_STATIC_CHECK_S1005,
            GO_STATIC_CHECK_S1006,
            GO_STATIC_CHECK_S1010,
            GO_STATIC_CHECK_S1012,
            GO_STATIC_CHECK_S1019,
            GO_STATIC_CHECK_S1020,
            GO_STATIC_CHECK_S1023,
            GO_STATIC_CHECK_S1024,
            GO_STATIC_CHECK_S1025,
            GO_STATIC_CHECK_S1028,
            GO_STATIC_CHECK_S1029,
            GO_STATIC_CHECK_S1032,
            GO_STATIC_CHECK_S1035,
            GO_STATIC_CHECK_S1037,
            GO_STATIC_CHECK_S1038,
            GO_STATIC_CHECK_S1039,
        ],
    })

    return all
}
