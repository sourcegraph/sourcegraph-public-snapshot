import { gql, useQuery } from '@sourcegraph/http-client'
import { TourIcon, type TourTaskType } from '@sourcegraph/shared/src/settings/temporary'

import type { OnboardingTourConfigResult, OnboardingTourConfigVariables } from '../../graphql-operations'

export const ONBOARDING_TOUR_QUERY = gql`
    query OnboardingTourConfig {
        onboardingTourContent {
            current {
                id
                value
            }
        }
    }
`

export const ONBOARDING_TOUR_MUTATION = gql`
    mutation OnboardingTourConfigMutation($json: String!) {
        updateOnboardingTourContent(input: $json) {
            alwaysNil
        }
    }
`

export interface TourConfig {
    tasks: TourTaskType[]
    defaultSnippets: Record<string, string[]>
}

export function parseTourConfig(json: string): TourConfig {
    return JSON.parse(json)
}

/**
 * Returns the configured or default tasks and snippets for the user onboarding tour.
 */
export const useOnboardingTasks = (): { loading: boolean; error?: Error; data?: TourConfig } => {
    const { data, loading, error } = useQuery<OnboardingTourConfigResult, OnboardingTourConfigVariables>(
        ONBOARDING_TOUR_QUERY,
        {}
    )

    let config: TourConfig | undefined
    if (!loading && !error) {
        if (data?.onboardingTourContent.current?.value) {
            try {
                config = parseTourConfig(data.onboardingTourContent.current.value)
            } catch (error) {
                return {
                    loading,
                    error,
                }
            }
        } else {
            config = {
                tasks: authenticatedTasks,
                defaultSnippets,
            }
        }
    }

    if (config && !config?.defaultSnippets) {
        config.defaultSnippets = defaultSnippets
    }

    return {
        loading,
        error,
        data: config,
    }
}

/**
 * Tour tasks for authenticated users. Extended/all use-cases.
 */
export const authenticatedTasks: TourTaskType[] = [
    {
        title: 'Code search with filters',
        icon: TourIcon.Search,
        steps: [
            {
                id: 'CodeSearch',
                label: 'Search all orgs or repositories matching a name for a literal code snippet',
                action: {
                    type: 'search-query',
                    query: 'repo:$$userrepo lang:$$userlang $$snippet',
                },
            },
        ],
    },
    {
        title: 'Commit search',
        icon: TourIcon.Search,
        steps: [
            {
                id: 'CommitSearch',
                label: 'Search commit titles and messages with-in a specific organization and repository',
                action: {
                    type: 'search-query',
                    query: 'repo:$$userrepo lang:$$userlang type:commit before:"last week"',
                },
            },
        ],
    },
    {
        title: 'Diff search',
        icon: TourIcon.Search,
        steps: [
            {
                id: 'DiffSearch',
                label: 'Search diffs for changes in code via filters like before, after, and author',
                action: {
                    type: 'search-query',
                    query: 'repo:$$userrepo type:diff after:"last month" $$snippet',
                },
            },
        ],
    },
    {
        title: 'Try Cody, our AI coding assistant',
        icon: TourIcon.Cody,
        steps: [
            {
                id: 'CodyVSCode',
                label: 'Install for VS Code',
                action: {
                    type: 'new-tab-link',
                    value: 'https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai',
                },
            },
            {
                id: 'CodyJetbrains',
                label: 'Install for JetBrains',
                action: {
                    type: 'new-tab-link',
                    value: 'https://plugins.jetbrains.com/plugin/9682-cody-ai-by-sourcegraph',
                },
            },
            {
                id: 'CodyWeb',
                label: 'Try Cody in Sourcegraph',
                action: {
                    type: 'new-tab-link',
                    value: '/cody/chat',
                },
            },
        ],
        requiredSteps: 1,
    },
    {
        title: 'Improve PRs With Sourcegraph',
        icon: TourIcon.Extension,
        steps: [
            {
                id: 'BrowserExtensions',
                label: 'Install the browser extension and leverage Sourcegraph code intelligence in reviews',
                action: {
                    type: 'new-tab-link',
                    value: 'https://docs.sourcegraph.com/integration/browser_extension',
                },
            },
        ],
    },
]

/**
 * Tour extra tasks for authenticated users.
 */
export const authenticatedExtraTask: TourTaskType = {
    title: 'All done!',
    icon: TourIcon.Check,
    steps: [
        {
            id: 'RestartTour',
            label: 'You can restart the tour to go through the steps again.',
            action: { type: 'restart', value: 'Restart tour' },
        },
    ],
}

export const defaultSnippets: Record<string, string[]> = {
    Go: ['stuct {', 'interface {', 'func ('],
    Java: ['import static', 'synchronized (', 'content:"class "'],
    Python: ['content:"with "', 'print(', 'content:"def "'],
    'C++': ['content:"class "', 'using namespace', '#include'],
    'C#': ['content:"class "', 'public void', '=>'],
    JavaScript: ['content:"class "', 'async function', '=>'],
    TypeScript: ['content:"class "', ': void', '=>'],
    PHP: ['content:"class "', 'fn(', '(Exception'],
    Ruby: ['undef', 'class <<', '.each do'],
    C: ['switch(', 'static void', 'if(', '){'],
    '*': ['todo OR fixme'],
}
