package types

import (
	"fmt"
	"regexp"

	cid "github.com/ipfs/go-cid"
	multibase "github.com/multiformats/go-multibase"
)

// URI is an interface type for content-addressable semantic URIs
type URI interface {
	Parse(uri string) (c cid.Cid, fragment string)
	String(c cid.Cid, fragment string) (uri string)
	Test(uri string) bool
}

type underlayURI struct{}

// var testUlURI = regexp.MustCompile("^ul:\\/ipfs\\/([a-zA-Z0-9]{59})(#(?:_:c14n\\d+)?)?$")
var testUlURI = regexp.MustCompile("^u:([a-zA-Z0-9]{59})(#(?:_:c14n\\d+)?)?$")

func (*underlayURI) Parse(uri string) (c cid.Cid, fragment string) {
	if match := testUlURI.FindStringSubmatch(uri); match != nil {
		c, _ = cid.Decode(match[1])
		fragment = match[2]
	}
	return
}

func (*underlayURI) String(c cid.Cid, fragment string) (uri string) {
	s, _ := c.StringOfBase(multibase.Base32)
	// return fmt.Sprintf("ul:/ipfs/%s%s", s, fragment)
	return "u:" + s + fragment
}

func (*underlayURI) Test(uri string) bool {
	return testUlURI.MatchString(uri)
}

// UnderlayURI are URIs that use a u: protocol scheme
var UnderlayURI URI = (*underlayURI)(nil)

type queryURI struct{}

var testQueryURI = regexp.MustCompile("^q:([a-zA-Z0-9]{59})(#(?:_:c14n\\d+)?)?$")

func (*queryURI) Parse(uri string) (c cid.Cid, fragment string) {
	if match := testQueryURI.FindStringSubmatch(uri); match != nil {
		c, _ = cid.Decode(match[1])
		// Queries *must* be raw codecs for now
		if c.Prefix().Codec == cid.Raw {
			fragment = match[2]
		} else {
			c = cid.Undef
		}
	}
	return
}

func (*queryURI) String(c cid.Cid, fragment string) (uri string) {
	s, _ := c.StringOfBase(multibase.Base32)
	// return fmt.Sprintf("ul:/ipfs/%s%s", s, fragment)
	return "u:" + s + fragment
}

func (*queryURI) Test(uri string) bool {
	return testQueryURI.MatchString(uri)
}

// QueryURI are URIs that use a q: protocol scheme
var QueryURI URI = (*queryURI)(nil)

type dwebURI struct{}

var testDwebURI = regexp.MustCompile("^dweb:\\/ipfs\\/([a-zA-Z0-9]{59})$")

func (*dwebURI) Parse(uri string) (c cid.Cid, fragment string) {
	if match := testDwebURI.FindStringSubmatch(uri); match != nil {
		c, _ = cid.Decode(match[1])
	}
	return
}

func (*dwebURI) String(c cid.Cid, fragment string) (uri string) {
	s, _ := c.StringOfBase(multibase.Base32)
	return fmt.Sprintf("dweb:/ipfs/%s%s", s, fragment)
}

func (*dwebURI) Test(uri string) bool {
	return testDwebURI.MatchString(uri)
}

// DwebURI are URIs that use a dweb: protocol scheme
var DwebURI URI = (*dwebURI)(nil)
