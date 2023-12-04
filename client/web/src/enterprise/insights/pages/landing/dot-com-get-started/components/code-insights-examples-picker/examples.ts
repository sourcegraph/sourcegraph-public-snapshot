import { InsightType } from '../../../../../core'
import {
    ALPINE_VERSIONS_INSIGHT,
    CSS_MODULES_VS_GLOBAL_STYLES_INSIGHT,
    DEPRECATED_API_USAGE_BY_TEAM,
    LINTER_OVERRIDES,
    LOG_4_J_INCIDENT_INSIGHT,
    OPENSSL_PYTHON,
    REPOS_WITH_CI_SYSTEM,
} from '../../../getting-started/components/code-insights-examples/examples'
import type {
    CaptureGroupExampleContent,
    SearchInsightExampleContent,
} from '../../../getting-started/components/code-insights-examples/types'

interface SearchInsightExample {
    type: InsightType.SearchBased
    content: SearchInsightExampleContent<any>
}

interface CaptureGroupInsightExample {
    type: InsightType.CaptureGroup
    content: CaptureGroupExampleContent<any>
}

type Example = (SearchInsightExample | CaptureGroupInsightExample) & {
    description: string
}

export const EXAMPLES: Example[] = [
    {
        type: InsightType.SearchBased,
        content: CSS_MODULES_VS_GLOBAL_STYLES_INSIGHT,
        description: 'Track migrations, adoption, and deprecations',
    },
    {
        type: InsightType.CaptureGroup,
        content: ALPINE_VERSIONS_INSIGHT,
        description: 'Detect and track versions of languages or packages',
    },
    {
        type: InsightType.SearchBased,
        content: LOG_4_J_INCIDENT_INSIGHT,
        description: 'Ensure removal of security vulnerabilities',
    },
    {
        type: InsightType.SearchBased,
        content: OPENSSL_PYTHON,
        description: 'Find vulnerable OpenSSL versions in the Python Ecosystem',
    },
    {
        type: InsightType.SearchBased,
        content: DEPRECATED_API_USAGE_BY_TEAM,
        description: 'Understand code by team',
    },
    {
        type: InsightType.SearchBased,
        content: LINTER_OVERRIDES,
        description: 'Track code smells and health',
    },
    {
        type: InsightType.SearchBased,
        content: REPOS_WITH_CI_SYSTEM,
        description: 'Visualize configurations and services',
    },
]
