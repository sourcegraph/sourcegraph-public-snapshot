import { mdiSourceRepository } from '@mdi/js'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import * as H from 'history'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'
import { Button, H1, H2, Icon, Link } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'

import { CopyPathAction } from './actions/CopyPathAction'
import { GoToPermalinkAction } from './actions/GoToPermalinkAction'
import { FilePathBreadcrumbs } from './FilePathBreadcrumbs'
import { RepoHeader, RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoRevisionContainerBreadcrumb } from './RepoRevisionContainer'

import repoRevisionContainerStyles from './RepoRevisionContainer.module.scss'

const mockUser = {
    id: 'userID',
    username: 'username',
    email: 'user@me.com',
    siteAdmin: true,
} as AuthenticatedUser

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/RepoHeader',
    component: RepoHeader,
    decorators: [decorator],
}

export default config

export const Default: Story = () => (
    <>
        <H1>Repo header</H1>
        <H2>Simple</H2>
        <div className="mb-3 b-1">
            <Simple />
        </div>
        <H2>Long Path</H2>
        <LongPath />
    </>
)
const useActionItemsToggle = () => ({
    isOpen: false,
    toggle: () => null,
    toggleReference: () => null,
    barInPage: false,
})
const LOCATION: H.Location = {
    hash: '',
    pathname: '/github.com/sourcegraph/sourcegraph/-/tree/',
    search: '',
    state: undefined,
}
const onLifecyclePropsChange = (lifecycleProps: RepoHeaderContributionsLifecycleProps) => {
    lifecycleProps.repoHeaderContributionsLifecycleProps?.onRepoHeaderContributionAdd({
        id: 'copy-path',
        position: 'left',
        children: () => <CopyPathAction />,
    })
    lifecycleProps.repoHeaderContributionsLifecycleProps?.onRepoHeaderContributionAdd({
        id: 'go-to-permalink',
        position: 'right',
        children: () => (
            <GoToPermalinkAction
                telemetryService={NOOP_TELEMETRY_SERVICE}
                revision="main"
                commitID="123"
                location={LOCATION}
                history={H.createMemoryHistory()}
                repoName="sourcegraph/sourcegraph"
                actionType="nav"
            />
        ),
    })
}
const createBreadcrumbs = (path: string) => [
    {
        breadcrumb: {
            key: 'repository',
            element: (
                <Button
                    to="/"
                    className="text-nowrap test-repo-header-repo-link"
                    variant="secondary"
                    outline={true}
                    size="sm"
                    as={Link}
                >
                    <Icon aria-hidden={true} svgPath={mdiSourceRepository} /> sourcegraph/sourcegraph
                </Button>
            ),
        },
        depth: 1,
    },
    {
        breadcrumb: {
            key: 'revision',
            divider: <span className={repoRevisionContainerStyles.divider}>@</span>,
            element: (
                <RepoRevisionContainerBreadcrumb
                    resolvedRevision={undefined}
                    revision="main"
                    repoName="sourcegraph/sourcegraph"
                    repo={undefined}
                />
            ),
        },
        depth: 2,
    },
    {
        breadcrumb: {
            key: 'treePath',
            className: 'flex-shrink-past-contents',
            element: (
                <FilePathBreadcrumbs
                    key="path"
                    repoName="sourcegraph/sourcegraph"
                    revision="main"
                    filePath={path}
                    isDir={false}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            ),
        },
        depth: 3,
    },
]

const Simple = () => (
    <RepoHeader
        actionButtons={[]}
        useActionItemsToggle={useActionItemsToggle}
        breadcrumbs={createBreadcrumbs('client/web/src/repo/RepoHeader.story.tsx')}
        repoName="sourcegraph/sourcegraph"
        revision="main"
        onLifecyclePropsChange={onLifecyclePropsChange}
        location={LOCATION}
        history={H.createMemoryHistory()}
        settingsCascade={EMPTY_SETTINGS_CASCADE}
        authenticatedUser={mockUser}
        platformContext={{} as any}
        extensionsController={null}
        telemetryService={NOOP_TELEMETRY_SERVICE}
    />
)
const LongPath = () => (
    <RepoHeader
        actionButtons={[]}
        useActionItemsToggle={useActionItemsToggle}
        breadcrumbs={createBreadcrumbs(
            'client/web/src/repo/client/web/src/repo/client/web/src/repo/MyJavaStyleManagerReducerSuperCalifragilisticExpialidocious.tsx'
        )}
        repoName="sourcegraph/sourcegraph"
        revision="main"
        onLifecyclePropsChange={onLifecyclePropsChange}
        location={LOCATION}
        history={H.createMemoryHistory()}
        settingsCascade={EMPTY_SETTINGS_CASCADE}
        authenticatedUser={mockUser}
        platformContext={{} as any}
        extensionsController={null}
        telemetryService={NOOP_TELEMETRY_SERVICE}
    />
)

Default.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
