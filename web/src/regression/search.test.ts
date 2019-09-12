import * as path from 'path'
import { saveScreenshotsUponFailuresAndClosePage } from '../../../shared/src/e2e/screenshotReporter'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { regressionTestInit, createAndInitializeDriver } from './util/init'
import * as GQL from '../../../shared/src/graphql/schema'
import { GraphQLClient } from './util/GraphQLClient'
import { ensureExternalService, waitForRepos } from './util/api'

const testRepoSlugs = [
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
    'sgtest/mux',
    'facebook/react',
]

describe('Search regression test suite', () => {
    regressionTestInit()
    const config = getConfig(['sudoToken', 'username', 'gitHubToken', 'sourcegraphBaseUrl'])
    let driver: Driver
    let gqlClient: GraphQLClient

    // Take a screenshot when a test fails.
    saveScreenshotsUponFailuresAndClosePage(
        path.resolve(__dirname, '..', '..', '..'),
        path.resolve(__dirname, '..', '..', '..', 'puppeteer'),
        () => driver.page
    )

    if (new URL(window.location.href).origin !== new URL(config.sourcegraphBaseUrl).origin) {
        throw new Error(
            `JSDOM URL "${window.location.href}" did not match Sourcegraph base URL "${config.sourcegraphBaseUrl}". Tests will fail with a same-origin violation. Try setting the environment variable SOURCEGRAPH_BASE_URL.`
        )
    }

    describe('Search over a dozen repositories', () => {
        beforeAll(
            async () => {
                driver = await createAndInitializeDriver()
                gqlClient = new GraphQLClient(config.sourcegraphBaseUrl, config.sudoToken, config.username)
                await ensureExternalService(gqlClient, {
                    kind: GQL.ExternalServiceKind.GITHUB,
                    uniqueDisplayName: 'GitHub (search-regression-test)',
                    config: {
                        url: 'https://github.com',
                        token: config.gitHubToken,
                        repos: testRepoSlugs,
                        repositoryQuery: ['none'],
                    },
                })
                await waitForRepos(gqlClient, ['github.com/' + testRepoSlugs[testRepoSlugs.length - 1]])
            },
            // Cloning the repositories takes ~1 minute, so give initialization 2 minutes
            2 * 60 * 1000
        )

        // Close browser.
        afterAll(async () => driver && (await driver.close()))

        test('Perform global text search (alksdjflaksjdflkasjdf) with 0 results.', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=alksdjflaksjdflkasjdf')
            await driver.page.waitForSelector('.e2e-search-results')
            await driver.page.waitForFunction(() => {
                const resultsElem = document.querySelector('.e2e-search-results')
                if (!resultsElem) {
                    return false
                }
                const resultsText = (resultsElem as HTMLElement).innerText
                if (!resultsText) {
                    return false
                }
                if (!resultsText.includes('No results')) {
                    throw new Error('Expected "No results" message, but didn\'t find it')
                }
                return true
            })
        })
        test('Global text search ("error type:") with a few results.', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q="error+type:%5Cn"')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length >= 3)
        })
        test('Global text search excluding repository ("error type:") with a few results.', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q="error+type:%5Cn"+-repo:google')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length > 0)
            await driver.page.waitForFunction(() => {
                const results = Array.from(document.querySelectorAll('.e2e-search-result'))
                if (results.length === 0) {
                    return false
                }
                const hasExcludedRepo = results.some(el => el.textContent && el.textContent.includes('google'))
                if (hasExcludedRepo) {
                    throw new Error('Results contain excluded repository')
                }
                return true
            })
        })
        test('Global text search (error) with many results.', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=error')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length > 10)
        })
        test('Global text search (error count:>1000), expect many results.', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=error+count:1000')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length > 10)
        })
        test('Global text search (repohasfile:copying), expect many results.', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=repohasfile:copying')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length >= 2)
        })
        test('Global text search for something with more than 1000 results and use "count:1000".', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=.+count:1000')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length > 10)
        })
        test('Global text search for a regular expression without indexing: (index:no ^func.*$), expect many results.', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=index:no+^func.*$')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length > 10)
        })
        test('Search for a repository by name.', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=repo:auth0/go-jwt-middleware$')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length === 1)
        })
        test('Single repository, case-sensitive search.', async () => {
            await driver.page.goto(
                config.sourcegraphBaseUrl + '/search?q=repo:%5Egithub.com/adjust/go-wrk%24+String+case:yes'
            )
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length === 2)
        })
        test('Global text search, fork:only, few results', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=fork:only+router')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length >= 5)
        })
        test('Global text search, fork:only, 1 result', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=fork:only+FORK_SENTINEL')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length === 1)
        })
        test('Global text search, fork:no, 0 results', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=fork:only+FORK_SENTINEL')
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length === 0)
        })
        test('Text search non-master branch, large repository, many results', async () => {
            // The string `var ExecutionEnvironment = require('ExecutionEnvironment');` occurs 10 times on this old branch, but 0 times in current master.
            await driver.page.goto(
                config.sourcegraphBaseUrl +
                    '/search?q=repo:%5Egithub%5C.com/facebook/react%24%400.3-stable+"var+ExecutionEnvironment+%3D+require%28%27ExecutionEnvironment%27%29%3B"'
            )
            await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-search-result').length === 10)
        })
    })
})
