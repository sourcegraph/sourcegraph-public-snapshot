import { resolveDiffusionFileInfo, resolveRevisionFileInfo, resolveDiffFileInfo } from './file_info'
import { mockRequestGraphQL } from '../code_intelligence/test_helpers'
import { QueryConduitHelper } from './backend'
import { Observable, throwError, of } from 'rxjs'
import { readFile } from 'mz/fs'

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

const loadFixture = async (pageName: string) => {
    const fixtureContent = await readFile(`${__dirname}/__fixtures__/pages/${pageName}`, 'utf-8')
    document.body.innerHTML = fixtureContent
}

const requestGraphQL = mockRequestGraphQL({
    addPhabricatorRepo: () =>
        of({
            data: {} as any,
            errors: undefined,
        }),
})

describe('Phabricator file info', () => {
    describe('resolveRevisionFileInfo()', () => {
        test('commit view', async () => {
            await loadFixture('commit-view.html')
            jsdom.reconfigure({
                url: 'https://phabricator.sgdev.org/rMUXeab9c4f3d22d907d728aa0f5918934357866249e',
            })
            const fileInfo = await resolveRevisionFileInfo(
                document.querySelector('.differential-changeset')! as HTMLElement,
                requestGraphQL,
                mockQueryConduit()
            ).toPromise()
            expect(fileInfo).toMatchSnapshot()
        })
    })

    describe('resolveDiffFileInfo()', () => {
        test.todo('Differential revision - no staging repo')

        test('Differential revision - staging repo not synced on Sourcegraph instance', async () => {
            await loadFixture('differential-revision.html')
            jsdom.reconfigure({
                url: 'https://phabricator.sgdev.org/D7',
            })
            const fileInfo = await resolveDiffFileInfo(
                document.querySelector('.differential-changeset')! as HTMLElement,
                requestGraphQL,
                mockQueryConduit({
                    '/api/differential.querydiffs': params =>
                        of({
                            [params.ids[0]]: {
                                branch: 'a',
                                sourceControlBaseRevision: 'b',
                                description: 'c',
                                changes: [],
                                dateCreated: '2013-07-08',
                                authorName: 'd',
                                authorEmail: 'e',
                                properties: {
                                    'arc.staging': {
                                        status:
                                    }
                                }
                            },
                        }),
                })
            ).toPromise()
            expect(fileInfo).toMatchSnapshot()
        })

        test.todo('Differential revision - staging repo synced on Sourcegraph instance')

        test.todo('Differential revision - comparing two diff IDs')
    })

    describe('resolveDiffusionFileInfo', () => {
        test('Diffusion - single file coode view', async () => {
            await loadFixture('diffusion.html')
            jsdom.reconfigure({
                url: 'https://phabricator.sgdev.org/source/gorilla/browse/master/mux.go',
            })
            const fileInfo = await resolveDiffusionFileInfo(
                document.querySelector('.diffusion-source')! as HTMLElement,
                requestGraphQL,
                mockQueryConduit()
            ).toPromise()
            expect(fileInfo).toMatchSnapshot()
        })
    })
})
