import http from 'http'

// A generic server that you can extend to mock completion based APIs. This is
// mainly used as a stop-gap while we're waiting for proper access to some
// services.
//
//
// To run:
//
//   - `cd client/cody`
//   - `pnpm ts-node ./scripts/mock-server.ts`

const endpoints = {
    '/batch': {
        completions: [
            {
                completion: 'foo()',
                log_prob: -0.7,
            },
            {
                completion: 'bar()',
                log_prob: -0.9,
            },
            {
                completion: 'baz()',
                log_prob: -0.91,
            },
        ],
    },
}

http.createServer((req, res) => {
    let body = ''
    req.on('data', chunk => {
        body += chunk
    })
    req.on('end', () => {
        const payload = JSON.parse(body)

        console.log()
        console.log('>', req.url)
        console.log(payload)

        const mockedResponse = (endpoints as any)[req.url as any]
        if (mockedResponse) {
            console.log('<', mockedResponse)
            res.setHeader('content-type', 'application/json')
            res.write(JSON.stringify(mockedResponse))
            res.end()

            return
        }

        console.log('< 404 Not Found')
        res.statusCode = 404
        res.write('Not Found')
        res.end()
    })
}).listen(3001)

console.log('Listening on port 3001...')
