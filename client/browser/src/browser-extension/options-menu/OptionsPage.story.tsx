import { action } from '@storybook/addon-actions'
import { boolean, text } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import GithubIcon from 'mdi-react/GithubIcon'
import React, { useState } from 'react'
import { Observable, of } from 'rxjs'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'

import brandedStyles from '../../branded.scss'

import { OptionsPage, OptionsPageProps } from './OptionsPage'

const validateSourcegraphUrl = (): Observable<string | undefined> => of(undefined)
const invalidSourcegraphUrl = (): Observable<string | undefined> => of('Arbitrary error string')

const requestPermissionsHandler = action('requestPermission')

const decorator: DecoratorFn = story => <BrandedStory styles={brandedStyles}>{() => story()}</BrandedStory>

const config: Meta = {
    title: 'browser/Options/OptionsPage',
    decorators: [decorator],
    parameters: { chromatic: { delay: 500 } },
}

export default config

const OptionsPageWrapper: React.FunctionComponent<Partial<OptionsPageProps>> = props => (
    <OptionsPage
        isFullPage={true}
        isActivated={true}
        onToggleActivated={action('onToggleActivated')}
        optionFlags={[
            { key: 'allowErrorReporting', label: 'Allow error reporting', value: false },
            { key: 'experimentalLinkPreviews', label: 'Experimental link previews', value: false },
        ]}
        onChangeOptionFlag={action('onChangeOptionFlag')}
        version={text('version', '0.0.0')}
        sourcegraphUrl={text('sourcegraphUrl', 'https://sourcegraph.com')}
        validateSourcegraphUrl={validateSourcegraphUrl}
        onChangeSourcegraphUrl={action('onChangeSourcegraphUrl')}
        showPrivateRepositoryAlert={boolean('showPrivateRepositoryAlert', false)}
        showSourcegraphCloudAlert={boolean('showSourcegraphCloudAlert', false)}
        suggestedSourcegraphUrls={['https://k8s.sgdev.org', 'https://sourcegraph.com']}
        {...props}
    />
)

export const Default: Story = () => <OptionsPageWrapper />

export const Interactive: Story = () => {
    const [isActivated, setIsActivated] = useState(false)
    return <OptionsPageWrapper isActivated={isActivated} onToggleActivated={setIsActivated} />
}
export const UrlValidationError: Story = () => {
    const [isActivated, setIsActivated] = useState(false)
    return (
        <OptionsPageWrapper
            isActivated={isActivated}
            onToggleActivated={setIsActivated}
            validateSourcegraphUrl={invalidSourcegraphUrl}
            sourcegraphUrl={text('sourcegraphUrl', 'https://not-sourcegraph.com')}
        />
    )
}

UrlValidationError.storyName = 'URL validation error'

export const AskingForPermission: Story = () => (
    <OptionsPageWrapper
        permissionAlert={{ name: 'GitHub', icon: GithubIcon }}
        requestPermissionsHandler={requestPermissionsHandler}
    />
)

AskingForPermission.storyName = 'Asking for permission'

export const OnPrivateRepository: Story = () => (
    <OptionsPageWrapper showPrivateRepositoryAlert={true} requestPermissionsHandler={requestPermissionsHandler} />
)

OnPrivateRepository.storyName = 'On private repository'

export const OnSourcegraphCloud: Story = () => (
    <OptionsPageWrapper requestPermissionsHandler={requestPermissionsHandler} showSourcegraphCloudAlert={true} />
)
