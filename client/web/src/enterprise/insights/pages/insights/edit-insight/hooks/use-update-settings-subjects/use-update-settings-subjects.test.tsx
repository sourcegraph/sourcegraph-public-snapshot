import { renderHook, act } from '@testing-library/react-hooks'
import React, { PropsWithChildren } from 'react'
import { Observable, of } from 'rxjs'
import sinon from 'sinon'

import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { stringify } from '@sourcegraph/shared/src/util/jsonc'

import { Settings } from '../../../../../../../schema/settings.schema'
import { InsightsApiContext } from '../../../../../core/backend/api-provider'
import { createMockInsightAPI } from '../../../../../core/backend/create-insights-api'
import { ApiService } from '../../../../../core/backend/types'
import { InsightType, LangStatsInsight } from '../../../../../core/types'
import { createGlobalSubject, createOrgSubject, createUserSubject } from '../../../../../mocks/settings-cascade'

import { useUpdateSettingsSubject } from './use-update-settings-subjects'

const DEFAULT_OLD_INSIGHT: LangStatsInsight = {
    type: InsightType.Extension,
    title: 'old extension lang stats insight',
    id: 'codeStatsInsights.insight.oldExtensionLangStatsInsight',
    visibility: 'personal-subject-id',
    repository: '',
    otherThreshold: 3,
}

const DEFAULT_NEW_INSIGHT: LangStatsInsight = {
    type: InsightType.Extension,
    title: 'new extension lang stats insight',
    id: 'codeStatsInsights.insight.newExtensionLangStatsInsight',
    visibility: 'org-1-subject-id',
    repository: '',
    otherThreshold: 5,
}

function createAPIContext(
    settingsCascade: SettingsCascadeOrError<Settings>
): { mockAPI: ApiService; Provider: React.FunctionComponent<{}> } {
    const mockAPI = createMockInsightAPI({
        getSubjectSettings: id => {
            const subject = settingsCascade.subjects?.find(subject => subject.subject.id === id)

            if (!subject) {
                throw new Error('No subject')
            }

            return of({ id: 1000, contents: stringify(subject.settings) })
        },
    })

    const Provider: React.FunctionComponent<{}> = props => (
        <InsightsApiContext.Provider value={mockAPI}>{props.children}</InsightsApiContext.Provider>
    )

    return {
        mockAPI,
        Provider,
    }
}

