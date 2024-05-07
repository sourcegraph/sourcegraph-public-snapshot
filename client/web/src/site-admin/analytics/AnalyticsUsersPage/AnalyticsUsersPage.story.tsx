import type { MockedResponse } from '@apollo/client/testing'
import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import type { UsersStatisticsResult } from '../../../graphql-operations'

import { AnalyticsUsersPage } from './index'
import { USERS_STATISTICS } from './queries'

const decorator: Decorator = story => <WebStory>{() => <div className="p-3 container">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/site-admin/analytics/AnalyticsUsersPage',
    decorators: [decorator],
}

export default config

const USER_ANALYTICS_QUERY_MOCK: MockedResponse<UsersStatisticsResult> = {
    request: {
        query: getDocumentNode(USERS_STATISTICS),
        variables: {
            dateRange: 'LAST_THREE_MONTHS',
            grouping: 'WEEKLY',
        },
    },
    result: {
        data: {
            site: {
                analytics: {
                    users: {
                        monthlyActiveUsers: [
                            {
                                date: '2022-08',
                                count: 99778,
                                __typename: 'AnalyticsMonthlyActiveUsers',
                            },
                            {
                                date: '2022-09',
                                count: 99778,
                                __typename: 'AnalyticsMonthlyActiveUsers',
                            },
                            {
                                date: '2022-10',
                                count: 99778,
                                __typename: 'AnalyticsMonthlyActiveUsers',
                            },
                        ],
                        activity: {
                            nodes: [
                                {
                                    date: '2022-09-05T00:00:00Z',
                                    count: 1679289,
                                    uniqueUsers: 1148,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-08-29T00:00:00Z',
                                    count: 2017644,
                                    uniqueUsers: 1280,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-08-22T00:00:00Z',
                                    count: 2022434,
                                    uniqueUsers: 1313,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-08-15T00:00:00Z',
                                    count: 2118379,
                                    uniqueUsers: 1338,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-08-08T00:00:00Z',
                                    count: 2092362,
                                    uniqueUsers: 1383,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-08-01T00:00:00Z',
                                    count: 2155192,
                                    uniqueUsers: 1502,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-07-25T00:00:00Z',
                                    count: 1896399,
                                    uniqueUsers: 1492,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-07-18T00:00:00Z',
                                    count: 1897174,
                                    uniqueUsers: 1672,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-07-11T00:00:00Z',
                                    count: 1847606,
                                    uniqueUsers: 1746,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-07-04T00:00:00Z',
                                    count: 1817582,
                                    uniqueUsers: 1797,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-06-27T00:00:00Z',
                                    count: 1808359,
                                    uniqueUsers: 1914,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-06-20T00:00:00Z',
                                    count: 2014862,
                                    uniqueUsers: 2091,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                                {
                                    date: '2022-06-13T00:00:00Z',
                                    count: 1967790,
                                    uniqueUsers: 2109,
                                    __typename: 'AnalyticsStatItemNode',
                                },
                            ],
                            summary: {
                                totalCount: 25888117,
                                totalUniqueUsers: 6663,
                                totalRegisteredUsers: 6663,
                                __typename: 'AnalyticsStatItemSummary',
                            },
                            __typename: 'AnalyticsStatItem',
                        },
                        frequencies: [
                            {
                                daysUsed: 1,
                                frequency: 928939,
                                percentage: 0.8690111238857171,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 2,
                                frequency: 76622,
                                percentage: 0.07167894806265149,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 3,
                                frequency: 24992,
                                percentage: 0.02337971170136235,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 4,
                                frequency: 10372,
                                percentage: 0.009702879712169106,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 5,
                                frequency: 5773,
                                percentage: 0.005400571208865431,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 6,
                                frequency: 3714,
                                percentage: 0.0034744017789236463,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 7,
                                frequency: 2754,
                                percentage: 0.0025763334677317506,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 8,
                                frequency: 2126,
                                percentage: 0.0019888471141603858,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 9,
                                frequency: 1723,
                                percentage: 0.001611845521024621,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 10,
                                frequency: 1376,
                                percentage: 0.0012872312460417172,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 11,
                                frequency: 1173,
                                percentage: 0.0010973272177375976,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 12,
                                frequency: 1029,
                                percentage: 0.0009626169710588131,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 13,
                                frequency: 841,
                                percentage: 0.0007867452601170669,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 14,
                                frequency: 737,
                                percentage: 0.0006894545264046116,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 15,
                                frequency: 653,
                                percentage: 0.0006108735491753207,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 16,
                                frequency: 567,
                                percentage: 0.0005304215962977134,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 17,
                                frequency: 494,
                                percentage: 0.000462130985134163,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 18,
                                frequency: 472,
                                percentage: 0.00044155025300268204,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 19,
                                frequency: 414,
                                percentage: 0.000387291959201505,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 20,
                                frequency: 355,
                                percentage: 0.00033209817757616977,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 21,
                                frequency: 346,
                                percentage: 0.0003236787871587457,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 22,
                                frequency: 317,
                                percentage: 0.0002965496402581572,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 23,
                                frequency: 241,
                                percentage: 0.00022545256562213215,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 24,
                                frequency: 209,
                                percentage: 0.00019551695524906896,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 25,
                                frequency: 232,
                                percentage: 0.00021703317520470813,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 26,
                                frequency: 199,
                                percentage: 0.0001861620770074867,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 27,
                                frequency: 181,
                                percentage: 0.00016932329617263867,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 28,
                                frequency: 162,
                                percentage: 0.0001515490275136324,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 29,
                                frequency: 131,
                                percentage: 0.00012254890496472744,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 30,
                                frequency: 123,
                                percentage: 0.00011506500237146163,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 31,
                                frequency: 126,
                                percentage: 0.00011787146584393631,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 32,
                                frequency: 119,
                                percentage: 0.00011132305107482874,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 33,
                                frequency: 119,
                                percentage: 0.00011132305107482874,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 34,
                                frequency: 104,
                                percentage: 0.00009729073371245536,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 35,
                                frequency: 88,
                                percentage: 0.00008232292852592377,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 36,
                                frequency: 85,
                                percentage: 0.0000795164650534491,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 37,
                                frequency: 67,
                                percentage: 0.00006267768421860106,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 38,
                                frequency: 58,
                                percentage: 0.00005425829380117703,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 39,
                                frequency: 64,
                                percentage: 0.00005987122074612638,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 40,
                                frequency: 67,
                                percentage: 0.00006267768421860106,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 41,
                                frequency: 55,
                                percentage: 0.00005145183032870236,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 42,
                                frequency: 59,
                                percentage: 0.000055193781625335255,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 43,
                                frequency: 56,
                                percentage: 0.00005238731815286058,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 44,
                                frequency: 38,
                                percentage: 0.00003554853731801254,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 45,
                                frequency: 54,
                                percentage: 0.00005051634250454413,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 46,
                                frequency: 32,
                                percentage: 0.00002993561037306319,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 47,
                                frequency: 33,
                                percentage: 0.000030871098197221414,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 48,
                                frequency: 47,
                                percentage: 0.000043967927735436557,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 49,
                                frequency: 34,
                                percentage: 0.000031806586021379636,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 50,
                                frequency: 33,
                                percentage: 0.000030871098197221414,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 51,
                                frequency: 33,
                                percentage: 0.000030871098197221414,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 52,
                                frequency: 27,
                                percentage: 0.000025258171252272065,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 53,
                                frequency: 22,
                                percentage: 0.000020580732131480942,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 54,
                                frequency: 21,
                                percentage: 0.000019645244307322716,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 55,
                                frequency: 18,
                                percentage: 0.000016838780834848044,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 56,
                                frequency: 27,
                                percentage: 0.000025258171252272065,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 57,
                                frequency: 14,
                                percentage: 0.000013096829538215145,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 58,
                                frequency: 16,
                                percentage: 0.000014967805186531595,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 59,
                                frequency: 16,
                                percentage: 0.000014967805186531595,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 60,
                                frequency: 15,
                                percentage: 0.00001403231736237337,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 61,
                                frequency: 20,
                                percentage: 0.000018709756483164494,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 62,
                                frequency: 16,
                                percentage: 0.000014967805186531595,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 63,
                                frequency: 9,
                                percentage: 0.000008419390417424022,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 64,
                                frequency: 6,
                                percentage: 0.000005612926944949348,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 65,
                                frequency: 7,
                                percentage: 0.000006548414769107573,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 66,
                                frequency: 8,
                                percentage: 0.000007483902593265797,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 67,
                                frequency: 4,
                                percentage: 0.0000037419512966328986,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 68,
                                frequency: 9,
                                percentage: 0.000008419390417424022,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 69,
                                frequency: 6,
                                percentage: 0.000005612926944949348,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 70,
                                frequency: 8,
                                percentage: 0.000007483902593265797,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 71,
                                frequency: 7,
                                percentage: 0.000006548414769107573,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 72,
                                frequency: 7,
                                percentage: 0.000006548414769107573,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 73,
                                frequency: 3,
                                percentage: 0.000002806463472474674,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 74,
                                frequency: 4,
                                percentage: 0.0000037419512966328986,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 75,
                                frequency: 6,
                                percentage: 0.000005612926944949348,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 76,
                                frequency: 3,
                                percentage: 0.000002806463472474674,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 77,
                                frequency: 2,
                                percentage: 0.0000018709756483164493,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 78,
                                frequency: 3,
                                percentage: 0.000002806463472474674,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 79,
                                frequency: 6,
                                percentage: 0.000005612926944949348,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 80,
                                frequency: 3,
                                percentage: 0.000002806463472474674,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 82,
                                frequency: 1,
                                percentage: 9.354878241582247e-7,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 83,
                                frequency: 1,
                                percentage: 9.354878241582247e-7,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 84,
                                frequency: 1,
                                percentage: 9.354878241582247e-7,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 85,
                                frequency: 1,
                                percentage: 9.354878241582247e-7,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 88,
                                frequency: 1,
                                percentage: 9.354878241582247e-7,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 89,
                                frequency: 1,
                                percentage: 9.354878241582247e-7,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 90,
                                frequency: 3,
                                percentage: 0.000002806463472474674,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                            {
                                daysUsed: 93,
                                frequency: 1,
                                percentage: 9.354878241582247e-7,
                                __typename: 'AnalyticsUsersFrequencyItem',
                            },
                        ],
                        __typename: 'AnalyticsUsersResult',
                    },
                    __typename: 'Analytics',
                },
                productSubscription: {
                    license: {
                        userCount: 999999,
                        __typename: 'ProductLicenseInfo',
                    },
                    __typename: 'ProductSubscriptionStatus',
                },
                __typename: 'Site',
            },
            users: {
                totalCount: 49036,
                __typename: 'UserConnection',
            },
            pendingAccessRequests: {
                totalCount: 123,
                __typename: 'AccessRequestConnection',
            },
        },
    },
}

export const AnalyticsUsersPageExample: StoryFn = () => (
    <MockedTestProvider mocks={[USER_ANALYTICS_QUERY_MOCK]}>
        <AnalyticsUsersPage telemetryRecorder={noOpTelemetryRecorder} />
    </MockedTestProvider>
)
