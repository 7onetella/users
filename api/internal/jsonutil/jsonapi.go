package jsonutil

import (
	"bytes"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/mfcochauxlaberge/jsonapi"
)

func SchemaCheck(v interface{}) (*jsonapi.Schema, []error) {
	schema := &jsonapi.Schema{}
	schema.AddType(jsonapi.MustBuildType(v))
	errors := schema.Check()
	return schema, errors
}

func MarshalDoc(requestURI string, schema *jsonapi.Schema, doc *jsonapi.Document) (string, error) {
	url, err := jsonapi.NewURLFromRaw(schema, requestURI)
	if err != nil {
		return "", err
	}
	payload, err := jsonapi.MarshalDocument(doc, url)
	if err != nil {
		return "", err
	}
	out := &bytes.Buffer{}
	out.Write(payload)
	return out.String(), nil
}

func NewCollectionDoc(users []User) *jsonapi.Document {
	doc := &jsonapi.Document{}
	sample := jsonapi.Wrap(&User{})
	wc := jsonapi.WrapCollection(sample)

	for i, _ := range users {
		w := jsonapi.Wrap(&users[i])
		wc.Add(w)
	}

	// this causes bug &u
	//for _, u := range users {
	//	w := jsonapi.Wrap(&u)
	//	wc.Add(w)
	//}

	doc.Data = wc
	return doc
}

func NewJSONDoc(user User) *jsonapi.Document {
	doc := &jsonapi.Document{}
	doc.Data = jsonapi.Wrap(&user)
	return doc
}
