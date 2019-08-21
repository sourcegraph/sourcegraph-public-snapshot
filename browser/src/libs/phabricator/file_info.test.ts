import { readFile } from 'mz/fs'
import { Observable, throwError, of } from 'rxjs'
import { resolveDiffusionFileInfo, resolveRevisionFileInfo, resolveDiffFileInfo } from './file_info'
import { GraphQLResponseMap, mockRequestGraphQL } from '../code_intelligence/test_helpers'
import { QueryConduitHelper } from './backend'
import { SuccessGraphQLResult } from '../../../../shared/src/graphql/graphql'
import { IMutation, IQuery } from '../../../../shared/src/graphql/schema'
import { resetAllMemoizationCaches } from '../../../../shared/src/util/memoizeObservable'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { FileInfo } from '../code_intelligence'

interface ConduitResponseMap {
    [endpoint: string]: (params: { [key: string]: any }) => Observable<any>
}

const DEFAULT_CONDUIT_RESPONSES: ConduitResponseMap = {
    '/api/diffusion.repository.search': () =>
        of({
            data: [
                {
                    fields: {
                        callsign: 'MUX',
                    },
                    attachments: {
                        uris: {
                            uris: [
                                {
                                    fields: {
                                        uri: {
                                            raw: 'https://github.com/gorilla/mux',
                                            normalized: 'https://github.com/gorilla/mux',
                                        },
                                    },
                                },
                            ],
                        },
                    },
                },
            ],
        }),
    '/api/differential.query': () =>
        of({
            0: {
                repositoryPHID: '1',
            },
        }),
    '/api/differential.querydiffs': params =>
        of({
            [params.ids[0]]: {
                id: params.ids[0],
                revisionID: params.revisionIDs[0],
                dateCreated: '1566329300',
                dateModified: '1566329305',
                sourceControlBaseRevision: 'base-revision',
                branch: 'test',
                description: '  - test',
                changes: [
                    {
                        currentPath: 'helpers/add.go',
                    },
                    {
                        currentPath: '.arcconfig',
                    },
                ],
                properties: {
                    'arc.staging': {
                        status: 'pushed',
                        refs: [
                            {
                                ref: `refs/tags/phabricator/base/${params.ids[0]}`,
                                type: 'base',
                                commit: `base-${params.ids[0]}`,
                                remote: { uri: 'https://github.com/lguychard/testing.git' },
                            },
                            {
                                ref: `refs/tags/phabricator/diff/${params.ids[0]}`,
                                type: 'diff',
                                commit: `diff-${params.ids[0]}`,
                                remote: { uri: 'https://github.com/lguychard/testing.git' },
                            },
                        ],
                    },
                },
                authorName: 'Loïc Guychard',
                authorEmail: 'loic@sourcegraph.com',
            },
        }),
    '/api/differential.getrawdiff': () => of('diff'),
}

const DEFAULT_GRAPHQL_RESPONSES: GraphQLResponseMap = {
    addPhabricatorRepo: () =>
        of({
            data: {},
            errors: undefined,
        } as SuccessGraphQLResult<IMutation>),
    ResolveRepo: () =>
        of({
            data: {
                repository: null,
            },
            errors: undefined,
        } as SuccessGraphQLResult<IQuery>),
    ResolveStagingRev: () =>
        of({
            data: { resolvePhabricatorDiff: { oid: 'staging-rev' } },
            errors: undefined,
        } as SuccessGraphQLResult<IMutation>),
}

function mockQueryConduit(responseMap?: ConduitResponseMap): QueryConduitHelper<any> {
    return (endpoint, params) => {
        const mock = (responseMap && responseMap[endpoint]) || DEFAULT_CONDUIT_RESPONSES[endpoint]
        if (!mock) {
            return throwError(new Error(`No mock for endpoint ${endpoint}`))
        }
        return mock(params)
    }
}

type Resolver = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit: QueryConduitHelper<any>
) => Observable<FileInfo>

interface Fixture {
    name: string
    htmlFixture: string
    url: string
    codeViewSelector: string
    graphQLResponseMap?: GraphQLResponseMap
    conduitResponseMap?: ConduitResponseMap
}

