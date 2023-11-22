import { useApolloClient } from '@apollo/client'
import type { MockedResponse } from '@apollo/client/testing'
import * as H from 'history'

import { getDocumentNode } from '@sourcegraph/http-client'
import type { Settings } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import {
    type FileNamesResult,
    type FuzzyFinderRepoResult,
    type FuzzyFinderSymbolsResult,
    SymbolKind,
} from '../../graphql-operations'
import { UserHistory } from '../useUserHistory'

import { FUZZY_GIT_LSFILES_QUERY } from './FuzzyFiles'
import { FuzzyFinderContainer } from './FuzzyFinder'
import { FUZZY_REPOS_QUERY } from './FuzzyRepos'
import { FUZZY_SYMBOLS_QUERY } from './FuzzySymbols'
import type { FuzzyTabKey } from './FuzzyTabs'

export interface FuzzyWrapperProps {
    url: string
    experimentalFeatures: Settings['experimentalFeatures']
    initialQuery?: string
    activeTab?: FuzzyTabKey
}

export const FuzzyWrapper: React.FunctionComponent<FuzzyWrapperProps> = props => {
    const history = H.createMemoryHistory()
    history.push(props.url)
    const client = useApolloClient()
    return (
        <FuzzyFinderContainer
            defaultActiveTab={props.activeTab}
            client={client}
            isVisible={true}
            setIsVisible={() => {}}
            isRepositoryRelatedPage={true}
            location={history.location}
            settingsCascade={{ final: { experimentalFeatures: props.experimentalFeatures }, subjects: null }}
            telemetryService={NOOP_TELEMETRY_SERVICE}
            initialQuery={props.initialQuery}
            userHistory={new UserHistory()}
        />
    )
}

export const FUZZY_FILES_MOCK: MockedResponse<FileNamesResult> = {
    request: {
        query: getDocumentNode(FUZZY_GIT_LSFILES_QUERY),
        variables: { repository: 'github.com/sourcegraph/sourcegraph', commit: 'main' },
    },
    result: {
        data: {
            repository: {
                id: '',
                commit: {
                    id: '',
                    fileNames: [
                        'babel.config.js',
                        'client/README.md',
                        'client/branded/.eslintignore',
                        'client/branded/.eslintrc.js',
                        // This line is intentionally long to test what happens during horizontal overflows
                        'client/branded/.src/components/BrandedStory.tsx/client/branded/srcndedStory.tsx/client/branded/src/components/BrandedStory.tsx/client/branded/src/components/BrandedStory.tsx',
                        'client/branded/.stylelintrc.json',
                        'client/branded/README.md',
                        'client/branded/jest.config.js',
                        'client/branded/package.json',
                        'client/branded/src/components/CodeSnippet.tsx',
                        'client/branded/src/components/Form.tsx',
                        'client/branded/src/components/LoaderInput.scss',
                        'client/branded/src/components/LoaderInput.story.tsx',
                    ],
                },
            },
        },
    },
}

export const FUZZY_REPOS_MOCK: MockedResponse<FuzzyFinderRepoResult> = {
    request: {
        query: getDocumentNode(FUZZY_REPOS_QUERY),
        variables: { query: 'type:repo count:10' },
    },
    result: {
        data: {
            search: {
                results: {
                    repositories: [{ name: 'github.com/sourcegraph/sourcegraph', stars: 1234 }],
                },
            },
        },
    },
}

export const FUZZY_SYMBOLS_MOCK: MockedResponse<FuzzyFinderSymbolsResult> = {
    request: {
        query: getDocumentNode(FUZZY_SYMBOLS_QUERY),
        variables: { query: 'repo:^github\\.com/sourcegraph/sourcegraph$@main type:symbol count:10' },
    },
    result: {
        data: {
            search: {
                results: {
                    results: [
                        {
                            __typename: 'FileMatch',
                            repository: { name: 'github.com/sourcegraph/sourcegraph' },
                            file: { path: 'path.go' },
                            symbols: [
                                {
                                    url: '/path.go?L10',
                                    containerName: 'pkg',
                                    name: 'hello',
                                    language: 'Go',
                                    kind: SymbolKind.CLASS,
                                },
                            ],
                        },
                    ],
                },
            },
        },
    },
}
