/**
 * PageVars are defined and documented in cmd/frontend/internal/app/ui2/handlers.go pageVars struct.
 *
 * That is currently and should remain the canonical source of information for these types, please do not
 * document them here.
 */
export class PageVars {
	Rev: string;
	CommitID: string;

	constructor(vars: any) {
		if (!vars) {
			throw new TypeError("expected window.pageVars to exist, but it does not");
		}
		Object.assign(this, vars);
	}
}

export const pageVars = new PageVars(window["pageVars"]);
