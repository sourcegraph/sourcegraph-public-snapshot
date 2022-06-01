import React, { useState } from 'react'

import { action } from '@storybook/addon-actions'
import { boolean, text } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import GithubIcon from 'mdi-react/GithubIcon'
import { Observable, of } from 'rxjs'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { Typography } from '@sourcegraph/wildcard'

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

const OptionsPageWrapper: React.FunctionComponent<React.PropsWithChildren<Partial<OptionsPageProps>>> = props => (
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
        showSourcegraphCloudAlert={boolean('showSourcegraphCloudAlert', false)}
        suggestedSourcegraphUrls={['https://k8s.sgdev.org', 'https://sourcegraph.com']}
        {...props}
    />
)

const Interactive: Story = () => {
    const [isActivated, setIsActivated] = useState(false)
    return <OptionsPageWrapper isActivated={isActivated} onToggleActivated={setIsActivated} />
}

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
        <Typography.H1 className="text-center mb-3">All Options Pages</Typography.H1>
        <div>
            <div className="d-flex justify-content-center">
                <div className="mx-4">
                    <Typography.H3 className="text-center">Interactive</Typography.H3>
                    <Interactive />
                </div>
                <div className="mx-4">
                    <Typography.H3 className="text-center">URL validation error</Typography.H3>
                    <OptionsPageWrapper
                        validateSourcegraphUrl={invalidSourcegraphUrl}
                        sourcegraphUrl={text('sourcegraphUrl', 'https://not-sourcegraph.com')}
                    />
                </div>
                <div className="mx-4">
                    <Typography.H3 className="text-center">With advanced settings</Typography.H3>
                    <WithAdvancedSettings />
                </div>
            </div>

            <div className="d-flex justify-content-center mt-5">
                <div className="mx-4">
                    <Typography.H3 className="text-center">On Sourcegraph Cloud</Typography.H3>
                    <OptionsPageWrapper
                        requestPermissionsHandler={requestPermissionsHandler}
                        showSourcegraphCloudAlert={true}
                    />
                </div>
                <div className="mx-4">
                    <Typography.H3 className="text-center">Asking for permission</Typography.H3>
                    <OptionsPageWrapper
                        permissionAlert={{ name: 'GitHub', icon: GithubIcon }}
                        requestPermissionsHandler={requestPermissionsHandler}
                    />
                </div>
            </div>

            <Typography.H2 className="mt-5 text-center">Not synced repository</Typography.H2>
            <div className="d-flex justify-content-center mb-3">
                <div className="mx-4">
                    <Typography.H3 className="text-center">Sourcegraph Cloud</Typography.H3>
                    <OptionsPageWrapper
                        sourcegraphUrl="https://sourcegraph.com"
                        currentUser={{ settingsURL: '/users/john-doe/settings', siteAdmin: false }}
                        hasRepoSyncError={true}
                        requestPermissionsHandler={requestPermissionsHandler}
                    />
                </div>
                <div className="mx-4">
                    <Typography.H3 className="text-center">Self-hosted</Typography.H3>
                    <OptionsPageWrapper
                        sourcegraphUrl={text('sourcegraphUrl', 'https://k8s.sgdev.org')}
                        currentUser={{ settingsURL: '/users/john-doe/settings', siteAdmin: false }}
                        hasRepoSyncError={true}
                        requestPermissionsHandler={requestPermissionsHandler}
                    />
                </div>
                <div className="mx-4">
                    <Typography.H3 className="text-center">Self-hosted instance, user is admin</Typography.H3>
                    <OptionsPageWrapper
                        sourcegraphUrl={text('sourcegraphUrl', 'https://k8s.sgdev.org')}
                        currentUser={{ settingsURL: '/users/john-doe/settings', siteAdmin: true }}
                        hasRepoSyncError={true}
                        requestPermissionsHandler={requestPermissionsHandler}
                    />
                </div>
            </div>
        </div>
    </div>
)

AllOptionsPages.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
