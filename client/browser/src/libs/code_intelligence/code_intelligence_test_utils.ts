import { readFile } from 'mz/fs'
import { SetIntersection } from 'utility-types'
import { CodeHost } from './code_intelligence'

const mountGetterKeys = [
    'getCommandPaletteMount',
    'getGlobalDebugMount',
    'getOverlayMount',
    'getViewContextOnSourcegraphMount',
] as const
type MountGetterKey = (typeof mountGetterKeys)[number]

export function testMountGetterInvariants<C extends CodeHost>(
    codeHost: C,
    containerHTMLPath: string | Record<SetIntersection<MountGetterKey, keyof C>, string>
): void {
    for (const mountGetterKey of mountGetterKeys) {
        const getMount = codeHost[mountGetterKey]
        if (!getMount) {
            continue
        }
        describe(mountGetterKey, () => {
            const getFixtureBody = async () => {
                // Do not append use global document to test that the mount getter only looks at the container
                const fixtureDocument = document.implementation.createHTMLDocument()
                const fixturePath =
                    typeof containerHTMLPath === 'string'
                        ? containerHTMLPath
                        : containerHTMLPath[mountGetterKey as keyof typeof containerHTMLPath]
                fixtureDocument.write(await readFile(fixturePath, 'utf-8'))
                return fixtureDocument.body
            }
            it('returns a mount in the container', async () => {
                const container = await getFixtureBody()
                const mount = getMount(container)
                expect(mount).toBeInstanceOf(HTMLElement)
                expect(container.contains(mount)).toBe(true)
            })
            it('is idempotent', async () => {
                const container = await getFixtureBody()
                const first = getMount(container)
                const outerHTMLAfterFirstCall = container.outerHTML
                const second = getMount(container)
                expect(first).toBe(second)
                expect(container.outerHTML).toBe(outerHTMLAfterFirstCall)
            })
            it('returns null if the mount does not belong into the container', async () => {
                const container = document.createElement('div')
                container.innerHTML = '<div>Hello</div><div>World</div>'
                const mount = getMount(container)
                expect(mount).toBe(null)
            })
        })
    }
}
