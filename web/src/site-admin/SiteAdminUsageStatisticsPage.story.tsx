import { SiteAdminUsageStatisticsPage, UsageStatistics } from './SiteAdminUsageStatisticsPage'
import { of, NEVER, throwError } from 'rxjs'
import { storiesOf } from '@storybook/react'
import { SuiteFunction } from 'mocha'
import * as H from 'history'
import React from 'react'
import webStyles from '../SourcegraphWebApp.scss'
import { IUserConnection } from '../../../shared/src/graphql/schema'

window.context = {} as SourcegraphContext & SuiteFunction

const { add } = storiesOf('web/SiteAdminUsageStatistics', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="theme-light container">{story()}</div>
    </>
))

const history = H.createMemoryHistory()

const commonProps = {
    history,
    location: history.location,
    match: {
        params: { id: '' },
        isExact: true,
        path: '',
        url: '',
    },
    now: () => new Date('2020-06-15T15:25:00+00:00'),
    isLightTheme: false,
}

// eslint-disable-next-line @typescript-eslint/consistent-type-assertions
const usageStatistics = {
    userCount: 1240,
    repositoryCount: 100128,
    mergedCampaignChangesets: 42,
    daus: [
        {
            userCount: 2,
            registeredUserCount: 2,
            anonymousUserCount: 0,
            searchActionCount: 12,
            searchUserCount: 1,
            startTime: '2020-07-03T00:00:00Z',
            codeIntelligenceUserCount: 0,
            codeIntelligenceActionCount: 19,
            integrationUserCount: 1,
            integrationActionCount: 19,
        },
        {
            userCount: 5,
            registeredUserCount: 5,
            anonymousUserCount: 0,
            startTime: '2020-07-02T00:00:00Z',
            searchUserCount: 1,
            searchActionCount: 30,
            codeIntelligenceUserCount: 4,
            codeIntelligenceActionCount: 44,
            integrationUserCount: 0,
            integrationActionCount: 17,
        },
        {
            userCount: 7,
            registeredUserCount: 7,
            anonymousUserCount: 0,
            startTime: '2020-07-01T00:00:00Z',
            searchUserCount: 0,
            searchActionCount: 38,
            codeIntelligenceUserCount: 0,
            codeIntelligenceActionCount: 65,
            integrationUserCount: 6,
            integrationActionCount: 14,
        },
        {
            userCount: 9,
            registeredUserCount: 9,
            anonymousUserCount: 0,
            startTime: '2020-06-30T00:00:00Z',
            searchUserCount: 4,
            searchActionCount: 4,
            codeIntelligenceUserCount: 3,
            codeIntelligenceActionCount: 67,
            integrationUserCount: 6,
            integrationActionCount: 50,
        },
        {
            userCount: 10,
            registeredUserCount: 10,
            anonymousUserCount: 0,
            startTime: '2020-06-29T00:00:00Z',
            searchUserCount: 4,
            searchActionCount: 12,
            codeIntelligenceUserCount: 6,
            codeIntelligenceActionCount: 27,
            integrationUserCount: 7,
            integrationActionCount: 63,
        },
        {
            userCount: 0,
            registeredUserCount: 0,
            anonymousUserCount: 0,
            startTime: '2020-06-28T00:00:00Z',
            searchUserCount: 0,
            searchActionCount: 0,
            codeIntelligenceUserCount: 0,
            codeIntelligenceActionCount: 0,
            integrationUserCount: 0,
            integrationActionCount: 0,
        },
        {
            userCount: 1,
            registeredUserCount: 1,
            anonymousUserCount: 0,
            startTime: '2020-06-27T00:00:00Z',
            searchUserCount: 0,
            searchActionCount: 4,
            codeIntelligenceUserCount: 0,
            codeIntelligenceActionCount: 3,
            integrationUserCount: 0,
            integrationActionCount: 5,
        },
        {
            userCount: 4,
            registeredUserCount: 4,
            anonymousUserCount: 0,
            startTime: '2020-06-26T00:00:00Z',
            searchUserCount: 1,
            searchActionCount: 9,
            codeIntelligenceUserCount: 1,
            codeIntelligenceActionCount: 29,
            integrationUserCount: 2,
            integrationActionCount: 13,
        },
        {
            userCount: 3,
            registeredUserCount: 3,
            anonymousUserCount: 0,
            startTime: '2020-06-25T00:00:00Z',
            searchUserCount: 2,
            searchActionCount: 19,
            codeIntelligenceUserCount: 1,
            codeIntelligenceActionCount: 6,
            integrationUserCount: 0,
            integrationActionCount: 4,
        },
        {
            userCount: 2,
            registeredUserCount: 2,
            anonymousUserCount: 0,
            startTime: '2020-06-24T00:00:00Z',
            searchUserCount: 0,
            searchActionCount: 19,
            codeIntelligenceUserCount: 0,
            codeIntelligenceActionCount: 6,
            integrationUserCount: 0,
            integrationActionCount: 13,
        },
        {
            userCount: 2,
            registeredUserCount: 2,
            anonymousUserCount: 0,
            startTime: '2020-06-23T00:00:00Z',
            searchUserCount: 0,
            searchActionCount: 2,
            codeIntelligenceUserCount: 1,
            codeIntelligenceActionCount: 6,
            integrationUserCount: 1,
            integrationActionCount: 3,
        },
        {
            userCount: 7,
            registeredUserCount: 7,
            anonymousUserCount: 0,
            startTime: '2020-06-22T00:00:00Z',
            searchUserCount: 4,
            searchActionCount: 21,
            codeIntelligenceUserCount: 0,
            codeIntelligenceActionCount: 32,
            integrationUserCount: 6,
            integrationActionCount: 54,
        },
        {
            userCount: 1,
            registeredUserCount: 1,
            anonymousUserCount: 0,
            startTime: '2020-06-21T00:00:00Z',
            searchUserCount: 0,
            searchActionCount: 7,
            codeIntelligenceUserCount: 0,
            codeIntelligenceActionCount: 7,
            integrationUserCount: 0,
            integrationActionCount: 8,
        },
        {
            userCount: 1,
            registeredUserCount: 1,
            anonymousUserCount: 0,
            startTime: '2020-06-20T00:00:00Z',
            searchUserCount: 0,
            searchActionCount: 4,
            codeIntelligenceUserCount: 0,
            codeIntelligenceActionCount: 5,
            integrationUserCount: 0,
            integrationActionCount: 2,
        },
    ],
    waus: [
        {
            userCount: 20,
            registeredUserCount: 20,
            anonymousUserCount: 0,
            startTime: '2020-06-28T00:00:00Z',
            searchUserCount: 0,
            searchActionCount: 172,
            codeIntelligenceUserCount: 19,
            codeIntelligenceActionCount: 41,
            integrationUserCount: 18,
            integrationActionCount: 81,
        },
        {
            userCount: 11,
            registeredUserCount: 11,
            anonymousUserCount: 0,
            startTime: '2020-06-21T00:00:00Z',
            searchUserCount: 4,
            searchActionCount: 89,
            codeIntelligenceUserCount: 1,
            codeIntelligenceActionCount: 6,
            integrationUserCount: 8,
            integrationActionCount: 76,
        },
        {
            userCount: 12,
            registeredUserCount: 12,
            anonymousUserCount: 0,
            startTime: '2020-06-14T00:00:00Z',
            searchUserCount: 8,
            searchActionCount: 66,
            codeIntelligenceUserCount: 2,
            codeIntelligenceActionCount: 61,
            integrationUserCount: 4,
            integrationActionCount: 81,
        },
        {
            userCount: 13,
            registeredUserCount: 13,
            anonymousUserCount: 0,
            startTime: '2020-06-07T00:00:00Z',
            searchUserCount: 8,
            searchActionCount: 60,
            codeIntelligenceUserCount: 4,
            codeIntelligenceActionCount: 43,
            integrationUserCount: 3,
            integrationActionCount: 127,
        },
        {
            userCount: 10,
            registeredUserCount: 10,
            anonymousUserCount: 0,
            startTime: '2020-05-31T00:00:00Z',
            searchUserCount: 7,
            searchActionCount: 46,
            codeIntelligenceUserCount: 9,
            codeIntelligenceActionCount: 34,
            integrationUserCount: 0,
            integrationActionCount: 81,
        },
        {
            userCount: 11,
            registeredUserCount: 11,
            anonymousUserCount: 0,
            startTime: '2020-05-24T00:00:00Z',
            searchUserCount: 4,
            searchActionCount: 58,
            codeIntelligenceUserCount: 10,
            codeIntelligenceActionCount: 20,
            integrationUserCount: 3,
            integrationActionCount: 94,
        },
        {
            userCount: 18,
            registeredUserCount: 18,
            anonymousUserCount: 0,
            startTime: '2020-05-17T00:00:00Z',
            searchUserCount: 16,
            searchActionCount: 34,
            codeIntelligenceUserCount: 12,
            codeIntelligenceActionCount: 28,
            integrationUserCount: 13,
            integrationActionCount: 31,
        },
        {
            userCount: 16,
            registeredUserCount: 16,
            anonymousUserCount: 0,
            startTime: '2020-05-10T00:00:00Z',
            searchUserCount: 4,
            searchActionCount: 28,
            codeIntelligenceUserCount: 6,
            codeIntelligenceActionCount: 110,
            integrationUserCount: 4,
            integrationActionCount: 156,
        },
        {
            userCount: 7,
            registeredUserCount: 7,
            anonymousUserCount: 0,
            startTime: '2020-05-03T00:00:00Z',
            searchUserCount: 4,
            searchActionCount: 5,
            codeIntelligenceUserCount: 6,
            codeIntelligenceActionCount: 59,
            integrationUserCount: 3,
            integrationActionCount: 42,
        },
        {
            userCount: 14,
            registeredUserCount: 14,
            anonymousUserCount: 0,
            startTime: '2020-04-26T00:00:00Z',
            searchUserCount: 13,
            searchActionCount: 101,
            codeIntelligenceUserCount: 1,
            codeIntelligenceActionCount: 20,
            integrationUserCount: 11,
            integrationActionCount: 122,
        },
    ],
    maus: [
        {
            userCount: 10,
            registeredUserCount: 10,
            anonymousUserCount: 0,
            startTime: '2020-07-01T00:00:00Z',
            searchUserCount: 1,
            searchActionCount: 95,
            codeIntelligenceUserCount: 6,
            codeIntelligenceActionCount: 45,
            integrationUserCount: 9,
            integrationActionCount: 26,
        },
        {
            userCount: 21,
            registeredUserCount: 21,
            anonymousUserCount: 0,
            startTime: '2020-06-01T00:00:00Z',
            searchUserCount: 0,
            searchActionCount: 58,
            codeIntelligenceUserCount: 19,
            codeIntelligenceActionCount: 191,
            integrationUserCount: 13,
            integrationActionCount: 84,
        },
        {
            userCount: 22,
            registeredUserCount: 22,
            anonymousUserCount: 0,
            startTime: '2020-05-01T00:00:00Z',
            searchUserCount: 8,
            searchActionCount: 32,
            codeIntelligenceUserCount: 1,
            codeIntelligenceActionCount: 129,
            integrationUserCount: 21,
            integrationActionCount: 41,
        },
    ],
} as UsageStatistics

