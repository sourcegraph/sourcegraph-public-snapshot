package main

import (
	"context"
	"flag"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
)

type termResult struct {
	definition    string
	referenceURLs []string
}

var (
	wutFlagSet = flag.NewFlagSet("sg teammate", flag.ExitOnError)
	wutCommand = &ffcli.Command{
		Name:       "wut",
		ShortUsage: "sg wut [term]",
		ShortHelp:  "Look up a term or abbreviation overheard at Sourcegraph.",
		LongHelp: `"I know of NGTTHTT introduction of language crafted for OAL to m14g."
("I know of [no greater trajedy to humanity than the] introduction of language
crafted for [optimization and leading] to [misunderstanding].")
                                                          - FILO coder @slimsag

Every industry and every company eventually develops its own esoteric,
confusing, and poorly-documented jargon that it loves to throw in the faces of
unsuspecting newcomers and industry veterans alike. While we try to discourage
an over-reliance on acronyms and lingo here at Sourcegraph, finding shortcuts
is something human brains are best at, for better or worse. (And we're all
humans here.)

In part to poke fun at us as humans, and in part because it's actually useful
to have a dictionary like this (we all know it can feel embarrassing or
disruptive to have to ask what something means), we've created a tool to help
de-obfuscate the abbreviations and jargon we use!

Definitions and concept shamelessly stolen from @slimsag's original document:
https://docs.google.com/spreadsheets/d/1E71lKEMl6sFvey0BAVWyQih2an64bSutZfBJ6haXY9c/edit#gid=0

Search term coming up blank? Please consider expanding our dictionary! (Or ask
your team to stop using so much jargon!)`,
		FlagSet: wutFlagSet,
		Exec:    wutExec,
	}
	// In no particular order. Make sure all keys are lowercase!
	dictionary = map[string]termResult{
		"gql":            {definition: "GraphQL (a type of API, sort of like HTTP or Rest APIs)", referenceURLs: []string{"https://graphql.org/learn/", "https://k8s.sgdev.org/api/console"}},
		"vsce":           {definition: "VS Code Extension, referring to either the Sourcegraph integration extension for the VS Code IDE, or the Microsoft tool called 'vsce' used to publish extensions", referenceURLs: []string{"https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph"}},
		"bext":           {definition: "Brower extension, sometimes pronounced literally as \"bext\" or \"baxt\"", referenceURLs: []string{}},
		"lsif":           {definition: "Language Server index Format, a specification created by Microsoft which Sourcegraph uses to provide code intelligence.", referenceURLs: []string{}},
		"rfh":            {definition: "Request for help", referenceURLs: []string{}},
		"rfc":            {definition: "Request for comments", referenceURLs: []string{}},
		"rce":            {definition: "Remote code execution (a security vulnerability)", referenceURLs: []string{}},
		"ssrf":           {definition: "Server side request forgery (a security vulnerability)", referenceURLs: []string{}},
		"seo":            {definition: "Search engine optimization, making Google understand our web pages better", referenceURLs: []string{}},
		"ssbc":           {definition: "Server-side batch changes. Large scale code refactoring that runs as part of the Sourcegraph server, rather than on a developer's laptop.", referenceURLs: []string{}},
		"poc":            {definition: "Proof of concept", referenceURLs: []string{}},
		"pr":             {definition: "Pull request, where code is sent to be reviewed before becoming a part of the product", referenceURLs: []string{}},
		"todo":           {definition: "A note left in the code as a comment indicating something we should do", referenceURLs: []string{}},
		"dfs":            {definition: "Damn Fine Source code", referenceURLs: []string{}},
		"easy stamp":     {definition: "Change that needs approval but not review", referenceURLs: []string{}},
		"stamp":          {definition: "Change that needs approval but not review", referenceURLs: []string{}},
		"ci":             {definition: "Continuous Integration, a server that runs our tests and ensures things are not broken. Often stated as \"CI is failing\" and \"CI is slow\"", referenceURLs: []string{}},
		"dogfood":        {definition: "Either k8s.sgdev.org (the \"dogfood\" instance) or just saying \"we should try what we built\" in general", referenceURLs: []string{}},
		"dev":            {definition: "Someone who can barely write code, but does so professionally", referenceURLs: []string{}},
		"k8s":            {definition: "Kubernetes, a thing for deploying software across multiple computers. The \"8\" is because there are 8 characters between the letters K and S: k[ubernete]s", referenceURLs: []string{}},
		"a11y":           {definition: "Accessibility, the \"11\" is because there are 11 characters between the letters A and Y: a[ccessibilit]y", referenceURLs: []string{}},
		"i18n":           {definition: "Internationalization, like having the UI show in multiple languages. The \"18\" is because there are 18 characters between the letters I and N: i[nternationalizatio]n", referenceURLs: []string{}},
		"gcp":            {definition: "Servers hosted on Google's cloud", referenceURLs: []string{}},
		"aws":            {definition: "Servers hosted on Amazon's cloud", referenceURLs: []string{}},
		"smtp":           {definition: "An email server protocol", referenceURLs: []string{}},
		"imap":           {definition: "An email server protocol", referenceURLs: []string{}},
		"ide":            {definition: "Integrated development environment, the text editor people use to write code", referenceURLs: []string{}},
		"hg":             {definition: "Horse Graph", referenceURLs: []string{}},
		"standup":        {definition: "Everyone sits down for 15/30/60m and says what they are doing. Sometimes they just type it.", referenceURLs: []string{}},
		"pgsql":          {definition: "Postgres database", referenceURLs: []string{}},
		"psql":           {definition: "Postgres database", referenceURLs: []string{}},
		"mvp":            {definition: "Minimum viable product, the bare minimum needed to see a feature working for example. Think \"very early stages, experimental\"", referenceURLs: []string{}},
		"mvc":            {definition: "Model-View-Controller frontend/JavaScript pattern. React. Google \"MVC\"", referenceURLs: []string{}},
		"tdd":            {definition: "Test driven development, you write the tests before you write the code that would pass the tests.", referenceURLs: []string{}},
		"api":            {definition: "Application programming interface; like when your browser makes a request to your bank's web server to send money", referenceURLs: []string{}},
		"dom":            {definition: "Document Object Model, a tree of buttons/text/etc that are displayed in browsers. \"The DOM\" refers to all the stuff making up the web page.", referenceURLs: []string{}},
		"loc":            {definition: "Lines of code", referenceURLs: []string{}},
		"loe":            {definition: "Level of effort", referenceURLs: []string{}},
		"cac":            {definition: "customer acquisition cost (how much $ spent to win a new customer)", referenceURLs: []string{"https://docs.google.com/document/d/1_fN2koxGL94q7M9NCfrzi9jPkvpc8VVt1JSV0_vklys/edit"}},
		"icp":            {definition: "ideal customer profile (the hypothetical perfect customer for your organization; lists all of the qualities that make them the best fit for the solutions you provide.)", referenceURLs: []string{}},
		"lcv":            {definition: "lifetime customer value (the total revenue a customer will generate for a business throughout the relationship)", referenceURLs: []string{}},
		"b2b":            {definition: "Business to business, i.e. we sell to other businesses", referenceURLs: []string{}},
		"b2c":            {definition: "Business to consumer, i.e. we sell directly to individual consumers", referenceURLs: []string{}},
		"ga":             {definition: "General availability, either \"lol we shipped to prod\" (unusual?) or \"it was approved by marketing, legal, product, there were design docs, etc.\"", referenceURLs: []string{}},
		"roi":            {definition: "Return on investment", referenceURLs: []string{}},
		"wow":            {definition: "Week over week, \"We currently have 2% WoW revenue growth\" means that last week had 2% more revenue than the previous week.", referenceURLs: []string{}},
		"kr":             {definition: "Key result, just means \"goal we will measure\"", referenceURLs: []string{}},
		"okr":            {definition: "Objective key result, just means \"goal we will measure\"", referenceURLs: []string{}},
		"q1":             {definition: "First quarter of the [fiscal] year.", referenceURLs: []string{"https://handbook.sourcegraph.com/departments/finance#fiscal-years"}},
		"fy22":           {definition: "Fiscal Year 2022. BEWARE: FY22 Q1 means the first quarter of *2021* not *2022*", referenceURLs: []string{"https://handbook.sourcegraph.com/departments/finance#fiscal-years"}},
		"nps score":      {definition: "one of those \"please rate our product on a scale of 1-10\" things.", referenceURLs: []string{}},
		"arr":            {definition: "Annual Recurring Revenue, if we have a contract with a customer over any timeframe, how much we get in a one year period.", referenceURLs: []string{}},
		"iarr":           {definition: "The change in ARR from one period to another.", referenceURLs: []string{}},
		"new iarr":       {definition: "IARR, but only including new customers", referenceURLs: []string{}},
		"expansion iarr": {definition: "IARR from customers who we already had at the start of the period", referenceURLs: []string{}},
		"booking":        {definition: "A customer committed to paying us money, a new customer signing a contract, a contract signed to expand the number of seats, etc.", referenceURLs: []string{}},
		"ap":             {definition: "Accounts Payable, i.e. our own bills we can pay right now", referenceURLs: []string{}},
		"ar":             {definition: "Accounts Receivable, bills customers should pay to us", referenceURLs: []string{"https://handbook.sourcegraph.com/departments/finance/process/ar"}},
		"cash":           {definition: "See reference", referenceURLs: []string{"https://handbook.sourcegraph.com/departments/finance#financial-planning-and-financial-statement-review"}},
		"accruals":       {definition: "See reference", referenceURLs: []string{"https://handbook.sourcegraph.com/departments/finance#financial-planning-and-financial-statement-review"}},
		"cla":            {definition: "Contributor license agreement, non-employees must sign this before contributing code", referenceURLs: []string{}},
		"ux":             {definition: "User experience, the experience a user would have going through a flow for example", referenceURLs: []string{}},
		"ui":             {definition: "User interface, how buttons look, the layout, etc.", referenceURLs: []string{}},
		"saas":           {definition: "Software as a service", referenceURLs: []string{}},
		"pql":            {definition: "Product Qualified lead", referenceURLs: []string{"https://handbook.sourcegraph.com/departments/bizops/process/product_led_growth"}},
		"plg":            {definition: "Product led growth", referenceURLs: []string{"https://handbook.sourcegraph.com/departments/bizops/process/product_led_growth"}},
		"mql":            {definition: "Marketing qualified lead", referenceURLs: []string{"https://handbook.sourcegraph.com/departments/bizops/process/product_led_growth"}},
		"sla":            {definition: "Service level agreement", referenceURLs: []string{}},
		"ae":             {definition: "1. Account executive - responsible for maintaining an enterprise customer account and our relationship with them\n2. Application engineer", referenceURLs: []string{}},
		"sdr":            {definition: "Sales Development representative - focused on reaching out to customers, determining if they're good leads to follow up on", referenceURLs: []string{}},
		"mba":            {definition: "Master of Business Administration - a degree, not to be confused with the NBA which is a basketball team", referenceURLs: []string{}},
		"gtm":            {definition: "Go to market", referenceURLs: []string{}},
		"cs":             {definition: "Customer Support", referenceURLs: []string{}},
		"em":             {definition: "Engineering manager", referenceURLs: []string{}},
		"pm":             {definition: "Product manager", referenceURLs: []string{}},
		"pd":             {definition: "1. Product designer\n2. Product document", referenceURLs: []string{}},
		"po":             {definition: "Purchase order", referenceURLs: []string{}},
		"cse":            {definition: "Customer Support Engineer, someone who can debug issues for example - no longer used, Application Engineer is preferred", referenceURLs: []string{}},
		"ce":             {definition: "Customer Engineer, e.g. helps customers deploy software and resolve their technical issues", referenceURLs: []string{}},
		"ic":             {definition: "Individual contributor, not managing other people. Still works on a team with other ICs.", referenceURLs: []string{}},
		"release guild":  {definition: "A captain of releasing the product, drives releases, gathers and informs others about the release, helps test the release, fixes and discovers issues in the release before it goes out.", referenceURLs: []string{}},
		"devrel":         {definition: "Developer relations, they post on Hacker News, Reddit, and Twitter about how cool we are. They give talks and go to conferences", referenceURLs: []string{}},
		"devexp":         {definition: "Developer experience, improving lives of devs working on sourcegraph", referenceURLs: []string{}},
		"devx":           {definition: "Developer experience, improving lives of devs working on sourcegraph", referenceURLs: []string{}},
		"dx":             {definition: "Developer experience, improving lives of devs working on sourcegraph", referenceURLs: []string{}},
		"people ops":     {definition: "The HR department, scheduling interviews etc.", referenceURLs: []string{}},
		"biz ops":        {definition: "Business operations, like financial modeling, managing legal aspects, etc", referenceURLs: []string{}},
		"it tech ops":    {definition: "IT team, if you have laptop issues or need to order a computer", referenceURLs: []string{}},
		"ops":            {definition: "Operations (at Sourcegraph, this includes the Finance and Accounting team, the Legal team, the People team, the Data & Analytics team, the Strategy team, and the Tech Ops team)", referenceURLs: []string{}},
		"dri":            {definition: "Directly responsible individual, the \"one true owner\" of something", referenceURLs: []string{}},
		"ptal":           {definition: "Please take a look", referenceURLs: []string{}},
		"lgtm":           {definition: "Looks good to me", referenceURLs: []string{}},
		"sgtm":           {definition: "Sounds good to me", referenceURLs: []string{}},
		"iiuc":           {definition: "If I understand correctly", referenceURLs: []string{}},
		"icymi":          {definition: "In case you missed it", referenceURLs: []string{}},
		"nbd":            {definition: "No big deal / not a big deal", referenceURLs: []string{}},
		"afaik":          {definition: "As far as I know", referenceURLs: []string{}},
		"lfg":            {definition: "Looking for Gutekanst (that's me) - sometimes Looking for Group", referenceURLs: []string{}},
		"yagni":          {definition: "You aren't gonna need it", referenceURLs: []string{}},
		"wg":             {definition: "Working group, a group of people working on a [typically] cross-functional project", referenceURLs: []string{}},
	}
)

func wutExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	termResult := dictionary[strings.ToLower(args[0])]

	if termResult.definition == "" {
		return errors.Newf("no definition found for term '%s'. Did you hear this term at Sourcegraph? Please consider adding it to our dictionary!", args[0])
	}

	stdout.Out.Writef(termResult.definition)

	if len(termResult.referenceURLs) > 0 {
		stdout.Out.Writef("\nreference:\n - %s", strings.Join(termResult.referenceURLs, "\n - "))
	}

	return nil
}
