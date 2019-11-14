# Using HubSpot

This page describes how we use HubSpot to maintain high data quality. This allows us to increase the effectiveness of all Sourcegraph teams through accurate insights. 

## Associating Contacts to Deals

It is critical to associate all of the most important contacts within a company with any new deal. This should include the technical decision-maker, the economic decision-maker (if they are different) and the original member who introduced Sourcegraph to the organization. 

Result: this allows us to evaluate the effectiveness of marketing channels and sales touchpoints that our team has with an organization. How we reached the person(s) who introduced Sourcegraph to their organization is one of the most important factors in evaluating the success of marketing activities.

## Ongoing HubSpot maintenance

Account Executives are responsible for maintaining the deal data in HubSpot to ensure accurate data collection. This includes:

Updating deal fields regularly, especially:
* Deal stage
* Deal size
* Number of engineers

If a deal comes through a referral or introduction, tell BizOps so they can make the adjustment in the database to reflect this

When a deal is won:
1. Mark the ‘Deal Status’ as ‘Closed Won’
2. Mark the column ‘End of contract’ with the last day of the contract. HubSpot will automatically create a renewal deal based on this date

When a deal is lost:
1. Update the ‘Closed Lost Dropdown’ property to reflect the reason. If the reason doesn’t exist in the dropdown, you can talk to BizOps about adding one
2. Expand upon the reason in the longform ‘Closed Lost Reason’ field

Categorize any outbound emails into the ‘Manual Outbound Workflow’, which sets their ‘First Touchpoint’ as ‘Outbound’ and adds an event. To do this:
1. Select all contacts in the ‘Contacts’
2. Click the ‘More’ dropdown → ‘Enroll in Workflow’
3. Enroll in ‘Manual Outbound Workflow’

Maintaining [Server Installers to Company List](https://docs.google.com/spreadsheets/d/1Y2Z23-2uAjgIEITqmR_tC368OLLbuz12dKjEl4CMINA/edit?usp=sharing) and [Server to Company List](https://docs.google.com/spreadsheets/d/1wo_KQIcGrNGCWYKa6iHJ7MImJ_aI7GN12E-T21Es8TU/edit?usp=sharing) spreadsheets for every new company on a trial and new customers. 
