import { ChatViewProvider } from './chat/ChatViewProvider'
import { ChatMessage } from './chat/transcript/messages'

// A one-slot channel which lets readers block on a value being
// available from a writer. Tests use this to wait for the
// extension to produce a value.
class Rendezvous<T> {
    private resolve: (value: T) => void
    private promise: Promise<T>

    constructor() {
        this.resolve = () => {}
        this.promise = new Promise(resolve => {
            this.resolve = resolve
        })
    }

    public set(value: T): void {
        this.resolve(value)
        // FIXME: The extension constructs *two* ChatViewProviders.
        // Tests need to hang onto the second one, so we reset the
        // Promise here.
        // console.log('setting rendezvous value', new Error().stack)
        this.promise = Promise.resolve(value)
    }

    public get(): Promise<T> {
        return this.promise
    }
}

// The interface to test hooks for the extension. If
// TestSupport.instance is set, the extension is running in an
// integration test.
export class TestSupport {
    public static instance: TestSupport | undefined

    public chatViewProvider = new Rendezvous<ChatViewProvider>()

    public async chatTranscript(): Promise<ChatMessage[]> {
        return (await this.chatViewProvider.get()).transcriptForTesting(this)
    }
}
