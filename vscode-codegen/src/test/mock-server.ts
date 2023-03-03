import * as http from 'http'
import * as ws from 'ws'

export const SERVER_PORT = 49300
export const EMBEDDING_PORT = 49301

// Runs a stub Cody service for testing.
export async function run(callback: () => Promise<any>) {
	// TODO: Extend these servers to support expectations.
	const socketServer = new ws.WebSocketServer({
		port: SERVER_PORT,
	})
	await new Promise(resolve => {
		socketServer.on('connection', socket => {
			socket.on('message', message => {
				const req = JSON.parse(message.toString())
				socket.send(`{"requestId": ${req.requestId}, "kind": "response:complete", "message": "hello, world"}`)
			})
		})
		socketServer.on('listening', resolve)
	})

	const embeddingServer = http.createServer()
	embeddingServer.on('request', (request, response) => {
		response.statusCode = 200
		response.setHeader('content-type', 'text/plain')
		response.write('hello, world')
		response.end()
	})
	embeddingServer.listen(EMBEDDING_PORT)

	await callback()

	socketServer.close()
	embeddingServer.close()
}
