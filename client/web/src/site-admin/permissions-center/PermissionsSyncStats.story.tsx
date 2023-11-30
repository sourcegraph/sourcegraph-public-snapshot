import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { PERMISSIONS_SYNC_JOBS_STATS } from './backend'
import { PermissionsSyncStats } from './PermissionsSyncStats'

const decorator: Decorator = Story => <Story />

const config: Meta = {
    title: 'web/src/site-admin/permissions-center/PermissionsSyncStats',
    decorators: [decorator],
}

export default config

export const Stats: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(PERMISSIONS_SYNC_JOBS_STATS),
                                variables: {},
                            },
                            result: {
                                data: {
                                    permissionsSyncingStats: {
                                        queueSize: 1337,
                                        usersWithLatestJobFailing: 228101,
                                        reposWithLatestJobFailing: 3,
                                        usersWithNoPermissions: 4,
                                        reposWithNoPermissions: 5,
                                        usersWithStalePermissions: 6,
                                        reposWithStalePermissions: 42,
                                    },
                                },
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                    ])
                }
            >
                <PermissionsSyncStats
                    polling={false}
                    filters={{
                        reason: '',
                        state: '',
                        searchType: '',
                        query: '',
                    }}
                    setFilters={() => null}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

Stats.storyName = 'Permissions sync statistics'
