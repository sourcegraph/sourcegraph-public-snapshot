/* eslint-disable @typescript-eslint/no-unsafe-call */
import path from 'path'

import express from 'express'

const app = express()
const port = 3888

const sourceboxRootPath = path.join(__dirname, '..')

app.use(express.static(__dirname))
app.use('/dist', express.static(path.join(sourceboxRootPath, 'dist')))

app.listen(port, () => {
    console.log(`Sourcebox sandbox started on port ${port}`)
})
