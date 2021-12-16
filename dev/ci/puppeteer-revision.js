#!/usr/bin/env node

const puppeteer = require('puppeteer/lib/cjs/puppeteer/revisions')

process.stdout.write(puppeteer.PUPPETEER_REVISIONS.chromium)
