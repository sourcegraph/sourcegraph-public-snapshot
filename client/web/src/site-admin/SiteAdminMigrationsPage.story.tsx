import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { type Observable, of } from 'rxjs'

import { WebStory } from '../components/WebStory'
import type { OutOfBandMigrationFields } from '../graphql-operations'

import { SiteAdminMigrationsPage } from './SiteAdminMigrationsPage'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/Site Admin/Migrations',
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
    decorators: [decorator],
}

export default config

// invalid, pre-introduction
export const InvalidPreIntroduction: StoryFn = () => (
    <WebStory>
        {props => (
            <SiteAdminMigrationsPage
                {...props}
                fetchAllMigrations={(): Observable<OutOfBandMigrationFields[]> => of(migrations)}
                fetchSiteUpdateCheck={() => of({ productVersion: '3.23.2' })}
                now={now}
            />
        )}
    </WebStory>
)

InvalidPreIntroduction.storyName = '3.23.2'

// downgrade warning
export const DowngradeWarning: StoryFn = () => (
    <WebStory>
        {props => (
            <SiteAdminMigrationsPage
                {...props}
                fetchAllMigrations={(): Observable<OutOfBandMigrationFields[]> => of(migrations)}
                fetchSiteUpdateCheck={() => of({ productVersion: '3.24.2' })}
                now={now}
            />
        )}
    </WebStory>
)

DowngradeWarning.storyName = '3.24.2'

// no warnings
export const NoWarnings: StoryFn = () => (
    <WebStory>
        {props => (
            <SiteAdminMigrationsPage
                {...props}
                fetchAllMigrations={(): Observable<OutOfBandMigrationFields[]> => of(migrations)}
                fetchSiteUpdateCheck={() => of({ productVersion: '3.25.2' })}
                now={now}
            />
        )}
    </WebStory>
)

NoWarnings.storyName = '3.25.2'

// upgrade warning
export const UpgradeWarning: StoryFn = () => (
    <WebStory>
        {props => (
            <SiteAdminMigrationsPage
                {...props}
                fetchAllMigrations={(): Observable<OutOfBandMigrationFields[]> => of(migrations)}
                fetchSiteUpdateCheck={() => of({ productVersion: '3.26.2' })}
                now={now}
            />
        )}
    </WebStory>
)

UpgradeWarning.storyName = '3.26.2'

// invalid, post-deprecation
export const InvalidPostDeprecation: StoryFn = () => (
    <WebStory>
        {props => (
            <SiteAdminMigrationsPage
                {...props}
                fetchAllMigrations={(): Observable<OutOfBandMigrationFields[]> => of(migrations)}
                fetchSiteUpdateCheck={() => of({ productVersion: '3.27.2' })}
                now={now}
            />
        )}
    </WebStory>
)

InvalidPostDeprecation.storyName = '3.27.2'

const migrations = [
    {
        id: 'migration-a',
        team: 'code-intelligence',
        component: 'lsif_data_documents',
        description: 'Denormalize diagnostic counts',
        introduced: '3.23',
        deprecated: '3.27',
        progress: 1,
        created: '2020-12-20T12:00+00:00',
        lastUpdated: '2020-12-20T14:00+00:00',
        nonDestructive: true,
        applyReverse: false,
        errors: [],
    },
    {
        id: 'migration-b',
        team: 'search',
        component: 'zoekt indexes',
        description: 'Apply rot13 to zoekt shards',
        introduced: '3.24',
        deprecated: '',
        progress: 0.31,
        created: '2021-01-20T12:00+00:00',
        lastUpdated: '2021-03-05T11:59:45+00:00',
        nonDestructive: false,
        applyReverse: false,
        errors: [
            { message: 'uh-oh 4', created: '2021-03-05T11:59:45+00:00' },
            { message: 'uh-oh 3', created: '2021-01-25T16:00+00:00' },
            { message: 'uh-oh 2', created: '2021-01-25T15:00+00:00' },
            { message: 'uh-oh 1', created: '2021-01-25T14:00+00:00' },
        ],
    },
    {
        id: 'migration-c',
        team: 'code-intelligence',
        component: 'main-db',
        description: 'Compress lsif_nearest_uploads',
        introduced: '3.25',
        deprecated: '3.27',
        progress: 0,
        created: '2021-01-20T12:00+00:00',
        lastUpdated: '2021-01-26T12:00+00:00',
        nonDestructive: false,
        applyReverse: true,
        errors: [],
    },
    {
        id: 'migration-d',
        team: 'code-insights',
        component: 'codeinsights-db',
        description: 'Reticulate the hypertables',
        introduced: '3.26',
        deprecated: '',
        progress: 0,
        created: '2021-02-20T12:00+00:00',
        lastUpdated: '2021-02-25T12:00+00:00',
        nonDestructive: false,
        applyReverse: true,
        errors: [],
    },
    {
        id: 'migration-e',
        team: 'moonshots',
        component: 'gitcoin',
        description: 'Use blockchain as a cache?',
        introduced: '3.33',
        deprecated: '',
        progress: 0,
        created: '2021-03-20T12:00+00:00',
        lastUpdated: '',
        nonDestructive: false,
        applyReverse: false,
        errors: [],
    },
]

const now = () => new Date('2021-03-05T12:00:00+00:00')
