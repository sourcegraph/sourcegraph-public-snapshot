// import { expect } from 'chai'
// import { Chromeless } from 'chromeless'

// describe('repo navigation', () => {
//     let chrome: Chromeless<any>
//     beforeEach(() =>  {
//          chrome = new Chromeless()
//         // const screenshot = await chromeless
//         //   .goto('https://www.google.com')
//         //   .type('chromeless', 'input[name="q"]')
//         //   .press(13)
//         //   .wait('#resultStats')
//         //   .screenshot()

//         // console.log(screenshot) // prints local file path or S3 url

//         // await chromeless.end()
//     })

//     it.only('does something', async () => {
//         await chrome.goto('http://localhost:3080/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
//         await chrome.click(`div[data-tree-path="mux.go"]`)
//         await chrome.mousedown('.repository__viewer > div > table > tbody > tr:nth-child(21) > td.code > span:nth-child(3)')
//         await chrome.mousedown('.repository__viewer > div > table > tbody > tr:nth-child(21) > td.code.annotated > span:nth-child(8)')
//         await chrome.click('.repository__viewer > div > table > tbody > tr:nth-child(21) > td.code.annotated > span:nth-child(8)')
//         await chrome.click('.sg-tooltip > .tooltip__actions > a:nth-child(1)')
//         const href = await chrome.evaluate(() => window.location.href)
//         expect(href).to.eql('http://localhost:3080/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L43:6')

//         await chrome.mousedown('div.repository__viewer > div > table > tbody > tr:nth-child(43) > td.code > span:nth-child(3)')
//         await chrome.click('div.repository__viewer > div > table > tbody > tr:nth-child(43) > td.code > span:nth-child(3)')
//         await chrome.evaluate(() => {
//             const rect = document.querySelector('.blame')!.getBoundingClientRect()

//             function measureTextWidth(text: string, font: string): number {
//                 const tmp = document.createElement('canvas').getContext('2d')
//                 tmp!.font = font
//                 return tmp!.measureText(text).width
//             }

//             function click(x, y) {
//                 const ev = document.createEvent('MouseEvent')
//                 const el = document.elementFromPoint(x, y)
//                 ev.initMouseEvent(
//                     'click',
//                     true /* bubble */, true /* cancelable */,
//                     window, null,
//                     x, y, 0, 0, /* coordinates */
//                     false, false, false, false, /* modifier keys */
//                     0 /*left*/, null
//                 )
//                 el.dispatchEvent(ev)
//             }
//             const elAtPoint = document.elementFromPoint(rect.left + rect.width + 3, rect.top + 3)! as any
//             click(rect.left + rect.width + 3, rect.top + 3)
//             elAtPoint.click()
//         })
//         // console.log('goot a coord', JSON.parse(coord))

//         // await chrome.mousedown('.blame')
//         // await chrome.click('.blame')
//         await chrome.wait(3000)
//         const screenshot = await chrome.screenshot()
//         console.log(screenshot)
//     })

// })
