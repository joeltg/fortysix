# styx

> Home-grown graph database inspired by [Hexastore](https://dl.acm.org/citation.cfm?id=1453965), tailored for use in the [Underlay](https://underlay.mit.edu).

## What's going on here?

"Hexastore" is a name for an indexing scheme for RDF Triples (a triple is a `<subject | predicate | object>` statement), and it's based on the silly idea that if you really care about indexing your triples in some key/value-ish store, you shouldn't just insert them once: you should actually insert them six times - one for each permutation of the three elements (spo, pos, osp, sop, pso, ops). Even sillier, Styx goes ahead and performs 12 (twelve!) database writes (holy shit!) for every triple you want to insert, but these upfront space- and insertion-time- costs pay for a fast & general query interface based on subgraph matching.

## What even _is_ a 'Graph Database'?

In a big way, a database is _defined_ by its query language - your data is only as real as the interface you have to retrieve it (your money in the bank is only as real as your ability to withdraw). This raises the natural question of what that interface should look like, which is easier to answer the more constraints you have, and harder to answer to more general you'd like to be.

'Graphs' are about as general as they come. Traditional query languages for relational databases are already a vague amalgam of abstraction-breaking miscellany (Wikipedia describes [SQL](https://en.wikipedia.org/wiki/SQL) as four 'informally' classified sublanguages) and there's no clear consensus on what a query language should really offer (the Wikipedia page for [Query language](https://en.wikipedia.org/wiki/Query_language) "has multiple issues", like only being a list of forty-six examples). Mostly they just provide whatever developers end up wanting, which is great, but hard to translate into a new context, especially a _more general_ one like graphs where we're explicitly trying to spark new use cases.

Prior work on graph queries is scattered, and no existing languages have found the mass adoption that SQL did for relational databases (none of SPARQL, MQL, Gremlin, Gizmo, or Cypher are supported by more than a few implementations each; G-CORE has high aspirations but is only exists as a proposal for now). This could be attributed to a few causes:

- Less development pressure. Most data in most domains ends up being relatively constrained/structured/homogeneous. Graph databases are not commonly necessary and so develop more slowly.
- Generalization is hard. Graphs are both harder to represent with computers and harder to reason about as humans.
- "Query languages" don't exist. _(what?)_ Our everyday use of "query langauge" ends up either meaning "stuff you want to do with databases" or "the things that SQL can do" without any reified grounding of what those things actually are. It's not clear that there even is an idea to generalize in the first place.

frequently inventing abstractions

All of this contributes to a sense of openness and 'pre-Newtonian-ness' - that graphs and their interfaces are waiting for a stronger rigorous theoretic framework to clear up this

category theoretic vibes

This jibes with recent approaches to the space of data structures, It's weird to say something like "we don't really understand

An incremental step toward understanding graphs better is building higher-level primitives and exploring the conceptual space they create.
Styx takes the stance that one **natural first-order interface to graph storage is subgraph matching**. These terms were carefully chosen:

- _graph storage_ as the abstract role of a graph database
- _interface_ as the abstract role of a query language or API
- _first-order_ as a primitive of the graph storage implementation itself
- _natural_ in a mathematical sense of being well-defined and intuitive

## So what's subgraph matching?

It's a (limited) way of querying graphs. It's similar to the idea behind GraphQL - that "an easy and useful way of asking for things is by structing your question as isomorphic to the answers you expect". It means you can give the database a "pattern" subgraph that has some blank variables in it, and the database will give you back a "result" subgraph, _serialized the exact same way_, that has all the variables filled in with values. It's a generalization of both path traversal and constraint satisfaction.

More concretely,

This thing I'm calling "subgraph matching" is related to [subgraph isomorphism](https://en.wikipedia.org/wiki/Subgraph_isomorphism_problem), a well-studied problem in graph theory, except that in most application contexts we want to allow two different variables (in the pattern subgraph) to resolve to the same value (from the target graph). This technically breaks isomorphism, so it's really _surjective subgraph homomorphism_, or **_subgraph epimorphism_** if you really took notes in class.

## So... what's subgraph matching?

### Just give me some examples already

So if you wanted to know the father of the mayor of the hometown of the author of The Shining, you could query (in JSON-LD!) like the:

skos
ckos (computational knowledge organization system)

```json
{
  "@context": { "@vocab": "ex": "http://example.com/" },
  "@id": "ex:The_Shining",
  "author": {
    "hometown": {
      "mayor": {
        "father": {}
      }
    }
  }
}
```

... and you'd get a JSON response of:

```json
{
	"@context": { "@vocab": "http://example.com/" },
	"@id": "ex:The_Shining",
	"author": {
		"hometown": {
			"mayor": {
				"father": {
					"@id": "ex:The_Father"
				}
			}
		}
	}
}
```

And since there are a million ways of serializing the same graph, you could also have asked

```json
{
	"@context": { "@vocab": "http://example.com/" },
	"@graph": [
		{ "@id": "ex:The_Shining", "author": { "@id": "_:author" } },
		{ "@id": "_:author", "hometown": { "@id": "_:town" } },
		{ "@id": "_:town", "mayor": { "@id": "_:mayor" } },
		{ "@id": "_:mayor", "mayor": { "@id": "_:father" } }
	]
}
```

**Subgraph matching is more like a graph analog of a key-value store** than a query language itself: a reliable, conceptually clean intermediate interface that more specialized DSLs should build off. This is the purpose of the Styx project: to be an abstract graph store that exposes natural high-level graph primitives while minimizing loss of generality.

It has some great properties! It's provably general: the operation is defined for every pair of graphs (usually returning the empty graph)

Yes but that's _silly_, nobody wants to learn a new language and _worse_, it can't be assembled by computers. What does that mean? It means thta

## Query Processing

## Dynamism and beyond

```golang
// Open an IPFS Shell
sh := ipfs.NewShell("localhost:5001")

// Open a Badger database
path := "/tmp/badger"
opts := badger.DefaultOptions
opts.Dir = path
opts.ValueDir = path

db, err := badger.Open(opts)

// Ingest some data as JSON-LD
var data map[string]interface{}
var dataBytes = []byte(`{
  "@context": { "@vocab": "http://schema.org/" },
  "@graph": [
    {
      "@type": "Person",
      "name": "Joel",
      "birthDate": "1996-02-02",
      "children": { "@id": "http://people.com/liljoel" }
    },
    {
      "@id": "http://people.com/liljoel",
      "@type": "Person",
      "name": "Little Joel",
      "birthDate": "2030-11-10"
    }
  ]
}`)

json.Unmarshal(dataBytes, &data)
Ingest(data, db, sh)

// Query by subgraph pattern
var query map[string]interface{}
var queryBytes = []byte(`{
  "@context": {
    "@vocab": "http://schema.org/",
    "parent": {
      "@reverse": "children"
    }
  },
  "@type": "Person",
  "birthDate": {},
  "parent": {
    "name": "Joel"
  }
}`)

json.Unmarshal(queryBytes, &query)
Query(query, func(result interface{}) error {
  // The result will be framed by the query,
  // as per https://w3c.github.io/json-ld-framing
  bytes, _ := json.MarshalIndent(result, "", "\t")
  fmt.Println(string(bytes))
  return nil
}, db, sh)
```

When ingested, every JSON-LD document is first [normalized as an RDF dataset](https://json-ld.github.io/normalization/spec/).

```
<http://people.com/liljoel> <http://schema.org/birthDate> "2030-11-10" .
<http://people.com/liljoel> <http://schema.org/name> "Little Joel" .
<http://people.com/liljoel> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://schema.org/Person> .
_:c14n0 <http://schema.org/birthDate> "1996-02-02" .
_:c14n0 <http://schema.org/children> <http://people.com/liljoel> .
_:c14n0 <http://schema.org/name> "Joel" .
_:c14n0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://schema.org/Person> .
```

This (canonicalized) dataset has an IPFS CID of `QmWMwTL4GZSEsAaNYUo7Co24HkAkVCSdPgMwGJmrH5TwMC` (you can view it from [any gateway](https://gateway.underlay.store/ipfs/QmWMwTL4GZSEsAaNYUo7Co24HkAkVCSdPgMwGJmrH5TwMC) or in the [Underlay explorer](https://underlay.github.io/explore/#QmWMwTL4GZSEsAaNYUo7Co24HkAkVCSdPgMwGJmrH5TwMC)!). So when the query processor wants to reference a blank node from that dataset, it'll use a URI staring with `dweb:/ipfs/QmWMwTL4GZSEsAaNYUo7Co24HkAkVCSdPgMwGJmrH5TwMC`, plus a fragment identifier for the (canonicalized) blank node id.

```
{
  "@context": {
    "@vocab": "http://schema.org/",
    "parent": {
      "@reverse": "http://schema.org/children"
    }
  },
  "@id": "http://people.com/liljoel",
  "@type": "Person",
  "birthDate": "2030-11-10",
  "parent": {
    "@id": "dweb:/ipfs/QmWMwTL4GZSEsAaNYUo7Co24HkAkVCSdPgMwGJmrH5TwMC#_:c14n0",
    "name": "Joel"
  }
}
```
