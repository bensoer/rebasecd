package utils

import (
	"bytes"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	yaml2 "sigs.k8s.io/yaml"
)

// DecodeYAMLManifest takes a multi-document YAML (like Helm rel.Manifest)
// and returns all decoded client.Objects.
func DecodeYAMLManifest(manifest string) ([]client.Object, error) {
	if len(manifest) == 0 {
		return nil, nil
	}

	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	reader := bytes.NewReader([]byte(manifest))

	docs, err := splitYAMLDocs(reader)
	if err != nil {
		return nil, fmt.Errorf("splitting YAML docs: %w", err)
	}

	var objs []client.Object
	for _, doc := range docs {
		if len(bytes.TrimSpace(doc)) == 0 {
			continue
		}

		u := &unstructured.Unstructured{}
		_, gvk, err := decoder.Decode(doc, nil, u)
		if err != nil {
			return nil, fmt.Errorf("decoding YAML doc: %w", err)
		}
		u.SetGroupVersionKind(*gvk)
		objs = append(objs, u)
	}

	return objs, nil
}

// splitYAMLDocs splits a multi-document YAML stream into individual []byte documents.
func splitYAMLDocs(r io.Reader) ([][]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Split on the YAML document separator `---`
	parts := bytes.Split(data, []byte("\n---"))
	var docs [][]byte
	for _, p := range parts {
		trimmed := bytes.TrimSpace(p)
		if len(trimmed) > 0 {
			docs = append(docs, trimmed)
		}
	}
	return docs, nil
}

func UnstructuredObjectsAreEqual(a, b client.Object) bool {
	// If they are Unstructured, remove system fields first
	aUn, aOk := a.(*unstructured.Unstructured)
	bUn, bOk := b.(*unstructured.Unstructured)

	if aOk && bOk {
		aCopy := aUn.DeepCopy()
		bCopy := bUn.DeepCopy()

		// Remove system-managed metadata fields
		for _, f := range []string{"resourceVersion", "uid", "creationTimestamp", "generation", "managedFields"} {
			unstructured.RemoveNestedField(aCopy.Object, "metadata", f)
			unstructured.RemoveNestedField(bCopy.Object, "metadata", f)
		}

		// Compare the object maps
		return equality.Semantic.DeepEqual(aCopy.Object, bCopy.Object)
	}

	// Fallback: marshal to YAML and compare
	aData, _ := yaml2.Marshal(a)
	bData, _ := yaml2.Marshal(b)
	return bytes.Equal(aData, bData)
}
