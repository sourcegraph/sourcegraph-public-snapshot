import assert from 'assert'
import { readFile } from 'mz/fs'
import { SetIntersection } from 'utility-types'
import { CodeHost, CodeViewSpec, MountGetter } from './code_intelligence'

const mountGetterKeys = [
    'getCommandPaletteMount',
    'getGlobalDebugMount',
    'getOverlayMount',
    'getViewContextOnSourcegraphMount',
] as const
type MountGetterKey = (typeof mountGetterKeys)[number]

export function testCodeHostMountGetters<C extends CodeHost>(
    codeHost: C,
    containerHtmlFixturePaths: string | Record<SetIntersection<MountGetterKey, keyof C>, string>
): void {
    for (const mountGetterKey of mountGetterKeys) {
        const getMount = codeHost[mountGetterKey]
        if (!getMount) {
            continue
        }
        describe(mountGetterKey, () => {
            const fixturePath =
                typeof containerHtmlFixturePaths === 'string'
                    ? containerHtmlFixturePaths
                    : containerHtmlFixturePaths[mountGetterKey as keyof typeof containerHtmlFixturePaths]
            testMountGetter(fixturePath, getMount, true, true)
        })
    }
}

export function testToolbarMountGetter(
    codeViewHtmlFixturePath: string,
    getToolbarMount: NonNullable<CodeViewSpec['getToolbarMount']>
): void {
    testMountGetter(codeViewHtmlFixturePath, getToolbarMount, false, false)
}

/**
 * @param isFullDocument Whether the fixture HTML is a document fragment or a full document (does it contain `<html>`?)
 * @param mayReturnNull Whether the mount getter might be called with containers where the mount does not belong.
 * It will be tested to return `null` in that case.
 */
function testMountGetter(
    containerHTMLPath: string,
    getMount: MountGetter,
    isFullDocument: boolean,
    mayReturnNull: boolean
): void {
    const getFixtureBody = async () => {
        const content = await readFile(containerHTMLPath, 'utf-8')
        // Do not append use global document to test that the mount getter only looks at the container
        if (isFullDocument) {
            // Create Document
            const fixtureDocument = document.implementation.createHTMLDocument()
            fixtureDocument.write(content)
            return fixtureDocument.documentElement
        }
        // Create DocumentFragment
        const template = document.createElement('template')
        template.innerHTML = content
        if (template.content.children.length !== 1) {
            throw new Error(
                `Fixture must have exactly one element, has ${template.content.children.length}: ${containerHTMLPath}`
            )
        }
        if (!(template.content.firstElementChild instanceof HTMLElement)) {
            throw new Error(`Fixture must be HTML: ${containerHTMLPath}`)
        }
        return template.content.firstElementChild
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
    if (mayReturnNull) {
        it('returns null if the mount does not belong into the container', async () => {
            const container = document.createElement('div')
            container.innerHTML = '<div>Hello</div><div>World</div>'
            const mount = getMount(container)
            expect(mount).toBe(null)
        })
    } else {
        it('throws an Error if given an unexpected code view', () => {
            const container = document.createElement('div')
            container.innerHTML = '<div>Hello</div><div>World</div>'
            try {
                getMount(container)
                assert.fail('Expected function to throw an Error')
            } catch (err) {
                // "Cannot read property foo of null" does not count!
                expect(err).not.toBeInstanceOf(TypeError)
            }
        })
    }
}
