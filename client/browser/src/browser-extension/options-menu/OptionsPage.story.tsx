import React, { useState } from 'react'
import { storiesOf } from '@storybook/react'
import { OptionsPage, OptionsPageProps } from './OptionsPage'
import brandedStyles from '../../branded.scss'
import { Observable, of } from 'rxjs'
import { action } from '@storybook/addon-actions'
import { boolean, text } from '@storybook/addon-knobs'
import GithubIcon from 'mdi-react/GithubIcon'
import { BrandedStory } from '../../../../branded/src/components/BrandedStory'

const validateSourcegraphUrl = (): Observable<string | undefined> => of(undefined)
const invalidSourcegraphUrl = (): Observable<string | undefined> => of('Arbitrary error string')

const onChangeOptionFlag = action('onChangeOptionFlag')
const optionFlags: OptionsPageProps['optionFlags'] = [
    { key: 'allowErrorReporting', label: 'Allow error reporting', value: false },
    { key: 'experimentalLinkPreviews', label: 'Experimental link previews', value: false },
    { key: 'experimentalTextFieldCompletion', label: 'Experimental text field completion', value: false },
]

const requestPermissionsHandler = action('requestPermission')

storiesOf('browser/Options/OptionsPage', module)
    .addDecorator(story => <BrandedStory styles={brandedStyles}>{() => story()}</BrandedStory>)
    .add('Default', () => (
        <OptionsPage
            version={text('version', '0.0.0')}
            showPrivateRepositoryAlert={boolean('isCurrentRepositoryPrivate', false)}
            showSourcegraphCloudAlert={boolean('showSourcegraphCloudAlert', false)}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            sourcegraphUrl={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
            optionFlags={optionFlags}
            onChangeOptionFlag={onChangeOptionFlag}
        />
    ))
    .add('Interactive', () => {
        const [isActivated, setIsActivated] = useState(false)
        return (
            <OptionsPage
                version={text('version', '0.0.0')}
                isActivated={isActivated}
                onToggleActivated={setIsActivated}
                validateSourcegraphUrl={validateSourcegraphUrl}
                sourcegraphUrl={text('sourcegraphUrl', 'https://sourcegraph.com')}
                showPrivateRepositoryAlert={boolean('showPrivateRepositoryAlert', false)}
                showSourcegraphCloudAlert={boolean('showSourcegraphCloudAlert', false)}
                isFullPage={true}
                optionFlags={optionFlags}
                onChangeOptionFlag={onChangeOptionFlag}
            />
        )
    })
    .add('URL validation error', () => {
        const [isActivated, setIsActivated] = useState(false)
        return (
            <OptionsPage
                version={text('version', '0.0.0')}
                isActivated={isActivated}
                onToggleActivated={setIsActivated}
                validateSourcegraphUrl={invalidSourcegraphUrl}
                sourcegraphUrl={text('sourcegraphUrl', 'https://not-sourcegraph.com')}
                isFullPage={true}
                optionFlags={optionFlags}
                onChangeOptionFlag={onChangeOptionFlag}
            />
        )
    })
    .add('Asking for permission', () => (
        <OptionsPage
            version={text('version', '0.0.0')}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            sourcegraphUrl={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
            optionFlags={optionFlags}
            onChangeOptionFlag={onChangeOptionFlag}
            currentHost="github.com"
            permissionAlert={{ name: 'GitHub', icon: GithubIcon }}
            requestPermissionsHandler={requestPermissionsHandler}
        />
    ))
    .add('On private repository', () => (
        <OptionsPage
            version={text('version', '0.0.0')}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            sourcegraphUrl={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
            optionFlags={optionFlags}
            onChangeOptionFlag={onChangeOptionFlag}
            currentHost="github.com"
            showPrivateRepositoryAlert={true}
            requestPermissionsHandler={requestPermissionsHandler}
        />
    ))
    .add('On Sourcegraph Cloud', () => (
        <OptionsPage
            version={text('version', '0.0.0')}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            sourcegraphUrl={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
            optionFlags={optionFlags}
            onChangeOptionFlag={onChangeOptionFlag}
            currentHost="sourcegraph.com"
            requestPermissionsHandler={requestPermissionsHandler}
            showSourcegraphCloudAlert={true}
        />
    ))
