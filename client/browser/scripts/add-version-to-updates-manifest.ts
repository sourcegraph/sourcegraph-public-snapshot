import * as fs from 'fs'
import * as path from 'path'

const extManifestPath = path.resolve(__dirname, '..', 'build/firefox/manifest.json')
const updatesManifestPath = path.resolve(__dirname, '..', 'src/extension/updates.manifest.json')

const updateLink = process.argv[2]

interface Update {
    version: string
    update_link: string
}

function addVersionToManifest(): void {
    if (!updateLink || updateLink === '') {
        console.log('Usage: add-version-to-updates-manifest.ts <update-link>')
        process.exit(1)
    }

    const extManifest = JSON.parse(fs.readFileSync(extManifestPath, 'utf8'))
    const updatesManifest = JSON.parse(fs.readFileSync(updatesManifestPath, 'utf8'))

    const version = extManifest.version as string
    ;(updatesManifest.addons['sourcegraph-for-firefox@sourcegraph.com'].updates as Update[]).push({
        version,
        update_link: updateLink,
    })

    fs.writeFileSync(updatesManifestPath, JSON.stringify(updatesManifest, null, 4), 'utf8')
}

addVersionToManifest()
