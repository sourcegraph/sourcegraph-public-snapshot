import { AbstractFormatter } from "../language/formatter/abstractFormatter";
import { RuleFailure } from "../language/rule/rule";
export declare class Formatter extends AbstractFormatter {
    format(failures: RuleFailure[]): string;
}
