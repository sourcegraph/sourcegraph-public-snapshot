import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascade } from '@sourcegraph/shared/src/settings/settings'

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
