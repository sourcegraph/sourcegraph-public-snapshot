/*
 * Copyright 2016 Palantir Technologies, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
 * This TS script reads the metadata from each TSLint built-in rule
 * and serializes it in a format appropriate for the docs website.
 *
 * This script expects there to be a tslint-gh-pages directory
 * parallel to the main tslint directory. The tslint-gh-pages should
 * have the gh-pages branch of the TSLint repo checked out.
 * One easy way to do this is with the following Git command:
 *
 * ```
 * git worktree add -b gh-pages ../tslint-gh-pages origin/gh-pages
 * ```
 *
 * See http://palantir.github.io/tslint/develop/docs/ for more info
 *
 */

import * as fs from "fs";
import * as glob from "glob";
import * as yaml from "js-yaml";
import * as path from "path";

import {AbstractRule} from "../lib/language/rule/abstractRule";
import {IRuleMetadata} from "../lib/language/rule/rule";

const DOCS_DIR = "../docs";
const DOCS_RULE_DIR = path.join(DOCS_DIR, "rules");

const rulePaths = glob.sync("../lib/rules/*Rule.js");
const rulesJson: IRuleMetadata[] = [];
for (const rulePath of rulePaths) {
    // tslint:disable-next-line:no-var-requires
    const ruleModule = require(rulePath);
    const Rule = ruleModule.Rule as typeof AbstractRule;
    if (Rule != null && Rule.metadata != null) {
        const { metadata } = Rule;
        const fileData = generateRuleFile(metadata);
        const fileDirectory = path.join(DOCS_RULE_DIR, metadata.ruleName);

        // write file for each specific rule
        if (!fs.existsSync(fileDirectory)) {
            fs.mkdirSync(fileDirectory);
        }
        fs.writeFileSync(path.join(fileDirectory, "index.html"), fileData);

        rulesJson.push(metadata);
    }
}

// write overall data file, this is used to generate the index page for the rules
const fileData = JSON.stringify(rulesJson, undefined, 2);
fs.writeFileSync(path.join(DOCS_DIR, "_data", "rules.json"), fileData);

/**
 * Based off a rule's metadata, generates a string Jekyll "HTML" file
 * that only consists of a YAML front matter block.
 */
function generateRuleFile(metadata: IRuleMetadata) {
    const yamlData: any = {};
    // TODO: Use Object.assign when Node 0.12 support is dropped (#1181)
    for (const key of Object.keys(metadata)) {
        yamlData[key] = (<any> metadata)[key];
    }
    yamlData.optionsJSON = JSON.stringify(metadata.options, undefined, 2);
    yamlData.layout = "rule";
    yamlData.title = `Rule: ${metadata.ruleName}`;
    return `---\n${yaml.safeDump(yamlData, <any> {lineWidth: 140})}---`;
}
