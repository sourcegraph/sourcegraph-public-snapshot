import { Connection } from 'cxp'

/**
 * This simple CXP (Code Extension Protocol) extension is like a very simple editor extension. It decorates the
 * first line of each file with a blue background and a "Hello, world!" end-of-line message.
 *
 * Unlike traditional editor extensions built for a single editor, you can use CXP extensions in many editors--and
 * on websites like GitHub and Sourcegraph for code browsing, search, and review. A CXP extension is just a server
 * that speaks CXP to a client, so it can be written in any language and can run anywhere.
 *
 * Try it out!
 *
 * 1. Run `npx cxp-cli serve hello-world.ts` (on this file)
 * 2. TODO(sqs): add steps for enabling this extension in a CXP client
 * 3. Open a file and see the blue line color and the "Hello, world!" message!
 *
 * Next, try changing this extension's code to use a different color and message. After saving, reload in your
 * editor (or other CXP client) to use the updated extension.
 */
export function run(connection: Connection): void {
    // CXP extensions written in JavaScript/TypeScript export a `run` function, which is called each time a client
    // (such as an editor) connects.
    //
    // If you're just using this extension with your editor locally, there will only be a single connection (or one
    // per editor window, depending on the editor).

    // Tell the client that this extension provides decorations (so that our onTextDocumentDecoration handler above
    // is called whenever the client opens a file).
    connection.onInitialize(() => ({
        capabilities: { hoverProvider: true },
    }))

    // Register a handler for when the client asks us "what should I show in the hover tooltip for the current
    // cursor position?".
    connection.onHover(() => ({ contents: 'Hello, world!' }))
}
