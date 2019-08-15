// import { gunzip, gzip } from "mz/zlib"

export async function decodeJSON(value: string): Promise<string> {
    return Promise.resolve(value)
    // return (await gunzip(new Buffer(value))).toString()
}

export async function encodeJSON(value: string): Promise<string> {
    return Promise.resolve(value)
    // return (await gzip(Buffer.from(value))).toString()
}
