import React, { useState } from 'react'

import { action } from '@storybook/addon-actions'
import type { Decorator, Meta, StoryFn } from '@storybook/react'
import GithubIcon from 'mdi-react/GithubIcon'
import { type Observable, of } from 'rxjs'

import { Grid, H1, H2, H3 } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { OptionsPage, type OptionsPageProps } from './OptionsPage'

import brandedStyles from '../../branded.scss'

const validateSourcegraphUrl = (): Observable<string | undefined> => of(undefined)
const invalidSourcegraphUrl = (): Observable<string | undefined> => of('Arbitrary error string')

const requestPermissionsHandler = action('requestPermission')

const decorator: Decorator = story => <BrandedStory styles={brandedStyles}>{() => story()}</BrandedStory>

const config: Meta = {
    title: 'browser/Options/OptionsPage',
    decorators: [decorator],
}

export default config

const OptionsPageWrapper: React.FunctionComponent<React.PropsWithChildren<Partial<OptionsPageProps>>> = props => {
    const [urls, setUrls] = useState([
        'https://sourcegraph.com',
        'https://k8s.sgdev.org',
        'https://sourcegraph.sourcegraph.com',
    ])

    return (
        <OptionsPage
            isFullPage={false}
            isActivated={true}
            onToggleActivated={action('onToggleActivated')}
            optionFlags={[{ key: 'allowErrorReporting', label: 'Allow error reporting', value: false }]}
            onChangeOptionFlag={action('onChangeOptionFlag')}
            version=""
            sourcegraphUrl=""
            validateSourcegraphUrl={validateSourcegraphUrl}
            onChangeSourcegraphUrl={action('onChangeSourcegraphUrl')}
            suggestedSourcegraphUrls={urls}
            onSuggestedSourcegraphUrlDelete={url => setUrls(urls.filter(item => item !== url))}
            {...props}
        />
    )
}

const Interactive: StoryFn = args => {
    const [isActivated, setIsActivated] = useState(false)
    return <OptionsPageWrapper isActivated={isActivated} onToggleActivated={setIsActivated} {...args} />
}

const WithAdvancedSettings: StoryFn = args => {
    const [optionFlagValues, setOptionFlagValues] = useState([
        { key: 'allowErrorReporting', label: 'Allow error reporting', value: false },
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

export const AllOptionsPages: StoryFn = (args = {}) => (
    <div>
        <H1 className="text-center mb-3">All Options Pages</H1>
        <Grid columnCount={3}>
            <div>
                <H3 className="text-center">Interactive</H3>
                <Interactive {...args} />
            </div>
            <div>
                <H3 className="text-center">URL validation error</H3>
                <OptionsPageWrapper validateSourcegraphUrl={invalidSourcegraphUrl} {...args} />
            </div>
            <div>
                <H3 className="text-center">With advanced settings</H3>
                <WithAdvancedSettings {...args} />
            </div>
            <div>
                <H3 className="text-center">No previous url suggestion</H3>
                <OptionsPageWrapper suggestedSourcegraphUrls={[]} {...args} />
            </div>
            <div>
                <H3 className="text-center">On Sourcegraph.com</H3>
                <OptionsPageWrapper
                    requestPermissionsHandler={requestPermissionsHandler}
                    showSourcegraphComAlert={true}
                    sourcegraphUrl={args.sourcegraphUrl}
                    version={args.version}
                />
            </div>
            <div>
                <H3 className="text-center">Asking for permission</H3>
                <OptionsPageWrapper
                    permissionAlert={{ name: 'GitHub', icon: GithubIcon }}
                    requestPermissionsHandler={requestPermissionsHandler}
                    {...args}
                />
            </div>
        </Grid>
        <H2 className="mt-5 text-center">Not synced repository</H2>
        <Grid columnCount={3}>
            <div>
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
            <div>
                <H3 className="text-center">Self-hosted</H3>
                <OptionsPageWrapper
                    currentUser={{ settingsURL: '/users/john-doe/settings', siteAdmin: false }}
                    hasRepoSyncError={true}
                    requestPermissionsHandler={requestPermissionsHandler}
                    {...args}
                />
            </div>
            <div>
                <H3 className="text-center">Self-hosted instance, user is admin</H3>
                <OptionsPageWrapper
                    currentUser={{ settingsURL: '/users/john-doe/settings', siteAdmin: true }}
                    hasRepoSyncError={true}
                    requestPermissionsHandler={requestPermissionsHandler}
                    {...args}
                />
            </div>
        </Grid>
    </div>
)
AllOptionsPages.argTypes = {
    sourcegraphUrl: {
        control: { type: 'text' },
    },
    version: {
        control: { type: 'text' },
    },
    showSourcegraphComAlert: {
        control: { type: 'boolean' },
    },
}
AllOptionsPages.args = {
    sourcegraphUrl: 'https://not-sourcegraph.com',
    version: '0.0.0',
    showSourcegraphComAlert: false,
}
