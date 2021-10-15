import { TestScheduler } from 'rxjs/testing'

import { CLOUD_SOURCEGRAPH_URL } from '../../util/context'

import { SourcegraphUrlService } from '.'

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('SourcegraphUrlService', () => {
    describe('[isExtension=false]', () => {
        afterEach(() => {
            delete window.SOURCEGRAPH_URL
            window.localStorage.removeItem('SOURCEGRAPH_URL')
        })

        it('returns correct URL for window.SOURCEGRAPH_URL', () => {
            window.SOURCEGRAPH_URL = 'mock_url'
            scheduler().run(({ expectObservable }) => {
                expectObservable(SourcegraphUrlService.observe(false)).toBe('(a|)', {
                    a: window.SOURCEGRAPH_URL,
                })
            })
        })

        it('returns correct URL for window.localStorage', () => {
            localStorage.setItem('SOURCEGRAPH_URL', 'local_storage_mock')
            scheduler().run(({ expectObservable }) => {
                expectObservable(SourcegraphUrlService.observe(false)).toBe('(a|)', {
                    a: localStorage.getItem('SOURCEGRAPH_URL'),
                })
            })
        })

        it('returns correct URL for CLOUD_SOURCEGRAPH_URL', () => {
            scheduler().run(({ expectObservable }) => {
                expectObservable(SourcegraphUrlService.observe(false)).toBe('(a|)', {
                    a: CLOUD_SOURCEGRAPH_URL,
                })
            })
        })
    })
    describe('[isExtension=true]', () => {
        it('returns correct initial URL', () => {
            scheduler().run(({ expectObservable }) => {
                expectObservable(SourcegraphUrlService.observe(false)).toBe('(a|)', {
                    a: CLOUD_SOURCEGRAPH_URL,
                })
            })
        })
    })
})
