import { MessageHandler } from './rpc'

const server = new MessageHandler()

server.registerRequest('recipes/list', async data => {
    console.log(data)
    return []
})
server.registerNotification('recipes/execute', async data => {})

const client = new MessageHandler()

client.messageEncoder.pipe(server.messageDecoder)

client.messageEncoder.send({
    id: 0,
    method: 'recipes/list',
    params: 'bruh',
})

client.messageEncoder.send({
    id: 1,
    method: 'recipes/list',
    params: 'bruh222',
})
