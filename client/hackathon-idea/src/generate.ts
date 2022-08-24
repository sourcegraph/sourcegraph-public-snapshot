import { createOrUpdateNotebook } from './createOrUpdateNotebook'
import { generatePackagesList } from './generatePackagesList'
import { Permutation } from './types'

import { readFileSync, writeFileSync } from 'fs'
import { createPermutations } from './createPermutations'

// Get the latest package list only if current list is empty
generatePackagesList()

let notebookMap: { [packageA: string]: { [packageB: string]: string } } = {}
try {
    notebookMap = JSON.parse(readFileSync('db/notebooks.json', 'utf8').toString() || '{}')
} catch {}

;(async function () {
    const permutations: Permutations = createPermutations()

    for (const [packageA, set] of permutations) {
        for (const packageB of set) {
            let notebookId: string | null = notebookMap[packageA]?.[packageB] ?? null

            notebookId = await createOrUpdateNotebook(notebookId, packageA, packageB)

            if (notebookMap[packageA] == null) {
                notebookMap[packageA] = {}
            }
            notebookMap[packageA][packageB] = notebookId

            writeFileSync('db/notebooks.json', JSON.stringify(notebookMap, null, 2))
        }
    }
})()
