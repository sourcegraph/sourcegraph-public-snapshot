import { getCombinedCommitPrompt, getCommitPrompt } from './ChangelogAnywhereList'

const mockInput = {
    input: {
        heading: 'updates customer information (#6625)',
        description: 'updated private with more examples of customer information',
        diff: 'content/company-info-and-process/policies/data-sharing.md content/company-info-and-process/policies/data-sharing.md\n@@ -53,3 +53,3 @@ Below you can find a matrix to help you make informed decisions about what data\n    </td>\n-   <td>Customer private source code\n+   <td>Customer private source code snippets (for support purposes)\n    </td>\n@@ -63,3 +63,3 @@ Below you can find a matrix to help you make informed decisions about what data\n    </td>\n-   <td>private repository names, legal contracts, company financials, incident reports for security issues \n+   <td>Customer roadmaps, customer number of codebases, customer challenges, private repository names, legal contracts, company financials, incident reports for security issues, private repository names, legal contracts, company financials, incident reports for security issues \n    </td>\n',
    },
}

const mockInput2 = {
    input: {
        heading: 'Update delivering-impact-reviews.md (#6630)',
        description: null,
        diff: 'content/departments/people-talent/people-ops/process/teammate-sentiment/impact-reviews/delivering-impact-reviews.md content/departments/people-talent/people-ops/process/teammate-sentiment/impact-reviews/delivering-impact-reviews.md\n@@ -4,3 +4,3 @@ \n \n-Impact reviews will be delivered synchronously in a 1:1 between the Manager and their direct report. Each Manager is responsible for scheduling a 30 - 60 minute (recommended) meeting with each Teammate to deliver their review packet, along with any corresponding promotion or compensation increases. All conversations must take place no later than **October 14 at the latest.**\n+Impact reviews will be delivered synchronously in a 1:1 between the Manager and their direct report. Each Manager is responsible for scheduling a 30 - 60 minute (recommended) meeting with each Teammate to deliver their review packet, along with any corresponding promotion or compensation increases. All conversations must take place no later than \\*_April 26, 2023 for H1 FY24 Impact Review Cycle_\n \n',
    },
}

describe('prompt is correct', () => {
    it('should return the correct prompt', () => {
        const prompt = getCommitPrompt(mockInput)
        expect(prompt).toMatchInlineSnapshot(`
            "
            Human:
            {\\"heading\\":\\"updates customer information (#6625)\\",\\"description\\":\\"updated private with more examples of customer information\\",\\"diff\\":\\"content/company-info-and-process/policies/data-sharing.md content/company-info-and-process/policies/data-sharing.md\\\\n@@ -53,3 +53,3 @@ Below you can find a matrix to help you make informed decisions about what data\\\\n    </td>\\\\n-   <td>Customer private source code\\\\n+   <td>Customer private source code snippets (for support purposes)\\\\n    </td>\\\\n@@ -63,3 +63,3 @@ Below you can find a matrix to help you make informed decisions about what data\\\\n    </td>\\\\n-   <td>private repository names, legal contracts, company financials, incident reports for security issues \\\\n+   <td>Customer roadmaps, customer number of codebases, customer challenges, private repository names, legal contracts, company financials, incident reports for security issues, private repository names, legal contracts, company financials, incident reports for security issues \\\\n    </td>\\\\n\\"}

            Generate a high-level summary of this change in a readable, plaintext, bullet-point list.

            Follow these rules strictly:
            Do: Use all the information available to build your summary.
            Don't: Mention details like specific files changed or commit hashes.
            Don't: Include anything that is not part of the bullet-point list.

            Assistant:
            -"
        `)
    })

    it('should return the correct combined prompt', () => {
        const prompt = getCombinedCommitPrompt([mockInput, mockInput2])
        expect(prompt).toMatchInlineSnapshot(`
            "
            Human:
            {\\"heading\\":\\"updates customer information (#6625)\\",\\"description\\":\\"updated private with more examples of customer information\\",\\"diff\\":\\"content/company-info-and-process/policies/data-sharing.md content/company-info-and-process/policies/data-sharing.md\\\\n@@ -53,3 +53,3 @@ Below you can find a matrix to help you make informed decisions about what data\\\\n    </td>\\\\n-   <td>Customer private source code\\\\n+   <td>Customer private source code snippets (for support purposes)\\\\n    </td>\\\\n@@ -63,3 +63,3 @@ Below you can find a matrix to help you make informed decisions about what data\\\\n    </td>\\\\n-   <td>private repository names, legal contracts, company financials, incident reports for security issues \\\\n+   <td>Customer roadmaps, customer number of codebases, customer challenges, private repository names, legal contracts, company financials, incident reports for security issues, private repository names, legal contracts, company financials, incident reports for security issues \\\\n    </td>\\\\n\\"}
            {\\"heading\\":\\"Update delivering-impact-reviews.md (#6630)\\",\\"description\\":null,\\"diff\\":\\"content/departments/people-talent/people-ops/process/teammate-sentiment/impact-reviews/delivering-impact-reviews.md content/departments/people-talent/people-ops/process/teammate-sentiment/impact-reviews/delivering-impact-reviews.md\\\\n@@ -4,3 +4,3 @@ \\\\n \\\\n-Impact reviews will be delivered synchronously in a 1:1 between the Manager and their direct report. Each Manager is responsible for scheduling a 30 - 60 minute (recommended) meeting with each Teammate to deliver their review packet, along with any corresponding promotion or compensation increases. All conversations must take place no later than **October 14 at the latest.**\\\\n+Impact reviews will be delivered synchronously in a 1:1 between the Manager and their direct report. Each Manager is responsible for scheduling a 30 - 60 minute (recommended) meeting with each Teammate to deliver their review packet, along with any corresponding promotion or compensation increases. All conversations must take place no later than \\\\\\\\*_April 26, 2023 for H1 FY24 Impact Review Cycle_\\\\n \\\\n\\"}

            Generate a high-level summary of these changes in a readable, plaintext, bullet-point list.

            Follow these rules strictly:
            Do: Use all the information available to build your summary.
            Don't: Mention details like specific files changed or commit hashes.
            Don't: Include anything that is not part of the bullet-point list.

            Assistant:
            "
        `)
    })
})
