import { DecoratorFn, Meta, Story } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { WebStory } from '../../components/WebStory'
import { MockedFeatureFlagsProvider } from '../../featureFlags/MockedFeatureFlagsProvider'
import { useExperimentalFeatures } from '../../stores'
import { ThemePreference } from '../../theme'

import { SearchPage, SearchPageProps } from './SearchPage'

const defaultProps = (props: ThemeProps): SearchPageProps => ({
    isSourcegraphDotCom: false,
    settingsCascade: {
        final: null,
        subjects: null,
    },
    extensionsController,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    themePreference: ThemePreference.Light,
    onThemePreferenceChange: () => undefined,
    authenticatedUser: null,
    globbing: false,
    platformContext: {} as any,
    searchContextsEnabled: true,
    selectedSearchContextSpec: '',
    setSelectedSearchContextSpec: () => {},
    isLightTheme: props.isLightTheme,
    fetchSearchContexts: mockFetchSearchContexts,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
})

window.context.allowSignup = true

const decorator: DecoratorFn = Story => {
    useExperimentalFeatures.setState({ showSearchContext: false })
    return <Story />
}

const config: Meta = {
    title: 'web/search/home/SearchPage',
    decorators: [decorator],
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
        chromatic: { viewports: [544, 577, 769, 993], disableSnapshot: false },
    },
}

export default config
export const CloudAuthedHome: Story = () => (
    <WebStory>{webProps => <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} />}</WebStory>
)

CloudAuthedHome.storyName = 'Cloud authenticated home'

export const ServerHome: Story = () => <WebStory>{webProps => <SearchPage {...defaultProps(webProps)} />}</WebStory>

ServerHome.storyName = 'Server home'

export const CloudMarketingHome: Story = () => (
    <WebStory>
        {webProps => (
            <MockedFeatureFlagsProvider overrides={{}}>
                <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} authenticatedUser={null} />
            </MockedFeatureFlagsProvider>
        )}
    </WebStory>
)

CloudMarketingHome.storyName = 'Cloud marketing home'
