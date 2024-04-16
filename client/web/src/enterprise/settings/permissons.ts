import type { BadgeProps } from '@sourcegraph/wildcard'

export const PermissionReasonBadgeProps: { [reason: string]: BadgeProps } = {
    'Permissions Sync': {
        variant: 'success',
        tooltip: 'The repository is accessible to the user due to permissions syncing from code host.',
    },
    Unrestricted: {
        variant: 'primary',
        tooltip:
            'The repository is accessible to all the users because one of the external services it belongs to does not have authorization enabled.',
    },
    'Site Admin': { variant: 'secondary', tooltip: 'The user is site admin and has access to all the repositories.' },
    'Explicit API': {
        variant: 'success',
        tooltip: 'The permission was granted through explicit permissions API.',
    },
    Public: {
        variant: 'primary',
        tooltip: 'This repository is marked as public on the code host and is accessible to all users.',
    },
}
