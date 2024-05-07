import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { ALL_PERMISSIONS, ROLES_QUERY, DELETE_ROLE, SET_PERMISSIONS } from './backend'
import { mockPermissions, mockRoles } from './mock'
import { SiteAdminRolesPage } from './SiteAdminRolesPage'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/src/site-admin/rbac',
    decorators: [decorator],
}

export default config

const mocks = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(ALL_PERMISSIONS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockPermissions },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(ROLES_QUERY),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockRoles },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(DELETE_ROLE),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { deleteRole: { alwaysNil: null } } },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(SET_PERMISSIONS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { setPermissions: { alwaysNil: null } } },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

export const RolesPage: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={mocks}>
                <SiteAdminRolesPage telemetryRecorder={noOpTelemetryRecorder} />
            </MockedTestProvider>
        )}
    </WebStory>
)

RolesPage.storyName = 'Role Management Page'
