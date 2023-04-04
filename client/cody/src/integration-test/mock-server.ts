import express from 'express'

export const SERVER_PORT = 49300

// Runs a stub Cody service for testing.
export async function run<T>(around: () => Promise<T>): Promise<T> {
    const app = express()
    app.use(express.json())

    app.post('/.api/completions/stream', (req, res) => {
        res.send('event: completion\ndata: {"completion": "hello, world"}\n\nevent: done\ndata: {}\n\n')
    })

    const server = app.listen(SERVER_PORT, () => {
        console.log(`Mock server listening on port ${SERVER_PORT}`)
    })

    const result = await around()
    server.close()

    return result
}
