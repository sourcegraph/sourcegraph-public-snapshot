import { SettingsCascade } from '@sourcegraph/shared/src/settings/settings'

export const SETTINGS_CASCADE: SettingsCascade = {
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
    ],
    final: {},
}
