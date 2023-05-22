import assert from 'assert'

import { BotResponseMultiplexer, BufferedBotResponseSubscriber } from './bot-response-multiplexer'

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

    it('treats unknown tags as content', async () => {
        const multiplexer = new BotResponseMultiplexer()
        const [published, publishedResult] = promise<void>()
        multiplexer.sub(
            BotResponseMultiplexer.DEFAULT_TOPIC,
            new BufferedBotResponseSubscriber(content => {
                assert.strictEqual(content, "<orator>I'm speechless</orator>")
                published(undefined)
                return Promise.resolve()
            })
        )
        await multiplexer.publish("<orator>I'm speechless</orator>")
        await multiplexer.notifyTurnComplete()
        await publishedResult
    })

    it('things which lookl like tags as content', async () => {
        const multiplexer = new BotResponseMultiplexer()
        const [published, publishedResult] = promise<void>()
        multiplexer.sub(
            BotResponseMultiplexer.DEFAULT_TOPIC,
            new BufferedBotResponseSubscriber(content => {
                assert.strictEqual(content, '[] <--insert coin </wow> party hat emoji O:>')
                published(undefined)
                return Promise.resolve()
            })
        )
        await multiplexer.publish('[] <--insert coin </wow> party hat emoji O:>')
        await multiplexer.notifyTurnComplete()
        await publishedResult
    })

    it('routes messages to subscribers', async () => {
        const multiplexer = new BotResponseMultiplexer()
        multiplexer.sub(
            'cashier',
            new BufferedBotResponseSubscriber(content => {
                assert.deepStrictEqual(content, 'one double tall latte please\nand a donut\n')
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
        await multiplexer.publish(`<cashier>one double tall latte please
and a donut
</cashier><barista> can I get that to go?</barista>`)
    })

    it('can route nested topics', async () => {
        const multiplexer = new BotResponseMultiplexer()

        const conspiracyTopic: string[] = []
        multiplexer.sub('conspiracy', {
            onResponse(content) {
                conspiracyTopic.push(content)
                return Promise.resolve()
            },
            onTurnComplete() {
                return Promise.resolve()
            },
        })

        const deeperConspiracyTopic: string[] = []
        multiplexer.sub('deeper-conspiracy', {
            onResponse(content) {
                deeperConspiracyTopic.push(content)
                return Promise.resolve()
            },
            onTurnComplete() {
                return Promise.resolve()
            },
        })

        // Note, no default topic this time.
        await multiplexer.publish(`everything is a-ok
<conspiracy>birds are not <deeper-conspiracy><--they are a government plot!!1!--></deeper-conspiracy>real</conspiracy>`)
        await multiplexer.notifyTurnComplete()
        assert.deepStrictEqual(conspiracyTopic, ['birds are not ', 'real'])
        assert.deepStrictEqual(deeperConspiracyTopic, ['<--they', ' are a government plot!!1!-->'])
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
<deep-state>birds are not real</deep-state>
<deep-state> they are a government plot</deep-state>`)
        await multiplexer.notifyTurnComplete()
        assert.deepStrictEqual(defaultTopic, ['everything is a-ok\n', '\n'])
        assert.deepStrictEqual(conspiracyTopic, ['birds are not real', ' they are a government plot'])
    })

    it('can handle sloppily closed tags, or unclosed tags', async () => {
        const multiplexer = new BotResponseMultiplexer()

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

        const rowTopic: string[] = []
        multiplexer.sub('row', {
            onResponse(content) {
                rowTopic.push(content)
                return Promise.resolve()
            },
            onTurnComplete() {
                return Promise.resolve()
            },
        })

        const cellTopic: string[] = []
        multiplexer.sub('cell', {
            onResponse(content) {
                cellTopic.push(content)
                return Promise.resolve()
            },
            onTurnComplete() {
                return Promise.resolve()
            },
        })

        await multiplexer.publish('<row>S, V F X<cell>variety</cell><cell>limburger</row>F U N E X')
        await multiplexer.notifyTurnComplete()
        assert.deepStrictEqual(defaultTopic, ['F U N E X'])
        assert.deepStrictEqual(rowTopic, ['S, V F X'])
        assert.deepStrictEqual(cellTopic, ['variety', 'limburger'])
    })
})
