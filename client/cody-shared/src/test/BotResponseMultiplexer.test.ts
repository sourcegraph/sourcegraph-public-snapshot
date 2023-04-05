import assert from 'assert'

import { BotResponseMultiplexer, BufferedBotResponseSubscriber } from '../chat/BotResponseMultiplexer'

function promise<T>(): [(value: T) => void, Promise<T>] {
    let resolver
    const promise = new Promise<T>(resolve => (resolver = resolve))
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
            onResponse(content): Promise<void> {
                assert.strictEqual(content, 'hello, world')
                published(undefined)
                return Promise.resolve()
            },
            onTurnComplete() {
                return Promise.resolve()
            },
        })
        await multiplexer.publish('hello, world')
        await multiplexer.notifyTurnComplete()
        await publishedResult
    })

    it('discards messages when there is no subscriber', async () => {
        const multiplexer = new BotResponseMultiplexer()
        await multiplexer.publish('is this thing on?')
    })

    it('routes messages with an unknown prefix to the default subscriber', async () => {
        const multiplexer = new BotResponseMultiplexer()
        const [published, publishedResult] = promise<void>()
        multiplexer.sub(
            BotResponseMultiplexer.DEFAULT_TOPIC,
            new BufferedBotResponseSubscriber(content => {
                assert.strictEqual(content, "orator: I'm speechless")
                published(undefined)
                return Promise.resolve()
            })
        )
        await multiplexer.publish("orator: I'm speechless")
        await multiplexer.notifyTurnComplete()
        await publishedResult
    })

    it('routes messages to subscribers', async () => {
        const multiplexer = new BotResponseMultiplexer()
        multiplexer.sub(
            'cashier',
            new BufferedBotResponseSubscriber(content => {
                assert.deepStrictEqual(content, ' one double tall latte please\nand a donut\n')
                return Promise.resolve()
            })
        )
        multiplexer.sub(
            'barista',
            new BufferedBotResponseSubscriber(content => {
                assert.deepStrictEqual(content, ' can I get that to go?')
                return Promise.resolve()
            })
        )
        await multiplexer.publish(`cashier: one double tall latte please
and a donut
barista: can I get that to go?`)
    })

    it('can route to specific subscribers and the default subscriber', async () => {
        const multiplexer = new BotResponseMultiplexer()

        const conspiracyTopic: string[] = []
        multiplexer.sub('deep-state', {
            onResponse(content) {
                conspiracyTopic.push(content)
                return Promise.resolve()
            },
            onTurnComplete() {
                return Promise.resolve()
            },
        })

        const defaultTopic: string[] = []
        multiplexer.sub(BotResponseMultiplexer.DEFAULT_TOPIC, {
            onResponse(content) {
                defaultTopic.push(content)
                return Promise.resolve()
            },
            onTurnComplete() {
                return Promise.resolve()
            },
        })
        await multiplexer.publish(`everything is a-ok
deep-state: birds are not real
deep-state: they are a government plot`)
        await multiplexer.notifyTurnComplete()
        assert.deepStrictEqual(defaultTopic, ['everything is a-ok\n'])
        assert.deepStrictEqual(conspiracyTopic, [' birds are not real\n', ' they are a government plot'])
    })
})
