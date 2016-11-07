export type Action =
	Toast |
	ClearToast;

export class Toast {
	msg: string;
	timeout: number;

	constructor(msg: string, timeout?: number) {
		this.msg = msg;
		this.timeout = timeout || 2500;
	}
}

export class ClearToast {}
