import * as path from 'path'
import { saveScreenshotsUponFailuresAndClosePage } from '../../../shared/src/e2e/screenshotReporter'
import { sourcegraphBaseUrl, createDriverForTest, Driver, gitHubToken } from '../../../shared/src/e2e/driver'

// 1 minute test timeout. This must be greater than the default Puppeteer
// command timeout of 30s in order to get the stack trace to point to the
// Puppeteer command that failed instead of a cryptic Jest test timeout
// location.
jest.setTimeout(1 * 60 * 1000)

process.on('unhandledRejection', error => {
    console.error('Caught unhandledRejection:', error)
})

process.on('rejectionHandled', error => {
    console.error('Caught rejectionHandled:', error)
})

describe('regression test suite', () => {
    let driver: Driver

    async function init(): Promise<void> {
        await driver.ensureLoggedIn()
    }

    // Start browser.
    beforeAll(
        async () => {
            driver = await createDriverForTest()
            await init()
        },
        // Cloning the repositories takes ~1 minute, so give initialization 2
        // minutes instead of 1 (which would be inherited from
        // `jest.setTimeout(1 * 60 * 1000)` above).
        2 * 60 * 1000
    )

    // Close browser.
    afterAll(async () => {
        if (driver) {
            await driver.close()
        }
    })

    // Take a screenshot when a test fails.
    saveScreenshotsUponFailuresAndClosePage(
        path.resolve(__dirname, '..', '..', '..'),
        path.resolve(__dirname, '..', '..', '..', 'puppeteer'),
        () => driver.page
    )

    describe('Search', () => {
        beforeAll(async () => {
            const repoSlugs = [
                'auth0/go-jwt-middleware',
                'kyoshidajp/ghkw',
                'PalmStoneGames/kube-cert-manager',
                'adjust/go-wrk',
                'P3GLEG/Whaler',
                'sajari/docconv',
                'marianogappa/chart',
                'divan/gobenchui',
                'tuna/tunasync',
                'mthbernardes/GTRS',
                'antonmedv/expr',
                'kshvakov/clickhouse',
                'xwb1989/sqlparser',
                'henrylee2cn/pholcus_lib',
                'itcloudy/ERP',
                'iovisor/kubectl-trace',
                'minio/highwayhash',
                'matryer/moq',
                'vkuznecovas/mouthful',
                'DirectXMan12/k8s-prometheus-adapter',
                'stephens2424/php',
                'ericchiang/k8s',
                'jonmorehouse/terraform-provisioner-ansible',
                'solo-io/supergloo',
                'intel-go/bytebuf',
                'xtaci/smux',
                'MatchbookLab/local-persist',
                'ossrs/go-oryx',
                'yep/eth-tweet',
                'deckarep/gosx-notifier',
                'zentures/sequence',
                'nishanths/license',
                'beego/mux',
                'status-im/status-go',
                'antonmedv/countdown',
                'lonng/nanoserver',
                'vbauerster/mpb',
                'evilsocket/sg1',
                'zhenghaoz/gorse',
                'nsf/godit',
                '3xxx/engineercms',
                'howtowhale/dvm',
                'gosuri/uitable',
                'github/vulcanizer',
                'metaparticle-io/package',
                'bwmarrin/snowflake',
                'wyh267/FalconEngine',
                'moul/sshportal',
                'fogleman/fauxgl',
                'DataDog/datadog-agent',
                'line/line-bot-sdk-go',
                'pinterest/bender',
                'esimov/diagram',
                'nytimes/openapi2proto',
                'iris-contrib/examples',
                'munnerz/kube-plex',
                'inbucket/inbucket',
                'golangci/awesome-go-linters',
                'htcat/htcat',
                'tidwall/pinhole',
                'gocraft/health',
                'ivpusic/grpool',
                'Antonito/gfile',
                'yinqiwen/gscan',
                'facebookarchive/httpcontrol',
                'josharian/impl',
                'salihciftci/liman',
                'kelseyhightower/konfd',
                'mohanson/daze',
                'google/ko',
                'freedomofdevelopers/fod',
            ]
            await driver.ensureHasExternalService({
                kind: 'github',
                displayName: 'GitHub (search regression test)',
                config: JSON.stringify({
                    url: 'https://github.com',
                    token: gitHubToken,
                    repos: repoSlugs,
                    repositoryQuery: ['none'],
                }),
                ensureRepos: [repoSlugs[repoSlugs.length - 1]],
            })
        })

        test(
            'Perform global text search for "alksdjflaksjdflkasjdf", expect 0 results.',
            async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/search?q=alksdjflaksjdflkasjdf')
                await driver.page.waitForSelector('.e2e-search-results')
                await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-results').length >= 1)
                await driver.page.evaluate(() => {
                    const resultsElem = document.querySelector('.e2e-search-results')
                    if (!resultsElem) {
                        throw new Error('No .e2e-search-results element found')
                    }
                    if (!(resultsElem as HTMLElement).innerText.includes('No results')) {
                        throw new Error('Expected "No results" message, but didn\'t find it')
                    }
                })
            },
            5 * 1000
        )
        test(
            'Perform global text search for "error type:", expect a few results.',
            async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/search?q=%22error+type:%22')
                await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length > 5)
            },
            5 * 1000
        )
        test(
            'Perform global text search for "error", expect many results.',
            async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/search?q=error')
                await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length > 10)
            },
            5 * 1000
        )
        test(
            'Perform global text search for "error", expect many results.',
            async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/search?q=error')
                await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length > 10)
            },
            5 * 1000
        )
        test(
            'Perform global text search for "error count:>1000", expect many results.',
            async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/search?q=error+count:1000')
                await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length > 10)
            },
            5 * 1000
        )
        test(
            'Perform global text search for "repohasfile:copying", expect many results.',
            async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/search?q=repohasfile:copying')
                await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length > 10)
            },
            5 * 1000
        )
    })
})
