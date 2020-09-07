import React from 'react'
import { createMemoryHistory } from 'history'
import { NOOP_TELEMETRY_SERVICE } from '../../../../shared/src/telemetry/telemetryService'
import { SearchPage, SearchPageProps } from './SearchPage'
import { SearchPatternType } from '../../graphql-operations'
import { Services } from '../../../../shared/src/api/client/services'
import { storiesOf } from '@storybook/react'
import { ThemePreference } from '../../theme'
import { ThemeProps } from '../../../../shared/src/theme'
import { WebStory } from '../../components/WebStory'

const history = createMemoryHistory()
const defaultProps = (props: ThemeProps): SearchPageProps => ({
    isSourcegraphDotCom: false,
    settingsCascade: {
        final: null,
        subjects: null,
    },
    location: history.location,
    history,
    extensionsController: {
        services: {} as Services,
    } as any,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    themePreference: ThemePreference.Light,
    onThemePreferenceChange: () => undefined,
    authenticatedUser: null,
    showCampaigns: false,
    setVersionContext: () => undefined,
    availableVersionContexts: [],
    globbing: false,
    patternType: SearchPatternType.literal,
    setPatternType: () => undefined,
    caseSensitive: false,
    setCaseSensitivity: () => undefined,
    platformContext: {} as any,
    keyboardShortcuts: [],
    filtersInQuery: {} as any,
    onFiltersInQueryChange: () => undefined,
    splitSearchModes: false,
    interactiveSearchMode: false,
    toggleSearchMode: () => undefined,
    copyQueryButton: false,
    versionContext: undefined,
    showRepogroupHomepage: false,
    showEnterpriseHomePanels: false,
    showOnboardingTour: false,
    isLightTheme: props.isLightTheme,
})

const { add } = storiesOf('web/search/input/SearchPage', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
    },
    chromatic: { viewports: [769, 993, 1200] },
})

add('Cloud without repogroups', () => (
    <WebStory>{webProps => <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} />}</WebStory>
))

add('Cloud with repogroups', () => (
    <WebStory>
        {webProps => <SearchPage {...defaultProps(webProps)} isSourcegraphDotCom={true} showRepogroupHomepage={true} />}
    </WebStory>
))

add('Server without panels', () => <WebStory>{webProps => <SearchPage {...defaultProps(webProps)} />}</WebStory>)

add('Server with panels', () => (
    <WebStory>{webProps => <SearchPage {...defaultProps(webProps)} showEnterpriseHomePanels={true} />}</WebStory>
))
