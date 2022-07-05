import { mdiBitbucket, mdiGithub, mdiGitlab } from '@mdi/js'

import { PhabricatorIcon } from '@sourcegraph/shared/src/components/icons'
import { IconType } from '@sourcegraph/wildcard'

export interface KnownCodeHost {
    host: string
    icon?: IconType
    name: string
}

export const knownCodeHosts: KnownCodeHost[] = [
    {
        host: 'github.com',
        name: 'GitHub',
        icon: mdiGithub,
    },
    {
        host: 'gitlab.com',
        name: 'GitLab',
        icon: mdiGitlab,
    },
    {
        host: 'bitbucket.org',
        name: 'Bitbucket',
        icon: mdiBitbucket,
    },
    {
        host: 'phabricator.com',
        name: 'Phabricator',
        icon: PhabricatorIcon,
    },
]
