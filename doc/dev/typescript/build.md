# TypeScript build documentation

We use TypeScript for two products:

- The main Sourcegraph web application
  - 2 different entrypoints: [OSS `main.tsx`](../../../web/src/main.tsx) and [Enterprise `main.tsx`](../../../web/src/enterprise/main.tsx)
- The Sourcegraph browser extension

These both use shared TypeScript code in [`../shared`](../shared). Each product has its own separate Webpack configuration.

We want the simplest possible build process for these products. Specifically:

- 

