import { action } from '@storybook/addon-actions'
import { boolean, text } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import GithubIcon from 'mdi-react/GithubIcon'
import React, { useState } from 'react'
import { Observable, of } from 'rxjs'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { subtypeOf } from '@sourcegraph/shared/src/util/types'

import brandedStyles from '../../branded.scss'

import { OptionsPage, OptionsPageProps } from './OptionsPage'
import { OptionsPageContext } from './OptionsPage.context'

const validateSourcegraphUrl = (): Observable<string | undefined> => of(undefined)
const invalidSourcegraphUrl = (): Observable<string | undefined> => of('Arbitrary error string')

const commonProps = subtypeOf<Partial<OptionsPageProps>>()({
    version: text('version', '0.0.0'),
    onSelfHostedSourcegraphURLChange: action('onSelfHostedSourcegraphURLChange'),
})

const requestPermissionsHandler = action('requestPermission')

storiesOf('browser/Options/OptionsPage', module)
    .addDecorator(story => (
        <BrandedStory styles={brandedStyles}>
            {() => (
                <OptionsPageContext.Provider
                    value={{
                        onChangeOptionFlag: action('onChangeOptionFlag'),
                        optionFlags: [
                            { key: 'allowErrorReporting', label: 'Allow error reporting', value: false },
                            { key: 'experimentalLinkPreviews', label: 'Experimental link previews', value: false },
                        ],
                        onBlocklistChange: () => {},
                    }}
                >
                    {story()}
                </OptionsPageContext.Provider>
            )}
        </BrandedStory>
    ))
    .addParameters({
        chromatic: { delay: 500 },
    })
    .add('Default', () => (
        <OptionsPage
            {...commonProps}
            showPrivateRepositoryAlert={boolean('isCurrentRepositoryPrivate', false)}
            showSourcegraphCloudAlert={boolean('showSourcegraphCloudAlert', false)}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            selfHostedSourcegraphURL={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
        />
    ))
    .add('Interactive', () => {
        const [isActivated, setIsActivated] = useState(false)
        return (
            <OptionsPage
                {...commonProps}
                isActivated={isActivated}
                onToggleActivated={setIsActivated}
                validateSourcegraphUrl={validateSourcegraphUrl}
                selfHostedSourcegraphURL={text('sourcegraphUrl', 'https://sourcegraph.com')}
                showPrivateRepositoryAlert={boolean('showPrivateRepositoryAlert', false)}
                showSourcegraphCloudAlert={boolean('showSourcegraphCloudAlert', false)}
                isFullPage={true}
            />
        )
    })
    .add('URL validation error', () => {
        const [isActivated, setIsActivated] = useState(false)
        return (
            <OptionsPage
                {...commonProps}
                isActivated={isActivated}
                onToggleActivated={setIsActivated}
                validateSourcegraphUrl={invalidSourcegraphUrl}
                selfHostedSourcegraphURL={text('sourcegraphUrl', 'https://sourcegraph.com')}
                isFullPage={true}
            />
        )
    })
    .add('Asking for permission', () => (
        <OptionsPage
            {...commonProps}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            selfHostedSourcegraphURL={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
            permissionAlert={{ name: 'GitHub', icon: GithubIcon }}
            requestPermissionsHandler={requestPermissionsHandler}
        />
    ))
    .add('On private repository', () => (
        <OptionsPage
            {...commonProps}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            selfHostedSourcegraphURL={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
            showPrivateRepositoryAlert={true}
            requestPermissionsHandler={requestPermissionsHandler}
        />
    ))
    .add('On Sourcegraph Cloud', () => (
        <OptionsPage
            {...commonProps}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            selfHostedSourcegraphURL={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
            requestPermissionsHandler={requestPermissionsHandler}
            showSourcegraphCloudAlert={true}
        />
    ))
