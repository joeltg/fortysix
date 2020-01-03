package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	ipfs "github.com/ipfs/go-ipfs-http-client"
	ld "github.com/underlay/json-gold/ld"

	styx "github.com/underlay/styx/db"
	types "github.com/underlay/styx/types"
)

var sampleDataBytes = []byte(`{
	"@context": {
		"@vocab": "http://schema.org/",
		"xsd": "http://www.w3.org/2001/XMLSchema#",
		"prov": "http://www.w3.org/ns/prov#",
		"prov:generatedAtTime": { "@type": "xsd:dateTime" },
		"birthDate": { "@type": "xsd:date" }
	},
	"prov:generatedAtTime": "2019-07-24T16:46:05.751Z",
	"@graph": {
		"@type": "Person",
		"name": ["John Doe", "Johnny Doe"],
		"birthDate": "1996-02-02",
		"knows": {
			"@id": "http://people.com/jane",
			"@type": "Person",
			"name": "Jane Doe",
			"familyName": { "@value": "Doe", "@language": "en" },
			"birthDate": "1995-01-01"
		}
	}
}`)

var sampleDataBytes2 = []byte(`{
	"@context": {
		"@vocab": "http://schema.org/",
		"xsd": "http://www.w3.org/2001/XMLSchema#",
		"birthDate": { "@type": "xsd:date" }
	},
	"@type": "Person",
	"name": "Johnanthan Appleseed",
	"birthDate": "1780-01-10",
	"knows": { "@id": "http://people.com/jane" }
}`)

var sampleData, sampleData2 map[string]interface{}

var httpAPI *ipfs.HttpApi

func init() {
	var err error
	_ = json.Unmarshal(sampleDataBytes, &sampleData)
	_ = json.Unmarshal(sampleDataBytes2, &sampleData2)
	httpAPI, err = ipfs.NewURLApiWithClient("http://localhost:5001", http.DefaultClient)
	if err != nil {
		log.Fatalln(err)
	}
}

func TestIngest(t *testing.T) {
	// Remove old db
	fmt.Println("removing path", styx.DefaultPath)
	err := os.RemoveAll(styx.DefaultPath)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	key, err := httpAPI.Key().Self(ctx)
	if err != nil {
		t.Fatal(err)
	}

	db, err := styx.OpenDB(styx.DefaultPath, key.ID().String(), httpAPI)
	if err != nil {
		t.Error(err)
		return
	}

	defer db.Close()

	if err = db.IngestJSONLd(ctx, sampleData); err != nil {
		t.Error(err)
		return
	}

	if err = db.Log(); err != nil {
		t.Error(err)
	}
}

func testQuery(query string, data ...interface{}) (db *styx.DB, pattern []*ld.Quad, err error) {
	// Remove old db
	fmt.Println("removing path", styx.DefaultPath)
	err = os.RemoveAll(styx.DefaultPath)
	if err != nil {
		return
	}

	ctx := context.Background()

	key, err := httpAPI.Key().Self(ctx)
	if err != nil {
		return
	}

	id := key.ID().String()

	db, err = styx.OpenDB(styx.DefaultPath, id, httpAPI)
	if err != nil {
		return
	}

	for _, d := range data {
		err = db.IngestJSONLd(ctx, d)
		if err != nil {
			return
		}
	}

	db.Log()

	var queryData map[string]interface{}
	err = json.Unmarshal([]byte(query), &queryData)
	if err != nil {
		return
	}

	proc := ld.NewJsonLdProcessor()
	opts := ld.NewJsonLdOptions("")
	opts.DocumentLoader = db.Opts.DocumentLoader
	opts.ProduceGeneralizedRdf = true
	opts.Algorithm = types.Algorithm

	dataset, err := proc.ToRDF(queryData, opts)
	if err != nil {
		return
	}

	d, _ := dataset.(*ld.RDFDataset)
	pattern = d.GetQuads("@default")
	return
}

func TestSPO(t *testing.T) {
	db, pattern, err := testQuery(`{
	"@context": { "@vocab": "http://schema.org/" },
	"@id": "http://people.com/jane",
	"name": { }
}`, sampleData)
	defer db.Close()
	if err != nil {
		t.Error(err)
		return
	}

	cursor, err := db.Query(pattern, nil, nil)
	defer cursor.Close()
	if err != nil {
		t.Error(err)
		return
	}

	if cursor != nil {
		for d := cursor.Domain(); d != nil; d, err = cursor.Next(nil) {
			for _, b := range d {
				fmt.Printf("%s: %s\n", b.Attribute, cursor.Get(b).GetValue())
			}
			fmt.Println("---")
		}
	}
}

