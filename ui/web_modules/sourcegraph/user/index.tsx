import {User} from "sourcegraph/api";

// inBeta tells if the given user is a part of the given beta program.
export function inBeta(u: User | null, b: string): boolean {
	if (!u || !u.Betas) { return false; }
	return u.Betas.indexOf(b) !== -1;
}

export interface ExternalToken {
	uid: number;
	host: string;
	scope: string;
};
