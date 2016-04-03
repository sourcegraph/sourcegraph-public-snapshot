declare function describe(label: string, body: (done: () => void) => void): void;
declare function it(label: string, body: (done: () => void) => void): void;
declare function before(fn: Function): void;
declare function after(fn: Function): void;
