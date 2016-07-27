import { RuleFailure } from "../rule/rule";
export interface IFormatter {
    format(failures: RuleFailure[]): string;
}
