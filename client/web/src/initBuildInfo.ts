/**
 * Use runtime env variables to expose useful debugging information to the user.
 */
window.buildInfo = {
    commitSHA: process.env.COMMIT_SHA,
    version: process.env.VERSION,
}
