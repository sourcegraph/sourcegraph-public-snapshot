const HttpsProxyAgent = require('https-proxy-agent')
const fetch = require('node-fetch')

const agent = new HttpsProxyAgent({
  protocol: 'http',
  path: '/Users/philipp/dev/domain-socket-proxy/unix.socket',
})

fetch('https://sourcegraph.com', { agent })
  .then(r => r.text())
  .then(console.log)
  .catch(console.error)
