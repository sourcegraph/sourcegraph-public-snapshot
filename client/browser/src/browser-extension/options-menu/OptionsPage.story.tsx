import React, { useState } from 'react'

import { action } from '@storybook/addon-actions'
import { boolean, text } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import GithubIcon from 'mdi-react/GithubIcon'
import { Observable, of } from 'rxjs'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'

import { OptionsPage, OptionsPageProps } from './OptionsPage'

import brandedStyles from '../../branded.scss'

const validateSourcegraphUrl = (): Observable<string | undefined> => of(undefined)
const invalidSourcegraphUrl = (): Observable<string | undefined> => of('Arbitrary error string')

const requestPermissionsHandler = action('requestPermission')

const decorator: DecoratorFn = story => <BrandedStory styles={brandedStyles}>{() => story()}</BrandedStory>

const config: Meta = {
    title: 'browser/Options/OptionsPage',
    decorators: [decorator],
}

export default config

const OptionsPageWrapper: React.FunctionComponent<Partial<OptionsPageProps>> = props => (
    <OptionsPage
        isFullPage={false}
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

const Interactive: Story = () => {
    const [isActivated, setIsActivated] = useState(false)
    return <OptionsPageWrapper isActivated={isActivated} onToggleActivated={setIsActivated} />
}

const UrlValidationError: Story = () => (
    <OptionsPageWrapper
        validateSourcegraphUrl={invalidSourcegraphUrl}
        sourcegraphUrl={text('sourcegraphUrl', 'https://not-sourcegraph.com')}
    />
)

const AskingForPermission: Story = () => (
    <OptionsPageWrapper
        permissionAlert={{ name: 'GitHub', icon: GithubIcon }}
        requestPermissionsHandler={requestPermissionsHandler}
    />
)

const OnPrivateRepository: Story = () => (
    <OptionsPageWrapper showPrivateRepositoryAlert={true} requestPermissionsHandler={requestPermissionsHandler} />
)

const OnSourcegraphCloud: Story = () => (
    <OptionsPageWrapper requestPermissionsHandler={requestPermissionsHandler} showSourcegraphCloudAlert={true} />
)

const WithAdvancedSettings: Story = () => {
    const [optionFlagValues, setOptionFlagValues] = useState([
        { key: 'allowErrorReporting', label: 'Allow error reporting', value: false },
        { key: 'experimentalLinkPreviews', label: 'Experimental link previews', value: true },
    ])
    const setOptionFlag = (key: string, value: boolean) => {
        setOptionFlagValues(optionFlagValues.map(flag => (flag.key === key ? { ...flag, value } : flag)))
    }

    return (
        <OptionsPageWrapper
            initialShowAdvancedSettings={true}
            optionFlags={optionFlagValues}
            onChangeOptionFlag={setOptionFlag}
        />
    )
}

export const AllOptionsPages: Story = () => (
    <div>
        <h1 className="text-center">All Options Pages</h1>
        <div>
            <div className="d-flex justify-content-center">
                <div className="mx-4">
                    <h2 className="text-center">Interactive</h2>
                    <Interactive />
                </div>
                <div className="mx-4">
                    <h2 className="text-center">URL validation error</h2>
                    <UrlValidationError />
                </div>
                <div className="mx-4">
                    <h2 className="text-center">With advanced settings</h2>
                    <WithAdvancedSettings />
                </div>
            </div>
            <div className="d-flex justify-content-center mt-5">
                <div className="mx-4">
                    <h2 className="text-center">On private repository</h2>
                    <OnPrivateRepository />
                </div>
                <div className="mx-4">
                    <h2 className="text-center">On Sourcegraph Cloud</h2>
                    <OnSourcegraphCloud />
                </div>
                <div className="mx-4">
                    <h2 className="text-center">Asking for permission</h2>
                    <AskingForPermission />
                </div>
            </div>
        </div>
    </div>
)

OnSourcegraphCloud.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
