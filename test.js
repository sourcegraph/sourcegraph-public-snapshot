const HttpsProxyAgent = require('https-proxy-agent')
const fetch = require('node-fetch')

const agent = new HttpsProxyAgent({
  protocol: 'http',
  host: '192.168.0.220',
  port: '9090',
})

fetch('https://sourcegraph.com', { agent })
  .then(r => r.text())
  .then(console.log)
  .catch(console.error)
