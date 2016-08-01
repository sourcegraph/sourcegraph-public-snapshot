import { AbstractFormatter } from "../language/formatter/abstractFormatter";
import { RuleFailure } from "../language/rule/rule";
export declare class Formatter extends AbstractFormatter {
    format(failures: RuleFailure[]): string;
    private pad(str, len);
    private getPositionMaxSize(failures);
    private getRuleMaxSize(failures);
}
