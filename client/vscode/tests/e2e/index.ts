const { BrowserWindow, app } = require('electron')

module.exports.run = (testsRoot, callback) => {
    console.log({ testsRoot })
    console.log(callback)
    console.log('test')
    console.log(typeof BrowserWindow)
    console.log(typeof app)
}
