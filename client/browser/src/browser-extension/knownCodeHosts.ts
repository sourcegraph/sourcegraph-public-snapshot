import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import MicrosoftGithubIcon from 'mdi-react/MicrosoftGithubIcon'
import { PhabricatorIcon } from '../../../shared/src/components/icons'

export interface KnownCodeHost {
    host: string
    icon?: React.ComponentType
    name: string
}

export const knownCodeHosts: KnownCodeHost[] = [
    {
        host: 'github.com',
        name: 'GitHub',
        icon: MicrosoftGithubIcon,
    },
    {
        host: 'gitlab.com',
        name: 'GitLab',
        icon: GitlabIcon,
    },
    {
        host: 'bitbucket.org',
        name: 'Bitbucket',
        icon: BitbucketIcon,
    },
    {
        host: 'phabricator.com',
        name: 'Phabricator',
        icon: PhabricatorIcon,
    },
]
