import * as fs from 'mz/fs'
import { Redis } from 'ioredis'

/**
 * The type of the Redis client with additional script commands defined.
 */
export type ScriptedRedis = Redis & {
    // runs ./search-jobs.lua
    searchJobs: (args: (string | number | boolean)[]) => Promise<[string[][], number | null]>
}

/**
 * Registers the search-jobs.lua script in the given Redis instance. This function
 * returns the same redis client with additional methods attached.
 *
 * @param client The redis client.
 */
export async function defineRedisCommands(client: Redis): Promise<ScriptedRedis> {
    client.defineCommand('searchJobs', {
        numberOfKeys: 2,
        lua: (await fs.readFile(`${__dirname}/../../search-jobs.lua`)).toString(),
    })

    // The defineCommand method on the client dynamically defines a new method, but
    // the type system doesn't know that. We need to do a dumb cast here. This only
    // requires us to know the return type of the script.
    return client as ScriptedRedis
}
