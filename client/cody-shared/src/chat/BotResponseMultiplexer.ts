/**
 * Processes the part of a response from Cody addressed to a specific topic.
 */
export interface BotResponseSubscriber {
    /**
     * Processes incremental content from the bot. This may be called multiple times during a turn.
     * @param content the incremental text from the bot that was addressed to the subscriber
     */
    onResponse(content: string): Promise<void>

    /**
     * Notifies the subscriber that a turn has completed.
     */
    onTurnComplete(): Promise<void>
}

/**
 * A bot response subscriber that provides the entire bot response in one shot without
 * surfacing incremental updates.
 */
export class BufferedBotResponseSubscriber implements BotResponseSubscriber {
    private buffer_: string[] = []

    /**
     * Creates a BufferedBotResponseSubscriber. `callback` is called once per
     * turn with the bot's entire output provided in one shot. If the topic
     * was not mentioned, `callback` is called with `undefined` signifying the
     * end of a turn.
     * @param callback the callback to handle content from the bot, if any.
     */
    constructor(
        private callback: (content: string | undefined) => Promise<void>
    ) {}

    // BotResponseSubscriber implementation

    async onResponse(content: string) {
        this.buffer_.push(content)
    }

    async onTurnComplete(): Promise<void> {
        await this.callback(this.buffer_.length ? this.buffer_.join('') : undefined)
        this.buffer_ = []
    }
}

/**
 * Splits a string in one or two places.
 *
 * For example, `splitAt('banana!', 2) => ['ba', 'nana!']`
 * but `splitAt('banana!', 2, 4) => ['ba', 'na!']`
 * @param str the string to split.
 * @param startIndex the index to break the left substring from the rest.
 * @param endIndex the index to break the right substring from the rest, for
 * skipping the middle of the `str` from `[startIndex..endIndex)`.
 * @returns an array with the two substring pieces.
 */
function splitAt(str: string, startIndex: number, endIndex?: number): [string, string] {
    return [str.substring(0, startIndex), str.substring(typeof endIndex === 'undefined' ? startIndex : endIndex)]
}

/**
 * Incrementally consumes a response from the bot, breaking out parts addressing
 * different topics and forwarding those parts to a registered subscriber, if any.
 */
export class BotResponseMultiplexer {
    /**
     * The default topic. Messages without a prefix are sent to the default
     * topic subscriber, if any.
     */
    public static readonly DEFAULT_TOPIC = 'Assistant:'

    // Matches topics, or prefixes of topics
    private static readonly TOPIC_RE = /^([A-Za-z]*)(:?)/m

    private subs_ = new Map<string, BotResponseSubscriber>()

    // The topic currently being addressed by the bot
    private currentTopic_: string = BotResponseMultiplexer.DEFAULT_TOPIC

    // Buffers responses until topics can be parsed
    private buffer_: string = ''

    /**
     * Subscribes to a topic in the bot response. Each topic can have only one
     * subscriber at a time. New subscribers overwrite old ones.
     * @param topic the string prefix to subscribe to.
     * @param subscriber the handler for the content produced by the bot.
     */
    public sub(topic: string, subscriber: BotResponseSubscriber): void {
        // This test needs to be kept in sync with `TOPIC_RE`
        if (!/^[A-Za-z]+$/.test(topic)) {
            throw new Error(`topics must be A-Za-z, was "${topic}`)
        }
        this.subs_.set(topic, subscriber)
    }

    /**
     * Notifies all subscribers that the bot response is complete.
     */
    public async notifyTurnComplete(): Promise<void> {
        // Flush buffered content, if any
        if (this.buffer_) {
            this.push(this.currentTopic_, this.buffer_)
        }

        // Reset to the default topic, ready for another turn
        this.currentTopic_ = BotResponseMultiplexer.DEFAULT_TOPIC
        this.buffer_ = ''

        // Let subscribers react to the end of the turn.
        await Promise.all([...this.subs_.values()].map(subscriber => subscriber.onTurnComplete()))
    }

    /**
     * Parses part of a compound response from the bot and forwards as much as possible to
     * subscribers.
     * @param response the text of the next incremental response from the bot.
     */
    public async publish(response: string): Promise<void> {
        this.buffer_ += response
        while (this.buffer_) {
            // Look for something that could be a new topic.
            const match = this.buffer_.match(BotResponseMultiplexer.TOPIC_RE)
            if (match) {
                if (typeof match.index === 'undefined') {
                    throw new Error('unreachable')
                }
                const matchEnd = match.index + match[0].length
                const topic = match[1]
                const completeTopic = match[2] === ':'
                if (completeTopic) {
                    if (this.subs_.has(topic)) {
                        // Flush the buffered content before the new topic
                        let content
                        [content, this.buffer_] = splitAt(this.buffer_, match.index, matchEnd)
                        await this.push(this.currentTopic_, content)
                        // Switch topics
                        this.currentTopic_ = topic
                    } else {
                        // The topic has no subscriber, so treat it as content.
                        let content
                        [content, this.buffer_] = splitAt(this.buffer_, matchEnd)
                        await this.push(this.currentTopic_, content)
                    }
                } else {
                    // A topic is forming, but is incomplete, so wait for more content
                    return
                }
            } else {
                // No new topic is forming, so publish the in-progress content to the current topic
                await this.push(this.currentTopic_, this.buffer_)
                this.buffer_ = ''
            }
        }
    }

    // Publishes one specific topic to its subscriber, if any.
    private async push(topic: string, content: string): Promise<void> {
        const sub = this.subs_.get(topic)
        if (!sub) {
            return
        }
        return sub.onResponse(content)
    }

    /** Produces a prompt to describe the response format to the bot. */
    public prompt(): string {
        return `Separate each part of the response with a blank line. Prefix each part of the response with one of ${[...this.subs_.keys()].join(': ')}:\n\n`
    }
}
