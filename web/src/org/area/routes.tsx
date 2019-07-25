import React from 'react'
import { Redirect } from 'react-router'
import { namespaceAreaRoutes } from '../../namespaces/routes'
import { lazyComponent } from '../../util/lazyComponent'
import { OrgAreaRoute } from './OrgArea'

const OrgSettingsArea = lazyComponent(() => import('../settings/OrgSettingsArea'), 'OrgSettingsArea')

export const orgAreaRoutes: ReadonlyArray<OrgAreaRoute> = [
    {
        path: '',
        exact: true,
        render: lazyComponent(() => import('./OrgOverviewPage'), 'OrgOverviewPage'),
    },
    {
        path: '/members',
        render: lazyComponent(() => import('./OrgMembersPage'), 'OrgMembersPage'),
    },
    {
        path: '/settings',
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <OrgSettingsArea {...props} isLightTheme={props.isLightTheme} />,
    },
    {
        path: '/searches',
        exact: true,
        render: lazyComponent(() => import('../saved-searches/OrgSavedSearchListPage'), 'OrgSavedSearchListPage'),
    },
    {
        path: '/searches/add',
        render: lazyComponent(
            () => import('../saved-searches/OrgSavedSearchesCreateForm'),
            'OrgSavedSearchesCreateForm'
        ),
    },
    {
        path: '/searches/:id',
        render: lazyComponent(
            () => import('../saved-searches/OrgSavedSearchesUpdateForm'),
            'OrgSavedSearchesUpdateForm'
        ),
    },
    ...namespaceAreaRoutes,

    // Redirect from previous /orgs/:orgname/account -> /orgs/:orgname/settings/profile.
    {
        path: '/account',
        render: props => <Redirect to={`${props.match.url}/settings/profile`} />,
    },
]
