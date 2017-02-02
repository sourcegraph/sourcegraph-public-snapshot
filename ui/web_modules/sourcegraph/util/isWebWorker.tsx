// isWebWorker is true if the current JavaScript context is in a Web
// Worker and false otherwise.
export default Boolean((global as any).importScripts); // tslint:disable-line no-default-export
