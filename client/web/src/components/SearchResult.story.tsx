import { storiesOf } from '@storybook/react'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React from 'react'

import { SearchResult } from './SearchResult'
import { WebStory } from './WebStory'

const defaultProps: Omit<React.ComponentProps<typeof SearchResult>, 'result' | 'icon'> = {
    isLightTheme: true,
    repoName: 'a/b',
}

const { add } = storiesOf('web/search/results/SearchResult', module).addParameters({
    chromatic: { viewports: [800] },
})

add('repository/source', () => (
    <WebStory>
        {() => (
            <SearchResult {...defaultProps} icon={SourceRepositoryIcon} result={{ type: 'repo', repository: 'a/b' }} />
        )}
    </WebStory>
))

add('repository/github', () => (
    <WebStory>
        {() => (
            <SearchResult
                {...defaultProps}
                icon={SourceRepositoryIcon}
                repoName="github.com/a/b"
                result={{ type: 'repo', repository: 'github.com/a/b' }}
            />
        )}
    </WebStory>
))

add('repository/github stars', () => (
    <WebStory>
        {() => (
            <SearchResult
                {...defaultProps}
                icon={SourceRepositoryIcon}
                repoName="github.com/a/b"
                result={{ type: 'repo', repository: 'github.com/a/b', repoStars: 123 }}
            />
        )}
    </WebStory>
))

add('repository/fork', () => (
    <WebStory>
        {() => (
            <SearchResult
                {...defaultProps}
                icon={SourceRepositoryIcon}
                result={{ type: 'repo', repository: 'a/b', fork: true }}
            />
        )}
    </WebStory>
))

add('repository/archived', () => (
    <WebStory>
        {() => (
            <SearchResult
                {...defaultProps}
                icon={SourceRepositoryIcon}
                result={{ type: 'repo', repository: 'a/b', archived: true }}
            />
        )}
    </WebStory>
))

add('repository/fork archived', () => (
    <WebStory>
        {() => (
            <SearchResult
                {...defaultProps}
                icon={SourceRepositoryIcon}
                result={{ type: 'repo', repository: 'a/b', fork: true, archived: true }}
            />
        )}
    </WebStory>
))

add('repository/stars fork archived', () => (
    <WebStory>
        {() => (
            <SearchResult
                {...defaultProps}
                icon={SourceRepositoryIcon}
                result={{ type: 'repo', repository: 'a/b', repoStars: 123, fork: true, archived: true }}
            />
        )}
    </WebStory>
))

add('commit', () => (
    <WebStory>
        {() => (
            <SearchResult
                {...defaultProps}
                icon={SourceCommitIcon}
                result={{
                    type: 'commit',
                    label: 'title',
                    detail: 'detail',
                    url: '/u',
                    content: 'abc',
                    ranges: [],
                    repository: 'a/b',
                }}
            />
        )}
    </WebStory>
))
