import { test, expect, type Page, type Browser } from '../../testing/integration'
import { Polly, type PollyServer } from '@pollyjs/core';
import { PlaywrightAdapter } from 'polly-adapter-playwright';
import RESTPersister from '@pollyjs/persister-rest';
import XHRAdapter from '@pollyjs/adapter-xhr';
import { chromium } from 'playwright-core';
import * as fs from 'fs';

test.describe('page.spec.ts', () => {

    let browser: Browser;
    let page: Page;
    let polly: Polly;
    let server: PollyServer;

    test.beforeEach(async () => {

        browser = await chromium.launch();
        page = await browser.newPage({
            baseURL: 'http://localhost:4173',
            recordVideo: {
                dir: './video'
            },
            recordHar: {
                path: './har'
            }
        });

        Polly.register(XHRAdapter);
        Polly.register(RESTPersister);
        Polly.register(PlaywrightAdapter);

        polly = new Polly('<Recording Name>', {
            adapters: ['playwright'],
            adapterOptions: {
                playwright: {
                    context: page
                }
            },
            persister: 'rest'
        });

        server = polly.server;

        // server.any('*').passthrough();

        server.options(new URL('/.api/graphql', 'http://localhost:4173').href).intercept((request, response) => {
            response
                .setHeader('Access-Control-Allow-Origin', '*')
                .setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization')
                .send(200)
        })
        // server.post(new URL('/.api/graphql', 'http://localhost:4173').href).passthrough();
        server.post(new URL('/.api/graphql', 'http://localhost:4173').href).intercept((req, res) => {
            res.json({
                'data': {},
                'errors': {}
            });
        })

        // todo: change this to just server assets, and introduce the indexHtml handler
        server.get(new URL('/*path', 'http://localhost:4173').href).intercept((req, res) => {
            res.setHeader('Cache-Control', 'public, max-age=31536000, immutable')
            // eslint-disable-next-line no-sync
            const content = fs.readFileSync(`./build/${req.params.path}`).toString('utf-8');
            res.status(200).send(content);
        });

        // server.get(new URL('/search/*path', 'http://localhost:4173').href).passthrough();

        // Let browser handle data: URIs
        server.get('data:*rest').passthrough();

        // Special URL: The browser redirects to chrome-extension://invalid
        // when requesting an extension resource that does not exist.
        server.get('chrome-extension://invalid/').passthrough();
    });

    test.afterEach(async () => {
        await polly.stop();
        await page.close();
        await browser.close();
    });

    test.skip('shows suggestions', async ({ sg,  }) => {
        await page.goto('/search')
        const searchInput = page.getByRole('textbox')
        await searchInput.click()

        // Default suggestions
        await expect(page.getByLabel('Narrow your search')).toBeVisible()

        sg.mockTypes({
            SearchResults: () => ({
                repositories: [{ name: 'github.com/sourcegraph/sourcegraph' }],
                results: [
                    {
                        __typename: 'FileMatch',
                        file: {
                            path: 'sourcegraph.md',
                            url: '',
                        },
                    },
                ],
            }),
        })

        // Repo suggestions
        await searchInput.fill('source')
        await expect(page.getByLabel('Repositories')).toBeVisible()
        await expect(page.getByLabel('Files')).toBeVisible()

        // Fills suggestion
        await page.getByText('github.com/sourcegraph/sourcegraph').click()
        await expect(searchInput).toHaveText('repo:^github\\.com/sourcegraph/sourcegraph$ ')
    })

    test.only('submits search on enter', async () => {
        await page.goto('/search')
        const searchInput = page.getByRole('textbox')
        await searchInput.fill('source')

        // Submit search
        await searchInput.press('Enter')
        await expect(page).toHaveURL(/\/search\?q=.+$/)
    })

    test('fills search query from URL', async () => {
        await page.goto('/search?q=test')
        await expect(page.getByRole('textbox')).toHaveText('test')
    })

})
