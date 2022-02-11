import * as fs from 'fs'
import * as path from 'path'

export function generateSourcegraphCodeTable(lines: string[]): string {
    const code = lines
        .map(
            (line, index) => `<tr>
                <td class="line" data-line="${index + 1}"></td>
                <td class="code">${line}</td>
            </tr>`
        )
        .join('\n')

    // eslint-disable-next-line no-sync
    const styles = fs.readFileSync(path.join(__dirname, 'styles.css'), 'utf-8')

    return `<div class="sourcegraph-testcase">
        <style>
            ${styles}
        </style>
        <div class="container">
            <div class="left"></div>
            <div class="blob-container">
                <div class="blob">
                    <code class="code">
                        <table><tbody>${code}</tbody></table>
                    </code>
                </div>
            </div>
        </div>
    </div>`
}
