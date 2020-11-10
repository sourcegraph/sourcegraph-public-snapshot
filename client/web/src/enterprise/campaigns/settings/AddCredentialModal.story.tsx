import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'
import { ExternalServiceKind } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { AddCredentialModal } from './AddCredentialModal'

const { add } = storiesOf('web/campaigns/settings/AddCredentialModal', module)
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)
    .addParameters({
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    })

add('GitHub', () => (
    <EnterpriseWebStory>
        {props => (
            <AddCredentialModal
                {...props}
                externalServiceKind={ExternalServiceKind.GITHUB}
                externalServiceURL="https://github.com/"
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </EnterpriseWebStory>
))

add('GitLab', () => (
    <EnterpriseWebStory>
        {props => (
            <AddCredentialModal
                {...props}
                externalServiceKind={ExternalServiceKind.GITLAB}
                externalServiceURL="https://gitlab.com/"
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </EnterpriseWebStory>
))

add('Bitbucket Server', () => (
    <EnterpriseWebStory>
        {props => (
            <AddCredentialModal
                {...props}
                externalServiceKind={ExternalServiceKind.BITBUCKETSERVER}
                externalServiceURL="https://bitbucket.sgdev.org/"
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </EnterpriseWebStory>
))
