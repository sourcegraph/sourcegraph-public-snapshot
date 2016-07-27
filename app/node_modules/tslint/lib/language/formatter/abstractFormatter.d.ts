import { RuleFailure } from "../rule/rule";
import { IFormatter } from "./formatter";
export declare abstract class AbstractFormatter implements IFormatter {
    abstract format(failures: RuleFailure[]): string;
}
