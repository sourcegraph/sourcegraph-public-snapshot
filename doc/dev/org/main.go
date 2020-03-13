package main

import (
	"fmt"

	"github.com/emicklei/dot"
)

// go run main.go | dot -Tpng  > test.png && open test.png

func main() {
	org := dot.NewGraph(dot.Directed)

	org.Label("Sourcegraph Org Graph")
	org.Attr("concentrate", "true")
	org.Attr("rankdir", "TD")
	org.Attr("ratio", "fill")
	org.Attr("ranksep", "1.4")
	org.Attr("nodesep", "0.4")

	const (
		reportsTo = "reports to"
		worksOn   = "works on"
		belongsTo = "belongs to"
	)

	people := make(map[string]dot.Node)
	for _, p := range []*struct {
		id   string
		name string
		role string
	}{
		{"sqs", "Quinn Slack", "CEO / Co-founder"},
		{"beyang", "Beyang Liu", "CTO / Co-founder"},
		{"christina", "Christina Forney", "Product Manager"},
		{"stephen", "Stephen Gutekanst", "Software Engineer"},
		{"rslade", "Ryan Slade", "Software Engineer"},
		{"rijnard", "Rijnard van Tonder", "Software Engineer"},
		{"eric", "Eric Fritz", "Software Engineer"},
		{"ericbm", "Eric Brody-Moore", "Growth & Bizops"},
		{"keegan", "Keegan Carruthers-Smith", "Software Engineer"},
		{"uwe", "Uwe Hoffmann", "Software Engineer"},
		{"joe", "Joe Chen", "Software Engineer"},
		{"vanesa", "Vanesa Ortiz", "Content Engineer"},
		{"noemi", "Noemi Mercado", "People Ops"},
		{"nick", "Nick Snyder", "VP Engineering"},
		{"thorsten", "Thorsten Ball", "Software Engineer"},
		{"loic", "Loïc Guychard", "Engineering Manager"},
		{"aileen", "Aileen Agricola", "Senior Digital\nMarketing Manager"},
		{"adam", "Adam Frankl", "Head of Marketing"},
		{"ryan", "Ryan Blunden", "Developer Advocate"},
		{"dan", "Dan Adler", "VP Business"},
		{"tomas", "Tomás Senart", "Engineering Manager"},
		{"michael", "Michael Fromberger", "Software Engineer"},
		{"herzog", "Adam Herzog", "Product Marketing\nManager"},
		{"geoffrey", "Geoffrey Gilmore", "Software Engineer"},
		{"farhan", "Farhan Attamimi", "Software Engineer"},
		{"erik", "Erik Seliger", "Software Engineer"},
		{"kai", "Kai Passo", "Account Executive"},
		{"julia", "Julia Gilinets", "Head of Sales"},
		{"tommy", "Tommy Roberts", "Product Designer"},
		{"felix", "Felix Becker", "Software Engineer"},
	} {
		people[p.id] = person(org, p.id, p.name, p.role)
	}

	teams := make(map[string]dot.Node)
	for _, m := range []*struct {
		id   string
		name string
	}{
		{"campaigns", "Campaigns"},
		{"search", "Search"},
		{"code-intel", "Code Intel"},
		{"core-services", "Core Services"},
		{"web", "Web"},
		{"distribution", "Distribution"},
		{"marketing", "Marketing"},
		{"product", "Product"},
		{"design", "Design"},
		{"sales", "Sales"},
		{"customer-success", "Customer\nSuccess"},
		{"biz-ops", "Business\nOperations"},
		{"people-ops", "People\nOperations"},
		{"eng-management", "Engineering\nManagement"},
	} {
		teams[m.id] = team(org, m.id, m.name)
	}

	people["rijnard"].Edge(people["tomas"], reportsTo)
	people["rijnard"].Edge(teams["search"], worksOn)
	people["rijnard"].Edge(teams["core-services"], worksOn)

	people["thorsten"].Edge(people["tomas"], reportsTo)
	people["thorsten"].Edge(teams["campaigns"], worksOn)
	people["thorsten"].Edge(teams["core-services"], worksOn)

	people["keegan"].Edge(people["tomas"], reportsTo)
	people["keegan"].Edge(teams["core-services"], worksOn)

	people["rslade"].Edge(people["tomas"], reportsTo)
	people["rslade"].Edge(teams["campaigns"], worksOn)
	people["rslade"].Edge(teams["core-services"], worksOn)

	people["joe"].Edge(people["tomas"], reportsTo)
	people["joe"].Edge(teams["core-services"], worksOn)

	people["farhan"].Edge(people["loic"], reportsTo)
	people["farhan"].Edge(teams["search"], worksOn)
	people["farhan"].Edge(teams["web"], worksOn)

	people["felix"].Edge(people["loic"], reportsTo)
	people["felix"].Edge(teams["web"], worksOn)

	people["erik"].Edge(people["loic"], reportsTo)
	people["erik"].Edge(teams["campaigns"], worksOn)
	people["erik"].Edge(teams["web"], worksOn)

	people["eric"].Edge(people["nick"], reportsTo)
	people["eric"].Edge(teams["code-intel"], worksOn)

	people["michael"].Edge(people["nick"], reportsTo)
	people["michael"].Edge(teams["code-intel"], worksOn)

	people["geoffrey"].Edge(people["nick"], reportsTo)
	people["geoffrey"].Edge(teams["distribution"], worksOn)

	people["uwe"].Edge(people["nick"], reportsTo)
	people["uwe"].Edge(teams["distribution"], worksOn)

	people["stephen"].Edge(people["nick"], reportsTo)
	people["stephen"].Edge(teams["distribution"], worksOn)

	people["beyang"].Edge(people["sqs"], reportsTo)
	people["beyang"].Edge(teams["distribution"], worksOn)
	people["beyang"].Edge(teams["eng-management"], worksOn)

	people["tomas"].Edge(teams["core-services"], worksOn)
	people["tomas"].Edge(teams["eng-management"], worksOn)
	people["tomas"].Edge(people["nick"], reportsTo)

	people["loic"].Edge(teams["web"], worksOn)
	people["loic"].Edge(teams["eng-management"], worksOn)
	people["loic"].Edge(people["nick"], reportsTo)

	people["nick"].Edge(people["sqs"], reportsTo)
	people["nick"].Edge(teams["eng-management"], worksOn)

	people["adam"].Edge(people["sqs"], reportsTo)
	people["adam"].Edge(teams["marketing"], worksOn)

	people["herzog"].Edge(people["adam"], reportsTo)
	people["herzog"].Edge(teams["marketing"], worksOn)

	people["aileen"].Edge(people["adam"], reportsTo)
	people["aileen"].Edge(teams["marketing"], worksOn)

	people["vanesa"].Edge(people["adam"], reportsTo)
	people["vanesa"].Edge(teams["marketing"], worksOn)

	people["ryan"].Edge(people["adam"], reportsTo)
	people["ryan"].Edge(teams["marketing"], worksOn)

	people["julia"].Edge(people["sqs"], reportsTo)
	people["julia"].Edge(teams["sales"], worksOn)

	people["kai"].Edge(people["julia"], reportsTo)
	people["kai"].Edge(teams["sales"], worksOn)

	people["dan"].Edge(people["sqs"], reportsTo)
	people["dan"].Edge(teams["customer-success"], worksOn)
	people["dan"].Edge(teams["biz-ops"], worksOn)
	people["dan"].Edge(teams["sales"], worksOn)

	people["ericbm"].Edge(people["dan"], reportsTo)
	people["ericbm"].Edge(teams["biz-ops"], worksOn)

	people["christina"].Edge(people["sqs"], reportsTo)
	people["christina"].Edge(teams["product"], worksOn)
	people["christina"].Edge(teams["campaigns"], worksOn)
	people["christina"].Edge(teams["search"], worksOn)
	people["christina"].Edge(teams["code-intel"], worksOn)
	people["christina"].Edge(teams["core-services"], worksOn)
	people["christina"].Edge(teams["web"], worksOn)

	people["tommy"].Edge(people["christina"], reportsTo)
	people["tommy"].Edge(teams["design"], worksOn)
	people["tommy"].Edge(teams["campaigns"], worksOn)
	people["tommy"].Edge(teams["search"], worksOn)
	people["tommy"].Edge(teams["code-intel"], worksOn)
	people["tommy"].Edge(teams["core-services"], worksOn)
	people["tommy"].Edge(teams["web"], worksOn)

	people["noemi"].Edge(people["dan"], reportsTo)
	people["noemi"].Edge(teams["people-ops"], worksOn)

	fmt.Println(org.String())
}

func person(org *dot.Graph, id, name, role string) dot.Node {
	return org.Node(id).
		Label(name+"\n"+role).
		Attr("shape", "ellipse").
		Attr("colorscheme", "set312").
		Attr("style", "filled").
		Attr("fillcolor", "3").
		Attr("fixedsize", "true").
		Attr("width", "2.2").
		Attr("height", "1.1")
}

func team(org *dot.Graph, id, name string) dot.Node {
	return org.Node(id).
		Label(name).
		Attr("shape", "house").
		Attr("colorscheme", "set312").
		Attr("style", "filled").
		Attr("fillcolor", "6")
}