func TestOPS(t *testing.T) {
	db, pattern, err := testQuery(`{
	"@context": { "@vocab": "http://schema.org/" },
	"name": "Jane Doe"
}`, sampleData)
	defer db.Close()
	if err != nil {
		t.Error(err)
		return
	}

	cursor, err := db.Query(pattern, nil, nil)
	defer cursor.Close()
	if err != nil {
		t.Error(err)
		return
	}

	if cursor != nil {
		for d := cursor.Domain(); d != nil; d, err = cursor.Next(nil) {
			for _, b := range d {
				fmt.Printf("%s: %s\n", b.Attribute, cursor.Get(b).GetValue())
			}
			fmt.Println("---")
		}
	}
}

func TestSimpleQuery(t *testing.T) {
	db, pattern, err := testQuery(`{
	"@context": { "@vocab": "http://schema.org/" },
	"@type": "Person",
	"birthDate": { }
}`, sampleData)
	defer db.Close()
	if err != nil {
		t.Error(err)
		return
	}

	cursor, err := db.Query(pattern, nil, nil)
	defer cursor.Close()
	if err != nil {
		t.Error(err)
		return
	}

	if cursor != nil {
		for d := cursor.Domain(); d != nil; d, err = cursor.Next(nil) {
			for _, b := range d {
				fmt.Printf("%s: %s\n", b.Attribute, cursor.Get(b).GetValue())
			}
			fmt.Println("---")
		}
	}
}

func TestSimpleQuery2(t *testing.T) {
	db, pattern, err := testQuery(`{
	"@context": { "@vocab": "http://schema.org/" },
	"@type": "Person",
	"knows": {
		"name": "Jane Doe"
	}
}`, sampleData)
	defer db.Close()
	if err != nil {
		t.Error(err)
		return
	}

	cursor, err := db.Query(pattern, nil, nil)
	if cursor != nil {
		defer cursor.Close()
		if err != nil {
			t.Error(err)
			return
		}
		for d := cursor.Domain(); d != nil; d, err = cursor.Next(nil) {
			for _, b := range d {
				fmt.Printf("%s: %s\n", b.Attribute, cursor.Get(b).GetValue())
			}
			fmt.Println("---")
		}
	}
}

func TestDomainQuery(t *testing.T) {
	db, pattern, err := testQuery(`{
	"@context": { "@vocab": "http://schema.org/" },
	"@id": "_:b0",
	"name": { "@id": "_:b1" }
}`, sampleData, sampleData2)
	defer db.Close()
	if err != nil {
		t.Error(err)
		return
	}

	node := ld.NewBlankNode("_:b0")
	domain := []*ld.BlankNode{node}
	cursor, err := db.Query(pattern, domain, nil)
	if cursor != nil {
		defer cursor.Close()
		if err != nil {
			t.Error(err)
			return
		}
		for d := cursor.Domain(); d != nil; d, err = cursor.Next(node) {
			for _, b := range d {
				fmt.Printf("%s: %s\n", b.Attribute, cursor.Get(b).GetValue())
			}
			fmt.Println("---")
		}
	}
}

func TestIndexQuery(t *testing.T) {
	db, pattern, err := testQuery(`{
		"@context": { "@vocab": "http://schema.org/" },
		"@id": "_:b0",
		"name": { "@id": "_:b1" }
	}`, sampleData, sampleData2)
	defer db.Close()
	if err != nil {
		t.Error(err)
		return
	}

	domain := []*ld.BlankNode{
		ld.NewBlankNode("_:b0"),
		ld.NewBlankNode("_:b1"),
	}
	index := []ld.Node{
		ld.NewIRI("u:bafkreichbq6iklce3y64lntglbcw6grdmote5ptsxph4c5vm3j77br5coa#_:c14n1"),
		ld.NewLiteral("Johnny Doe", "", ""),
	}
	cursor, err := db.Query(pattern, domain, index)
	if cursor != nil {
		defer cursor.Close()
		if err != nil {
			t.Error(err)
			return
		}
		for d := cursor.Domain(); d != nil; d, err = cursor.Next(nil) {
			for _, b := range d {
				fmt.Printf("%s: %s\n", b.Attribute, cursor.Get(b).GetValue())
			}
			fmt.Println("---")
		}
	}
}
