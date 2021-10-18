import { SettingsCascade, SettingsSubject } from '@sourcegraph/shared/src/settings/settings'

import { Settings } from '../../../schema/settings.schema';

export const SETTINGS_CASCADE_MOCK: SettingsCascade<Settings> = {
    subjects: [
        {
            lastID: 102,
            settings: {},
            subject: {
                __typename: 'User' as const,
                id: 'user_test_id',
                username: 'testusername',
                displayName: 'test',
                viewerCanAdminister: true,
            },
        },
        {
            lastID: 101,
            settings: {},
            subject: {
                __typename: 'Org' as const,
                name: 'test organization 2',
                displayName: 'Test organization 2',
                viewerCanAdminister: true,
                id: 'test_org_2_id',
            },
        },
        {
            lastID: 101,
            settings: {},
            subject: {
                __typename: 'Site' as const,
                viewerCanAdminister: true,
                allowSiteSettingsEdits: true,
                id: 'global_id',
            },
        },
    ],
    final: {},
}

export const createOrgSubject = (id: string): SettingsSubject => ({
    __typename: 'Org' as const,
    id,
    name: `test_org_${id}`,
    displayName: `Test organization ${id}`,
    viewerCanAdminister: true,
})

export const createUserSubject = (id: string): SettingsSubject => ({
    __typename: 'User' as const,
    id,
    username: `test_username_${id}`,
    displayName: `Test user ${id}`,
    viewerCanAdminister: true,
})

export const createGlobalSubject = (id: string): SettingsSubject => ({
    __typename: 'Site' as const,
    id,
    viewerCanAdminister: true,
    allowSiteSettingsEdits: true,
})
