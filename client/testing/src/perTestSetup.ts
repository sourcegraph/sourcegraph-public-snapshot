import { TextEncoder, TextDecoder } from 'node:util'

// TextEncoder and TextDecoder are not available in a jsdom environment but
// required in various contexts.
if (!global.TextEncoder) {
    global.TextEncoder = TextEncoder
    // @ts-expect-error The interface of TextDecoder is not compatible with whatever TS thinks gobal.TextDecoder is.
    global.TextDecoder = TextDecoder
}
