import * as assert from 'assert'
import * as path from 'path'

import * as fs from 'mz/fs'

import { languageSpecs } from './languages'

describe('all defined languages', () => {
    it('should have an icon file', async () => {
        const languageIDs = (
            await Promise.all(
                languageSpecs.map(async ({ languageID }) => {
                    const filePath = path.join(__dirname, '..', '..', '..', 'icons', `${languageID}.png`)

                    return {
                        languageID,
                        filePath,
                        exists: await fs.exists(filePath),
                    }
                })
            )
        )
            .filter(({ exists }) => !exists)
            .map(({ languageID }) => languageID)

        if (languageIDs.length > 0) {
            assert.fail(
                `Each LANG must have a corresponding icons/LANG.png file. Icons missing for ${languageIDs.join(', ')}.`
            )
        }
    })
})
