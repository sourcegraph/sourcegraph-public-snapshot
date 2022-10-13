import React, { useState } from 'react'

import { action } from '@storybook/addon-actions'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import GithubIcon from 'mdi-react/GithubIcon'
import { Observable, of } from 'rxjs'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { H1, H2, H3 } from '@sourcegraph/wildcard'

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
        version=""
        sourcegraphUrl=""
        validateSourcegraphUrl={validateSourcegraphUrl}
        onChangeSourcegraphUrl={action('onChangeSourcegraphUrl')}
        suggestedSourcegraphUrls={['https://k8s.sgdev.org', 'https://sourcegraph.com']}
        {...props}
    />
)

const Interactive: Story = args => {
    const [isActivated, setIsActivated] = useState(false)
    return <OptionsPageWrapper isActivated={isActivated} onToggleActivated={setIsActivated} {...args} />
}

const WithAdvancedSettings: Story = args => {
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
            {...args}
        />
    )
}

export const AllOptionsPages: Story = (args = {}) => (
    <div>
        <H1 className="text-center mb-3">All Options Pages</H1>
        <div>
            <div className="d-flex justify-content-center">
                <div className="mx-4">
                    <H3 className="text-center">Interactive</H3>
                    <Interactive {...args} />
                </div>
                <div className="mx-4">
                    <H3 className="text-center">URL validation error</H3>
                    <OptionsPageWrapper validateSourcegraphUrl={invalidSourcegraphUrl} {...args} />
                </div>
                <div className="mx-4">
                    <H3 className="text-center">With advanced settings</H3>
                    <WithAdvancedSettings {...args} />
                </div>
            </div>

            <div className="d-flex justify-content-center mt-5">
                <div className="mx-4">
                    <H3 className="text-center">On Sourcegraph.com</H3>
                    <OptionsPageWrapper
                        requestPermissionsHandler={requestPermissionsHandler}
                        showSourcegraphComAlert={true}
                        sourcegraphUrl={args.sourcegraphUrl}
                        version={args.version}
                    />
                </div>
                <div className="mx-4">
                    <H3 className="text-center">Asking for permission</H3>
                    <OptionsPageWrapper
                        permissionAlert={{ name: 'GitHub', icon: GithubIcon }}
                        requestPermissionsHandler={requestPermissionsHandler}
                        {...args}
                    />
                </div>
            </div>

            <H2 className="mt-5 text-center">Not synced repository</H2>
            <div className="d-flex justify-content-center mb-3">
                <div className="mx-4">
                    <H3 className="text-center">Sourcegraph.com</H3>
                    <OptionsPageWrapper
                        sourcegraphUrl="https://sourcegraph.com"
                        currentUser={{ settingsURL: '/users/john-doe/settings', siteAdmin: false }}
                        hasRepoSyncError={true}
                        requestPermissionsHandler={requestPermissionsHandler}
                        showSourcegraphComAlert={args.showSourcegraphComAlert}
                        version={args.version}
                    />
                </div>
                <div className="mx-4">
                    <H3 className="text-center">Self-hosted</H3>
                    <OptionsPageWrapper
                        currentUser={{ settingsURL: '/users/john-doe/settings', siteAdmin: false }}
                        hasRepoSyncError={true}
                        requestPermissionsHandler={requestPermissionsHandler}
                        {...args}
                    />
                </div>
                <div className="mx-4">
                    <H3 className="text-center">Self-hosted instance, user is admin</H3>
                    <OptionsPageWrapper
                        currentUser={{ settingsURL: '/users/john-doe/settings', siteAdmin: true }}
                        hasRepoSyncError={true}
                        requestPermissionsHandler={requestPermissionsHandler}
                        {...args}
                    />
                </div>
            </div>
        </div>
    </div>
)
AllOptionsPages.argTypes = {
    sourcegraphUrl: {
        control: { type: 'text' },
        defaultValue: 'https://not-sourcegraph.com',
    },
    version: {
        control: { type: 'text' },
        defaultValue: '0.0.0',
    },
    showSourcegraphComAlert: {
        control: { type: 'boolean' },
        defaultValue: false,
    },
}

AllOptionsPages.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
