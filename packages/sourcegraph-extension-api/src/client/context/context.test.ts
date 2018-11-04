import assert from 'assert'
import { EMPTY_ENVIRONMENT, Environment } from '../environment'
import { applyContextUpdate, Context, getComputedContextProperty } from './context'

describe('applyContextUpdate', () => {
    it('merges properties', () =>
        assert.deepStrictEqual(applyContextUpdate({ a: 1, b: null, c: 2, d: 3, e: null }, { a: null, b: 1, c: 3 }), {
            b: 1,
            c: 3,
            d: 3,
            e: null,
        } as Context))
})

describe('getComputedContextProperty', () => {
    it('provides config', () => {
        const env: Environment = {
            ...EMPTY_ENVIRONMENT,
            configuration: {
                merged: {
                    a: 1,
                    'a.b': 2,
                    'c.d': 3,
                },
            },
        }
        assert.strictEqual(getComputedContextProperty(env, 'config.a'), 1)
        assert.strictEqual(getComputedContextProperty(env, 'config.a.b'), 2)
        assert.strictEqual(getComputedContextProperty(env, 'config.c.d'), 3)
        assert.strictEqual(getComputedContextProperty(env, 'config.x'), null)
    })

    describe('environment with component', () => {
        const env: Environment = {
            ...EMPTY_ENVIRONMENT,
            visibleTextDocuments: [
                {
                    uri: 'file:///a/b.c',
                    languageId: 'l',
                    text: 't',
                },
            ],
        }

        describe('resource', () => {
            it('provides resource.uri', () =>
                assert.strictEqual(getComputedContextProperty(env, 'resource.uri'), 'file:///a/b.c'))
            it('provides resource.basename', () =>
                assert.strictEqual(getComputedContextProperty(env, 'resource.basename'), 'b.c'))
            it('provides resource.dirname', () =>
                assert.strictEqual(getComputedContextProperty(env, 'resource.dirname'), 'file:///a'))
            it('provides resource.extname', () =>
                assert.strictEqual(getComputedContextProperty(env, 'resource.extname'), '.c'))
            it('provides resource.language', () =>
                assert.strictEqual(getComputedContextProperty(env, 'resource.language'), 'l'))
            it('provides resource.textContent', () =>
                assert.strictEqual(getComputedContextProperty(env, 'resource.textContent'), 't'))
            it('provides resource.type', () =>
                assert.strictEqual(getComputedContextProperty(env, 'resource.type'), 'textDocument'))

            it('returns undefined when the environment has no component', () =>
                assert.strictEqual(getComputedContextProperty(EMPTY_ENVIRONMENT, 'resource.uri'), undefined))
        })

        describe('component', () => {
            it('provides component.type', () =>
                assert.strictEqual(getComputedContextProperty(env, 'component.type'), 'textEditor'))

            it('returns undefined when the environment has no component', () =>
                assert.strictEqual(getComputedContextProperty(EMPTY_ENVIRONMENT, 'component.type'), undefined))
        })
    })

    it('falls back to the environment context', () => {
        assert.strictEqual(getComputedContextProperty({ ...EMPTY_ENVIRONMENT, context: { x: 1 } }, 'x'), 1)
        assert.strictEqual(getComputedContextProperty({ ...EMPTY_ENVIRONMENT, context: {} }, 'y'), undefined)
    })
})
