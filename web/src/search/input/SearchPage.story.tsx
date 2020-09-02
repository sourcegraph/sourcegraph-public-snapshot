import React from 'react'
import webStyles from '../../SourcegraphWebApp.scss'
import { Controller } from '../../../../shared/src/extensions/controller'
import { createMemoryHistory } from 'history'
import { NOOP_TELEMETRY_SERVICE } from '../../../../shared/src/telemetry/telemetryService'
import { SearchPage, SearchPageProps } from './SearchPage'
import { Services } from '../../../../shared/src/api/client/services'
import { storiesOf } from '@storybook/react'

const extensionsController = {
    services: {} as Services,
    executeCommand: () => Promise.resolve(undefined),
} as Pick<Controller, 'executeCommand' | 'services'>

const history = createMemoryHistory()
const defaultProps = {
    isSourcegraphDotCom: false,
    settingsCascade: {
        final: null,
        subjects: null,
    },
    location: history.location,
    extensionsController,
    telemetryService: NOOP_TELEMETRY_SERVICE,
} as SearchPageProps

const { add } = storiesOf('web/search/input/SearchPage', module)
    .addParameters({
        percy: { widths: [993] },
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
        chromatic: { viewports: [769, 993, 1200] },
    })
    .addDecorator(story => (
        <>
            <style>{webStyles}</style>
            <div className="theme-light">{story()}</div>
        </>
    ))

add('Cloud without repogroups', () => <SearchPage {...defaultProps} isSourcegraphDotCom={true} />)

add('Cloud with repogroups', () => (
    <SearchPage {...defaultProps} isSourcegraphDotCom={true} showRepogroupHomepage={true} />
))

add('Server without panels', () => <SearchPage {...defaultProps} />)

add('Server with panels', () => <SearchPage {...defaultProps} showEnterpriseHomePanels={true} />)
