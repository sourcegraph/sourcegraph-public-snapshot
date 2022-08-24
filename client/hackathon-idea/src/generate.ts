import { createOrUpdateNotebook } from './createOrUpdateNotebook'
import { Permutation } from './types'

import { readFileSync, writeFileSync } from 'fs'

let notebookMap: { [packageA: string]: { [packageB: string]: string } } = {}
try {
    notebookMap = JSON.parse(readFileSync('db/notebooks.json', 'utf8').toString() || '{}')
} catch {}

console.log('loading notebookMap', notebookMap)
;(async function () {
    const permutations: Permutation[] = [['react', 'redux']]

    for (let [packageA, packageB] of permutations) {
        let notebookId: string | null = notebookMap[packageA]?.[packageB] ?? null

        notebookId = await createOrUpdateNotebook(notebookId, 'react', 'redux')

        if (notebookMap[packageA] == null) {
            notebookMap[packageA] = {}
        }
        notebookMap[packageA][packageB] = notebookId
    }

    writeFileSync('db/notebooks.json', JSON.stringify(notebookMap, null, 2))
})()
