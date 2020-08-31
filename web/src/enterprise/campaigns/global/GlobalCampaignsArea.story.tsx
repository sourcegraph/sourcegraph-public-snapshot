import { storiesOf } from '@storybook/react'
import webStyles from '../../../enterprise.scss'
import { boolean } from '@storybook/addon-knobs'
import React from 'react'
import { GlobalCampaignsArea } from './GlobalCampaignsArea'
import { AuthenticatedUser } from '../../../auth'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { WebStory } from '../../../components/WebStory'

const { add } = storiesOf('web/campaigns/GlobalCampaignsArea', module).addDecorator(story => (
    <React.Suspense fallback={<LoadingSpinner />}>
        <div className="p-3 container">{story()}</div>
    </React.Suspense>
))

add('Dotcom', () => (
    <WebStory webStyles={webStyles}>
        {props => (
            <GlobalCampaignsArea
                {...props}
                isSourcegraphDotCom={true}
                platformContext={undefined as any}
                extensionsController={undefined as any}
                authenticatedUser={
                    boolean('isAuthenticated', false) ? ({ username: 'alice' } as AuthenticatedUser) : null
                }
                match={{ isExact: true, path: '/campaigns', url: 'http://test.test/campaigns', params: {} }}
            />
        )}
    </WebStory>
))
