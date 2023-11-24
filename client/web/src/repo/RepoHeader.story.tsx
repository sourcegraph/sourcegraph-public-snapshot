import { mdiSourceRepository } from '@mdi/js'
import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { CopyPathAction } from '@sourcegraph/branded'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { Button, H1, H2, Icon, Link } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import type { AuthenticatedUser } from '../auth'

import { GoToPermalinkAction } from './actions/CopyPermalinkAction'
import { FilePathBreadcrumbs } from './FilePathBreadcrumbs'
import { RepoHeader, type RepoHeaderContributionsLifecycleProps } from './RepoHeader'
import { RepoRevisionContainerBreadcrumb } from './RepoRevisionContainer'

import webStyles from '../SourcegraphWebApp.scss'
import repoRevisionContainerStyles from './RepoRevisionContainer.module.scss'

const mockUser = {
    id: 'userID',
    username: 'username',
    emails: [{ email: 'user@me.com', isPrimary: true, verified: true }],
    siteAdmin: true,
} as AuthenticatedUser

const decorator: Decorator = story => (
    <BrandedStory initialEntries={['/github.com/sourcegraph/sourcegraph/-/tree/']} styles={webStyles}>
        {() => <div className="container mt-3">{story()}</div>}
    </BrandedStory>
)

const config: Meta = {
    title: 'wildcard/RepoHeader',
    component: RepoHeader,
    decorators: [decorator],
}

export default config

export const Default: StoryFn = () => (
    <>
        <H1>Repo header</H1>
        <H2>Simple</H2>
        <div className="mb-3 b-1">
            <RepoHeader {...createProps('client/web/src/repo/RepoHeader.story.tsx')} />
        </div>
        <H2>Constrained width</H2>
        <div className="mb-3 b-1" style={{ maxWidth: 480 }}>
            <RepoHeader {...createProps('client/web/src/repo/RepoHeader.story.tsx', true)} />
        </div>
        <H2>Long path</H2>
        <RepoHeader
            {...createProps(
                'client/web/src/repo/client/web/src/repo/client/web/src/repo/MyJavaStyleManagerReducerSuperCalifragilisticExpialidocious.tsx'
            )}
        />
        <H2>Many subfolders</H2>
        <RepoHeader {...createProps('client/web/src/repo/client/web/src/repo/client/web/src/repo/main.tsx')} />
        <H2>Many subfolders and constrained width</H2>
        <div className="mb-3 b-1" style={{ maxWidth: 480 }}>
            <RepoHeader
                {...createProps('client/web/src/repo/client/web/src/repo/client/web/src/repo/main.tsx', true)}
            />
        </div>
    </>
)

const onLifecyclePropsChange = (lifecycleProps: RepoHeaderContributionsLifecycleProps) => {
    lifecycleProps.repoHeaderContributionsLifecycleProps?.onRepoHeaderContributionAdd({
        id: 'copy-path',
        position: 'left',
        children: () => <CopyPathAction filePath="foobar" telemetryService={NOOP_TELEMETRY_SERVICE} />,
    })
    lifecycleProps.repoHeaderContributionsLifecycleProps?.onRepoHeaderContributionAdd({
        id: 'go-to-permalink',
        position: 'right',
        children: () => (
            <GoToPermalinkAction
                telemetryService={NOOP_TELEMETRY_SERVICE}
                revision="main"
                commitID="123"
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
            className: 'flex-shrink-past-contents flex-grow-1',
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

const createProps = (path: string, forceWrap: boolean = false): React.ComponentProps<typeof RepoHeader> => ({
    breadcrumbs: createBreadcrumbs(path),
    repoName: 'sourcegraph/sourcegraph',
    revision: 'main',
    onLifecyclePropsChange,
    settingsCascade: EMPTY_SETTINGS_CASCADE,
    authenticatedUser: mockUser,
    platformContext: {} as any,
    forceWrap,
})

Default.parameters = {
    chromatic: {
        enableDarkMode: false,
        disableSnapshot: false,
    },
}
