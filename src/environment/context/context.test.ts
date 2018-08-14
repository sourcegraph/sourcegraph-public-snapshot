import assert from 'assert'
import { EMPTY_ENVIRONMENT } from '../environment'
import { Context, contextFilter, createChildContext, environmentContext } from './context'

const throwContext: Context = {
    get(): any {
        throw new Error('no next context')
    },
}

describe('createChildContext', () => {
    it('creates a child context', () => {
        const parent: Context = new Map([['a', 1]])
        const child = createChildContext(parent)
        assert.strictEqual(child.get('a'), 1)
        assert.strictEqual(child.get('b'), undefined)
        child.set('b', 2)
        assert.strictEqual(child.get('b'), 2)
        child.set('a', 3)
        assert.strictEqual(child.get('a'), 3)
        assert.strictEqual(parent.get('a'), 1)
        child.set('a', undefined)
        assert.strictEqual(child.get('a'), 1)
        assert.strictEqual(parent.get('a'), 1)
    })
})

describe('environmentContext', () => {
    it('provides config', () => {
        const context = environmentContext(
            {
                ...EMPTY_ENVIRONMENT,
                configuration: {
                    merged: {
                        a: 1,
                        'a.b': 2,
                        'c.d': 3,
                    },
                },
            },
            throwContext
        )
        assert.strictEqual(context.get('config.a'), 1)
        assert.strictEqual(context.get('config.a.b'), 2)
        assert.strictEqual(context.get('config.c.d'), 3)
        assert.strictEqual(context.get('config.x'), null)
    })

    describe('environment with component', () => {
        const context = environmentContext(
            {
                ...EMPTY_ENVIRONMENT,
                component: {
                    document: {
                        uri: 'file:///a/b.c',
                        languageId: 'l',
                        version: 0,
                        text: 't',
                    },
                    selections: [],
                    visibleRanges: [],
                },
            },
            throwContext
        )

        describe('resource', () => {
            it('provides resource.uri', () => assert.strictEqual(context.get('resource.uri'), 'file:///a/b.c'))
            it('provides resource.basename', () => assert.strictEqual(context.get('resource.basename'), 'b.c'))
            it('provides resource.dirname', () => assert.strictEqual(context.get('resource.dirname'), 'file:///a'))
            it('provides resource.extname', () => assert.strictEqual(context.get('resource.extname'), '.c'))
            it('provides resource.language', () => assert.strictEqual(context.get('resource.language'), 'l'))
            it('provides resource.textContent', () => assert.strictEqual(context.get('resource.textContent'), 't'))
            it('provides resource.type', () => assert.strictEqual(context.get('resource.type'), 'textDocument'))

            it('returns undefined when the environment has no component', () =>
                assert.strictEqual(environmentContext(EMPTY_ENVIRONMENT, throwContext).get('resource.uri'), undefined))
        })

        describe('component', () => {
            it('provides component.type', () => assert.strictEqual(context.get('component.type'), 'textEditor'))

            it('returns undefined when the environment has no component', () =>
                assert.strictEqual(
                    environmentContext(EMPTY_ENVIRONMENT, throwContext).get('component.type'),
                    undefined
                ))
        })
    })

    it('falls back to the next context', () => {
        assert.strictEqual(environmentContext(EMPTY_ENVIRONMENT, { get: () => 1 }).get('x'), 1)
        assert.throws(() => environmentContext(EMPTY_ENVIRONMENT, throwContext).get('x'))
    })
})

describe('contextFilter', () => {
    const FIXTURE_CONTEXT = new Map<string, any>(
        Object.entries({
            a: 1,
            b: 1,
            c: 2,
            x: 'y',
        })
    )

    it('filters', () =>
        assert.deepStrictEqual(
            contextFilter(FIXTURE_CONTEXT, [
                { x: 1 },
                { x: 2, when: 'a' },
                { x: 3, when: 'a == b' },
                { x: 4, when: 'a == c' },
            ]),
            [{ x: 1 }, { x: 2, when: 'a' }, { x: 3, when: 'a == b' }]
        ))
})
