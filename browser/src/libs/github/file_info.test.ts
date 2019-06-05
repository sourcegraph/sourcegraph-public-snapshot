import { startCase } from 'lodash'
import { readFile } from 'mz/fs'
import { getFilePath } from './util'

describe('github/file_info', () => {
    for (const version of ['github.com', 'ghe-2.14.11']) {
        describe(version, () => {
            for (const extension of ['vanilla', 'refined-github']) {
                describe(startCase(extension), () => {
                    it('finds the file path', async () => {
                        document.body.innerHTML = await readFile(
                            `${__dirname}/__fixtures__/${version}/blob/${extension}/page.html`,
                            'utf-8'
                        )
                        expect(getFilePath()).toMatchSnapshot()
                    })
                })
            }
        })
    }
})
