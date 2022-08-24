import fetch from 'node-fetch'
import { readFileSync, writeFileSync } from 'fs'

// @ts-ignore
process.env['NODE_TLS_REJECT_UNAUTHORIZED'] = 0

// generate packages list only if current list is empty
export async function generatePackagesList(): Promise<void> {
    try {
        const currentPackagesList = JSON.parse(readFileSync('src/packages.json', 'utf8').toString() || '{}')
        const apiKey = ''
        if (!currentPackagesList && apiKey) {
            const packageList: { name: any; description: any }[] = []
            for (let i = 1; i <= 10; i++) {
                const parameters = {
                    api_key: apiKey,
                    per_page: 100,
                    page: i,
                    platforms: 'NPM',
                    sort: 'rank',
                }
                const uri = new URL('/search', 'https://libraries.io/api')
                // @ts-ignore
                const parametersString = new URLSearchParams({ ...parameters }).toString()
                uri.search = parametersString
                // FETCH
                const response = await fetch(uri.href)
                const results: any = await response.json()
                // @ts-ignore
                results.forEach(result => {
                    const packageInfo = {
                        name: result.name,
                        description: result.description,
                        kStars: result.stars,
                    }
                    packageList.push(packageInfo)
                })
                if (!response) return
            }
            if (packageList.length === 1000) {
                writeFileSync('./src/package.jsons', JSON.stringify(packageList))
            }
        }
    } catch (error) {
        console.error(error)
    }
}
