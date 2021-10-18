import { action } from '@storybook/addon-actions'
import { boolean, text } from '@storybook/addon-knobs'
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

export default {
    title: 'browser/Options/OptionsPage',

    decorators: [
        story => (
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
        ),
    ],

    parameters: {
        chromatic: { delay: 500 },
    },
}

export const Default = () => (
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
)

export const Interactive = () => {
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
}

export const UrlValidationError = () => {
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
}

UrlValidationError.story = {
    name: 'URL validation error',
}

export const AskingForPermission = () => (
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
)

AskingForPermission.story = {
    name: 'Asking for permission',
}

export const OnPrivateRepository = () => (
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
)

OnPrivateRepository.story = {
    name: 'On private repository',
}

export const OnSourcegraphCloud = () => (
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
)
