module.exports = Date.now || now

function now() {
    return new Date().getTime()
}