describe('useUpdateSettingsSubject', () => {
    test("shouldn't update settings if there is no subjects", async () => {
        // Setup
        const updateSettingsSpy = sinon.spy<() => Observable<void>>(() => of())
        const mockAPI = createMockInsightAPI({
            updateSubjectSettings: updateSettingsSpy,
            getSubjectSettings: () => of({ id: 1000, contents: '' }),
        })

        const wrapper: React.FunctionComponent<PropsWithChildren<{}>> = props => (
            <InsightsApiContext.Provider value={mockAPI}>{props.children}</InsightsApiContext.Provider>
        )

        const { result } = renderHook(
            () =>
                useUpdateSettingsSubject({
                    platformContext: {} as any,
                }),
            { wrapper }
        )

        // Act
        await act(async () => {
            await result.current.updateSettingSubjects({
                oldInsight: DEFAULT_OLD_INSIGHT,
                newInsight: DEFAULT_NEW_INSIGHT,
                settingsCascade: {
                    final: {},
                    subjects: null,
                },
            })
        })

        expect(updateSettingsSpy.notCalled).toBe(true)
    })

    describe('when a user transferred (shared) insights', () => {
        const settingsCascade: SettingsCascadeOrError<Settings> = {
            final: {},
            subjects: [
                {
                    subject: createUserSubject('client-subject-id'),
                    settings: {
                        'codeStatsInsights.insight.oldExtensionLangStatsInsight': DEFAULT_OLD_INSIGHT,
                        'insights.dashboards': {
                            somePersonalDashboard: {
                                id: 'uuid-personal-dashboard',
                                title: 'Some personal dashboard',
                                // Include personal insight to some personal dashboard for testing
                                // insight dashboard visibility levels
                                insightIds: ['codeStatsInsights.insight.oldExtensionLangStatsInsight'],
                            },
                        },
                    },
                    lastID: 1000,
                },
                {
                    subject: createOrgSubject('org-1-subject-id'),
                    settings: {
                        'insights.dashboards': {
                            someFirstOrganizationDashboard: {
                                id: 'uuid-first-organization-dashboard',
                                title: 'Some first organization dashboard',
                                insightIds: [],
                            },
                        },
                    },
                    lastID: 1100,
                },
                {
                    subject: createGlobalSubject('global-subject-id'),
                    settings: {
                        'insights.dashboards': {
                            someGlobalOrganizationDashboard: {
                                id: 'uuid-global-dashboard',
                                title: 'Some global dashboard',
                                insightIds: [],
                            },
                        },
                    },
                    lastID: 1110,
                },
            ],
        }

        test('from the personal to some org visibility level', async () => {
            const oldInsight = { ...DEFAULT_OLD_INSIGHT, visibility: 'client-subject-id' }
            const newInsight = { ...DEFAULT_NEW_INSIGHT, visibility: 'org-1-subject-id' }

            const { Provider, mockAPI } = createAPIContext(settingsCascade)
            const updateSettingsSpy = sinon.stub(mockAPI, 'updateSubjectSettings')

            updateSettingsSpy.callsFake(() => of())

            const { result } = renderHook(
                () =>
                    useUpdateSettingsSubject({
                        platformContext: {} as any,
                    }),
                { wrapper: Provider }
            )

            // Act
            await act(async () => {
                await result.current.updateSettingSubjects({
                    oldInsight,
                    newInsight,
                    settingsCascade,
                })
            })

            expect(updateSettingsSpy.calledTwice).toBe(true)
            expect(updateSettingsSpy.firstCall.args).toStrictEqual([
                {},
                'client-subject-id',
                stringify({
                    'insights.dashboards': {
                        somePersonalDashboard: {
                            id: 'uuid-personal-dashboard',
                            title: 'Some personal dashboard',
                            insightIds: ['codeStatsInsights.insight.newExtensionLangStatsInsight'],
                        },
                    },
                }),
            ])
            expect(updateSettingsSpy.secondCall.args).toStrictEqual([
                {},
                'org-1-subject-id',
                stringify({
                    'insights.dashboards': {
                        someFirstOrganizationDashboard: {
                            id: 'uuid-first-organization-dashboard',
                            title: 'Some first organization dashboard',
                            insightIds: [],
                        },
                    },
                    'codeStatsInsights.insight.newExtensionLangStatsInsight': {
                        title: 'new extension lang stats insight',
                        repository: '',
                        otherThreshold: 5,
                    },
                }),
            ])
        })

        test('when a user transferred an insight from the personal to global visibility level', async () => {
            const oldInsight = { ...DEFAULT_OLD_INSIGHT, visibility: 'client-subject-id' }
            const newInsight = { ...DEFAULT_NEW_INSIGHT, visibility: 'global-subject-id' }

            const { Provider, mockAPI } = createAPIContext(settingsCascade)
            const updateSettingsSpy = sinon.stub(mockAPI, 'updateSubjectSettings')
            updateSettingsSpy.callsFake(() => of())

            const { result } = renderHook(() => useUpdateSettingsSubject({ platformContext: {} as any }), {
                wrapper: Provider,
            })
            // Act
            await act(async () => {
                await result.current.updateSettingSubjects({
                    oldInsight,
                    newInsight,
                    settingsCascade,
                })
            })

            expect(updateSettingsSpy.calledTwice).toBe(true)
            expect(updateSettingsSpy.firstCall.args).toStrictEqual([
                {},
                'client-subject-id',
                stringify({
                    'insights.dashboards': {
                        somePersonalDashboard: {
                            id: 'uuid-personal-dashboard',
                            title: 'Some personal dashboard',
                            insightIds: ['codeStatsInsights.insight.newExtensionLangStatsInsight'],
                        },
                    },
                }),
            ])
            expect(updateSettingsSpy.secondCall.args).toStrictEqual([
                {},
                'global-subject-id',
                stringify({
                    'insights.dashboards': {
                        someGlobalOrganizationDashboard: {
                            id: 'uuid-global-dashboard',
                            title: 'Some global dashboard',
                            insightIds: [],
                        },
                    },
                    'codeStatsInsights.insight.newExtensionLangStatsInsight': {
                        title: 'new extension lang stats insight',
                        repository: '',
                        otherThreshold: 5,
                    },
                }),
            ])
        })
    })

    describe('when a user moved (make it private) insights', () => {
        test('from global level to some organization level', async () => {
            const settingsCascade: SettingsCascadeOrError<Settings> = {
                final: {},
                subjects: [
                    {
                        subject: createUserSubject('client-subject-id'),
                        settings: {
                            'insights.dashboards': {
                                somePersonalDashboard: {
                                    id: 'uuid-personal-dashboard',
                                    title: 'Some personal dashboard',
                                    // Include personal insight to some personal dashboard for testing
                                    // insight dashboard visibility levels
                                    insightIds: ['codeStatsInsights.insight.someAnotherLangInsight'],
                                },
                            },
                        },
                        lastID: 1000,
                    },
                    {
                        subject: createOrgSubject('org-1-subject-id'),
                        settings: {
                            'insights.dashboards': {
                                someFirstOrganizationDashboard: {
                                    id: 'uuid-first-organization-dashboard',
                                    title: 'Some first organization dashboard',
                                    insightIds: ['codeStatsInsights.insight.oldExtensionLangStatsInsight'],
                                },
                            },
                        },
                        lastID: 1100,
                    },
                    {
                        subject: createGlobalSubject('global-subject-id'),
                        settings: {
                            'codeStatsInsights.insight.oldExtensionLangStatsInsight': DEFAULT_OLD_INSIGHT,
                            'insights.dashboards': {
                                someGlobalOrganizationDashboard: {
                                    id: 'uuid-global-dashboard',
                                    title: 'Some global dashboard',
                                    insightIds: ['codeStatsInsights.insight.oldExtensionLangStatsInsight'],
                                },
                            },
                        },
                        lastID: 1110,
                    },
                ],
            }

            const oldInsight = { ...DEFAULT_OLD_INSIGHT, visibility: 'global-subject-id' }
            const newInsight = { ...DEFAULT_NEW_INSIGHT, visibility: 'client-subject-id' }

            const { Provider, mockAPI } = createAPIContext(settingsCascade)
            const updateSettingsSpy = sinon.stub(mockAPI, 'updateSubjectSettings')
            updateSettingsSpy.callsFake(() => of())

            const { result } = renderHook(() => useUpdateSettingsSubject({ platformContext: {} as any }), {
                wrapper: Provider,
            })

            // Act
            await act(async () => {
                await result.current.updateSettingSubjects({
                    oldInsight,
                    newInsight,
                    settingsCascade,
                })
            })

            expect(updateSettingsSpy.callCount).toBe(3)
            expect(updateSettingsSpy.firstCall.args).toStrictEqual([
                {},
                'global-subject-id',
                stringify({
                    'insights.dashboards': {
                        someGlobalOrganizationDashboard: {
                            id: 'uuid-global-dashboard',
                            title: 'Some global dashboard',
                            // Moved insight was removed from global custom dashboard due
                            // it's public insight now
                            insightIds: [],
                        },
                    },
                }),
            ])
            expect(updateSettingsSpy.secondCall.args).toStrictEqual([
                {},
                'client-subject-id',
                stringify({
                    'insights.dashboards': {
                        somePersonalDashboard: {
                            id: 'uuid-personal-dashboard',
                            title: 'Some personal dashboard',
                            // Include personal insight to some personal dashboard for testing
                            // insight dashboard visibility levels
                            insightIds: ['codeStatsInsights.insight.someAnotherLangInsight'],
                        },
                    },
                    'codeStatsInsights.insight.newExtensionLangStatsInsight': {
                        title: 'new extension lang stats insight',
                        repository: '',
                        otherThreshold: 5,
                    },
                }),
            ])
            expect(updateSettingsSpy.thirdCall.args).toStrictEqual([
                {},
                'org-1-subject-id',
                stringify({
                    'insights.dashboards': {
                        someFirstOrganizationDashboard: {
                            id: 'uuid-first-organization-dashboard',
                            title: 'Some first organization dashboard',
                            // Moved insight was removed from global custom dashboard due
                            // it's public insight now
                            insightIds: [],
                        },
                    },
                }),
            ])
        })

        test('from on organization level to another organization level', async () => {
            const settingsCascade: SettingsCascadeOrError<Settings> = {
                final: {},
                subjects: [
                    {
                        subject: createUserSubject('client-subject-id'),
                        settings: {
                            'insights.dashboards': {
                                somePersonalDashboard: {
                                    id: 'uuid-personal-dashboard',
                                    title: 'Some personal dashboard',
                                    // Include personal insight to some personal dashboard for testing
                                    // insight dashboard visibility levels
                                    insightIds: ['codeStatsInsights.insight.someAnotherLangInsight'],
                                },
                            },
                        },
                        lastID: 1000,
                    },
                    {
                        subject: createOrgSubject('org-1-subject-id'),
                        settings: {
                            'codeStatsInsights.insight.oldExtensionLangStatsInsight': DEFAULT_OLD_INSIGHT,
                            'insights.dashboards': {
                                someFirstOrganizationDashboard: {
                                    id: 'uuid-first-organization-dashboard',
                                    title: 'Some first organization dashboard',
                                    insightIds: ['codeStatsInsights.insight.oldExtensionLangStatsInsight'],
                                },
                            },
                        },
                        lastID: 1100,
                    },
                    {
                        subject: createOrgSubject('org-2-subject-id'),
                        settings: {
                            'insights.dashboards': {
                                someSecondOrganizationDashboard: {
                                    id: 'uuid-second-organization-dashboard',
                                    title: 'Some second organization dashboard',
                                    insightIds: [],
                                },
                            },
                        },
                        lastID: 1101,
                    },
                ],
            }

            const oldInsight = { ...DEFAULT_OLD_INSIGHT, visibility: 'org-1-subject-id' }
            const newInsight = { ...DEFAULT_NEW_INSIGHT, visibility: 'org-2-subject-id' }

            const { Provider, mockAPI } = createAPIContext(settingsCascade)
            const updateSettingsSpy = sinon.stub(mockAPI, 'updateSubjectSettings')
            updateSettingsSpy.callsFake(() => of())

            const { result } = renderHook(() => useUpdateSettingsSubject({ platformContext: {} as any }), {
                wrapper: Provider,
            })

            // Act
            await act(async () => {
                await result.current.updateSettingSubjects({
                    oldInsight,
                    newInsight,
                    settingsCascade,
                })
            })

            expect(updateSettingsSpy.callCount).toBe(2)
            expect(updateSettingsSpy.firstCall.args).toStrictEqual([
                {},
                'org-1-subject-id',
                stringify({
                    'insights.dashboards': {
                        someFirstOrganizationDashboard: {
                            id: 'uuid-first-organization-dashboard',
                            title: 'Some first organization dashboard',
                            insightIds: [],
                        },
                    },
                }),
            ])
            expect(updateSettingsSpy.secondCall.args).toStrictEqual([
                {},
                'org-2-subject-id',
                stringify({
                    'insights.dashboards': {
                        someSecondOrganizationDashboard: {
                            id: 'uuid-second-organization-dashboard',
                            title: 'Some second organization dashboard',
                            insightIds: [],
                        },
                    },
                    'codeStatsInsights.insight.newExtensionLangStatsInsight': {
                        title: 'new extension lang stats insight',
                        repository: '',
                        otherThreshold: 5,
                    },
                }),
            ])
        })
    })
})