const userUsageStatistics = {
    nodes: [
        {
            id: 'VXNlcjoy',
            username: 'jane',
            usageStatistics: {
                searchQueries: 265,
                pageViews: 949,
                codeIntelligenceActions: 51,
                lastActiveTime: '2020-07-01T00:13:56Z',
                lastActiveCodeHostIntegrationTime: null,
            },
        },
        {
            id: 'VXNlcjoz',
            username: 'john',
            usageStatistics: {
                searchQueries: 105,
                pageViews: 581,
                codeIntelligenceActions: 836,
                lastActiveTime: '2020-04-29T13:05:34Z',
                lastActiveCodeHostIntegrationTime: '2019-07-18T14:42:10Z',
            },
        },
    ],
    totalCount: 100,
} as IUserConnection

add('Loading status', () => (
    <SiteAdminUsageStatisticsPage
        {...commonProps}
        fetchSiteStatistics={() => NEVER}
        fetchUserStatistics={() => of(userUsageStatistics)}
    />
))

add('Error status', () => (
    <SiteAdminUsageStatisticsPage
        {...commonProps}
        fetchSiteStatistics={() => throwError(new Error('Error fetching usage statistics'))}
        fetchUserStatistics={() => of(userUsageStatistics)}
    />
))

add('Interactive', () => (
    <SiteAdminUsageStatisticsPage
        {...commonProps}
        fetchSiteStatistics={() => of(usageStatistics)}
        fetchUserStatistics={() => of(userUsageStatistics)}
    />
))
