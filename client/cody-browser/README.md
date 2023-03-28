<div align="center">
  <img src="public/cody.png" alt="logo"/>
  <h1> Cody - Chrome Extension</h1>
  <p>Cody: An AI-Powered Programming Assistant</p>
</div>

## Cody

Cody is a coding assistant that answers code questions and writes code for you by reading your entire codebase and the code graph.

Status: experimental ([request access](https://about.sourcegraph.com/cody))

## Features

- Select and right click on code snippet to ask Cody to:
  - Explain selected code
  - Optimized selected code
  - Debug selected code
- Start a new conversation with Cody in pop-up window

## Setup

1. Clone this repository.
2. Run `pnpm install` to install dependencies
3. Run `pnpm watch` to run developer service
4. Load Extension on Chrome
   1. Open - Chrome browser
   2. Access - `chrome://extensions`
   3. Check - Developer mode
   4. Find - Load unpacked extension
   5. Select - `dist` folder in this project (after `pnpm dev` or `pnpm build`)

## Build

To build in production, run `pnpm run build`.

The extension is built using [manifest v3](https://developer.chrome.com/docs/extensions/mv3/intro/), which is currently supported by Chrome and Edge but not by Firefox and Safari.

## Release

1. Update version nubmer
2. `pnpm run build`
3. Compress `dist` folder
4. Upload compressed `dist` folder to the Chrome Web Store [Developer Dashboard](https://chrome.google.com/webstore/devconsole)

## Tech Docs

- [Vite Plugin](https://vitejs.dev/guide/api-plugin.html)
- [Chrome Extension with manifest 3](https://developer.chrome.com/docs/extensions/mv3/)
- [Rollup](https://rollupjs.org/guide/en/)
- [Rollup-plugin-chrome-extension](https://www.extend-chrome.dev/rollup-plugin)
- [Tailwind CSS](https://tailwindcss.com/docs/configuration)
- [Chrome manifest v2 support](https://developer.chrome.com/docs/extensions/mv2/).
- [Firefox Manifest v3 support](https://discourse.mozilla.org/t/manifest-v3/94564).
- [Update Chrome Web Store item](https://developer.chrome.com/docs/webstore/update/)

## Credit

Created using Chrome Extension Boilerplate by [JohnBra/vite-web-extension](https://github.com/JohnBra/vite-web-extension)
