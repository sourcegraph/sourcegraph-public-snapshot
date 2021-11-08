export const BODY_JSON = '{"this is":"valid JSON","that should be":["re","indented"]}'
export const BODY_PLAIN = 'this is definitely not valid JSON\n\tand should not be reformatted in any way'

export const HEADERS_JSON = [
    {
        name: 'Content-Type',
        values: ['application/json; charset=utf-8'],
    },
    {
        name: 'Content-Length',
        values: [BODY_JSON.length.toString()],
    },
    {
        name: 'X-Complex-Header',
        values: ['value 1', 'value 2'],
    },
]

export const HEADERS_PLAIN = [
    {
        name: 'Content-Type',
        values: ['text/plain'],
    },
    {
        name: 'Content-Length',
        values: [BODY_PLAIN.length.toString()],
    },
    {
        name: 'X-Complex-Header',
        values: ['value 1', 'value 2'],
    },
]
