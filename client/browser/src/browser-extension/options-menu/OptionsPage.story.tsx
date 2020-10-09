import React, { useState } from 'react'
import { storiesOf } from '@storybook/react'
import { OptionsPage, OptionsPageProps } from './OptionsPage'
import optionsStyles from '../../options.scss'
import { Observable, of } from 'rxjs'
import { action } from '@storybook/addon-actions'
import { boolean, text } from '@storybook/addon-knobs'
import MicrosoftGithubIcon from 'mdi-react/MicrosoftGithubIcon'

const validateSourcegraphUrl = (): Observable<string | undefined> => of(undefined)
const invalidSourcegraphUrl = (): Observable<string | undefined> => of('Arbitrary error string')

const onChangeOptionFlag = action('onChangeOptionFlag')
const optionFlags: OptionsPageProps['optionFlags'] = [
    {key: 'allowErrorReporting', label: 'Allow error reporting', value: false},
    {key: 'experimentalLinkPreviews', label: 'Experimental link previews', value: false},
    {key: 'experimentalTextFieldCompletion', label: 'Experimental text field completion', value: false}
]

storiesOf('browser/Options/OptionsPage', module)
    .addDecorator(story => (
        <>
            <style>{optionsStyles}</style>
            <div>{story()}</div>
        </>
    ))
    .add('Default', () => (
        <OptionsPage
            version={text('version', '0.0.0')}
            // isCurrentRepositoryPrivate={boolean('isCurrentRepositoryPrivate', false)}
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
                // isCurrentRepositoryPrivate={boolean('isCurrentRepositoryPrivate', false)}
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
                // isCurrentRepositoryPrivate={boolean('isCurrentRepositoryPrivate', false)}
                isFullPage={true}
                optionFlags={optionFlags}
                onChangeOptionFlag={onChangeOptionFlag}
            />
        )
    })
    .add('Asking for permission', () => (
            <OptionsPage
            version={text('version', '0.0.0')}
            // isCurrentRepositoryPrivate={boolean('isCurrentRepositoryPrivate', false)}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={action('onToggleActivated')}
            isActivated={true}
            sourcegraphUrl={text('sourcegraphUrl', 'https://sourcegraph.com')}
            isFullPage={true}
            optionFlags={optionFlags}
            onChangeOptionFlag={onChangeOptionFlag}
            currentHost="github.com"
            permissionAlert={{name: 'GitHub', icon: MicrosoftGithubIcon}}
        />
        ))
