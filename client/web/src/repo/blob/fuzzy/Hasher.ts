export class Hasher {
  private h: number = 0;
  constructor() {}
  public update(ch: string): Hasher {
    for (let i = 0; i < ch.length; i++) {
      this.h = (Math.imul(31, this.h) + ch.charCodeAt(i)) | 0;
    }
    return this;
  }
  public digest(): number {
    return this.h;
  }
  public reset() {
    this.h = 0;
  }
}
