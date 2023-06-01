import { registeredRecipes } from '@sourcegraph/cody-shared/src/chat/recipes/agent-recipes'

import { MessageHandler } from './rpc'

;(async () => {
    const server = new MessageHandler()

    server.registerRequest('recipes/list', async data => {
        return Object.values(registeredRecipes).map(({ id, title }) => ({
            id,
            title,
        }))
    })
    server.registerNotification('recipes/execute', async data => {})

    const client = new MessageHandler()

    client.messageEncoder.pipe(server.messageDecoder)
    server.messageEncoder.pipe(client.messageDecoder)

    console.log(await client.request('recipes/list', void {}))
})()
