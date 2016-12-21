declare module 'windows-mutex' {
    class Mutex {
        constructor(name: string);
        isActive(): boolean;
        release(): void;
    }
}
