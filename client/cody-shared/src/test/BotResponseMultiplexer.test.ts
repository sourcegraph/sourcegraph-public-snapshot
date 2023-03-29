import assert from 'assert'

import { BotResponseMultiplexer } from '../../chat/BotResponseMultiplexer'

function promise<T>(): [(value: T) => void, Promise<T>] {
    let resolver
    let promise = new Promise<T>(resolve => resolver = resolve)
    if (!resolver) {
        throw new Error('unreachable')
    }
    return [resolver, promise]
}

describe('BotResponseMultiplexer', () => {
    it('routes messages with no prefix to the default topic', async () => {
        const multiplexer = new BotResponseMultiplexer()
        const [published, publishedResult] = promise<void>()
        multiplexer.sub(BotResponseMultiplexer.DEFAULT_TOPIC, {
            async onResponse(topic, content): Promise<void> {
                assert.strictEqual(topic, BotResponseMultiplexer.DEFAULT_TOPIC)
                assert.strictEqual(content, 'hello, world')
                published(void(0))
            },
        })
        await multiplexer.publish('hello, world')
        await publishedResult
    })

    it('discards messages when there is no subscriber', async () => {
        const multiplexer = new BotResponseMultiplexer()
        await multiplexer.publish('is this thing on?')
    })

    it('routes messages with an unknown prefix to the default subscriber', async () => {
        const multiplexer = new BotResponseMultiplexer()
        const [published, publishedResult] = promise<void>()
        multiplexer.sub(BotResponseMultiplexer.DEFAULT_TOPIC, {
            async onResponse(topic, content): Promise<void> {
                assert.strictEqual(topic, BotResponseMultiplexer.DEFAULT_TOPIC)
                assert.strictEqual(content, "orator: I'm speechless")
                published(void(0))
            },
        })
        await multiplexer.publish("orator: I'm speechless")
        await publishedResult
    })

    it('routes messages to subscribers', async () => {
        const multiplexer = new BotResponseMultiplexer()
        multiplexer.sub('cashier', {
            async onResponse(_, content) {
                assert.deepStrictEqual(content, ' one double tall latte please\nand a donut\n')
            }
        })
        multiplexer.sub('barista', {
            async onResponse(_, content) {
                assert.deepStrictEqual(content, ' can I get that to go?')
             }
        })
        await multiplexer.publish(`cashier: one double tall latte please
and a donut
barista: can I get that to go?`)
    })

    it('can route to specific subscribers and the default subscriber', async () => {
        const multiplexer = new BotResponseMultiplexer()

        const conspiracyTopic: string[] = []
        multiplexer.sub('deep-state', {
            async onResponse(_, content) {
                conspiracyTopic.push(content)
            }
        })

        const defaultTopic: string[] = []
        multiplexer.sub(BotResponseMultiplexer.DEFAULT_TOPIC, {
            async onResponse(_, content) {
                defaultTopic.push(content)
            }
        })
        await multiplexer.publish(`everything is a-ok
deep-state: birds are not real
deep-state: they are a government plot`)
        assert.deepStrictEqual(defaultTopic, ['everything is a-ok\n'])
        assert.deepStrictEqual(conspiracyTopic, [' birds are not real\n', ' they are a government plot'])
    })
})
