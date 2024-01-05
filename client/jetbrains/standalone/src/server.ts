import path from 'path'

import express from 'express'

const app = express()
const port = 3000

const javaResourcesPath = path.join(__dirname, '..', '..', 'src', 'main', 'resources')

app.use(express.static(__dirname))
app.use('/html', express.static(path.join(javaResourcesPath, 'html')))
app.use('/dist', express.static(path.join(javaResourcesPath, 'dist')))
app.use('/icons', express.static(path.join(javaResourcesPath, 'icons')))

app.listen(port, () => {
    console.log(`Standalone JetBrains extension started on port ${port}`)
})
