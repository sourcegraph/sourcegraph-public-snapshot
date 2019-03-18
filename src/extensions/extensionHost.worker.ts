import { extensionHostWorkerMain } from 'sourcegraph/module/extension/workerMain'

// We're running in a Web Worker, so `self` is a DedicatedWorkerGlobalScope. For simplicity, our TypeScript config
// doesn't include lib.webworker.d.ts, so `self` here is actually not the correct type and requires this cast.
extensionHostWorkerMain(self as any)