const runFixture = (
    { name, url, htmlFixture, codeViewSelector, graphQLResponseMap, conduitResponseMap }: Fixture,
    resolver: Resolver
): void => {
    test(name, async () => {
        const fixtureContent = await readFile(`${__dirname}/__fixtures__/pages/${htmlFixture}`, 'utf-8')
        document.body.innerHTML = fixtureContent
        jsdom.reconfigure({ url })
        const codeView = document.querySelector(codeViewSelector)
        if (!codeView) {
            throw new Error(`Code view matching selector ${codeViewSelector} not found`)
        }
        const fileInfo = await resolver(
            codeView as HTMLElement,
            mockRequestGraphQL({
                ...DEFAULT_GRAPHQL_RESPONSES,
                ...(graphQLResponseMap || {}),
            }),
            mockQueryConduit(conduitResponseMap)
        ).toPromise()
        expect(fileInfo).toMatchSnapshot()
    })
}

describe('Phabricator file info', () => {
    beforeEach(() => {
        resetAllMemoizationCaches()
    })

    describe('resolveRevisionFileInfo()', () => {
        runFixture(
            {
                name: 'Commit view',
                htmlFixture: 'commit-view.html',
                url: 'https://phabricator.sgdev.org/rMUXeab9c4f3d22d907d728aa0f5918934357866249e',
                codeViewSelector: '.differential-changeset',
            },
            resolveRevisionFileInfo
        )
    })

    describe('resolveDiffusionFileInfo()', () => {
        runFixture(
            {
                name: 'Diffusion - single file code view',
                htmlFixture: 'diffusion.html',
                url: 'https://phabricator.sgdev.org/source/gorilla/browse/master/mux.go',
                codeViewSelector: '.diffusion-source',
            },
            resolveDiffusionFileInfo
        )
    })

    describe('resolveDiffFileInfo()', () => {
        const fixtures: Fixture[] = [
            {
                name: 'Differential revision - no staging repo',
                htmlFixture: 'differential-revision.html',
                url: 'https://phabricator.sgdev.org/D7',
                codeViewSelector: '.differential-changeset',
                conduitResponseMap: {
                    // Returns diff details without staging details
                    '/api/differential.querydiffs': params =>
                        of({
                            [params.ids[0]]: {
                                id: params.ids[0],
                                revisionID: params.revisionIDs[0],
                                dateCreated: '1566329300',
                                dateModified: '1566329305',
                                sourceControlBaseRevision: 'base-revision',
                                branch: 'test',
                                description: '  - test',
                                changes: [],
                                properties: {},
                                authorName: 'Loïc Guychard',
                                authorEmail: 'loic@sourcegraph.com',
                            },
                        }),
                },
            },
            {
                name: 'Differential revision - staging repo not synced',
                htmlFixture: 'differential-revision.html',
                url: 'https://phabricator.sgdev.org/D7',
                codeViewSelector: '.differential-changeset',
            },
            {
                name: 'Differential revision - staging repo synced',
                htmlFixture: 'differential-revision.html',
                url: 'https://phabricator.sgdev.org/D7',
                codeViewSelector: '.differential-changeset',
                graphQLResponseMap: {
                    // Echoes the raw repo name, to represent the fact that the repository
                    // exists on the Sourcegraph instance.
                    ResolveRepo: variables =>
                        of({
                            data: {
                                repository: {
                                    name: variables.rawRepoName,
                                },
                            },
                            errors: undefined,
                        } as SuccessGraphQLResult<IQuery>),
                },
            },
            {
                name: 'Differential revision - comparing diffs - staging repo not synced',
                htmlFixture: 'differential-diff-comparison.html',
                url: 'https://phabricator.sgdev.org/D1?vs=2&id=3&whitespace=ignore-most#toc',
                codeViewSelector: '.differential-changeset',
                graphQLResponseMap: {
                    ResolveStagingRev: variables =>
                        of({
                            data: { resolvePhabricatorDiff: { oid: `staging-rev-${variables.patch}` } },
                            errors: undefined,
                        } as SuccessGraphQLResult<IMutation>),
                },
                conduitResponseMap: {
                    '/api/differential.getrawdiff': params => of(`raw-diff-for-diffid-${params.diffID}`),
                },
            },
            {
                name: 'Differential revision - comparing diffs - staging repo synced',
                htmlFixture: 'differential-diff-comparison.html',
                url: 'https://phabricator.sgdev.org/D1?vs=2&id=3&whitespace=ignore-most#toc',
                codeViewSelector: '.differential-changeset',
                graphQLResponseMap: {
                    // Echoes the raw repo name, to represent the fact that the repository
                    // exists on the Sourcegraph instance.
                    ResolveRepo: variables =>
                        of({
                            data: {
                                repository: {
                                    name: variables.rawRepoName,
                                },
                            },
                            errors: undefined,
                        } as SuccessGraphQLResult<IQuery>),
                },
            },
        ]
        for (const fixture of fixtures) {
            runFixture(fixture, resolveDiffFileInfo)
        }
    })
})
