package generalprobe

import (
	"encoding/json"
	"fmt"
	"io"
)

type cdkoutManifest struct {
	Artifacts map[string]*cdkoutArtifact `json:"artifacts"`
	Runtime   cdkoutRuntime              `json:"runtime"`
	Version   string                     `json:"version"`
}

type cdkoutArtifact struct {
	Environment string                       `json:"environment"`
	Metadata    map[string][]*cdkoutMetadata `json:"metadata"`
	Properties  cdkoutManifestProperties     `json:"properties"`
	Type        string                       `json:"type"`
}

type cdkoutManifestProperties struct {
	TemplateFile string `json:"templateFile"`
}

type cdkoutMetadata struct {
	Data  interface{} `json:"data"`
	Trace []string    `json:"trace"`
	Type  string      `json:"type"`
}

type cdkoutRuntime struct {
	Libraries map[string]string `json:"libraries"`
}

func newManifest(rd io.Reader) (*cdkoutManifest, error) {
	var manifest cdkoutManifest
	dec := json.NewDecoder(rd)
	if err := dec.Decode(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (x *cdkoutManifest) lookupLogicalID(stackName, resourceID string) (*string, error) {
	artifact, ok := x.Artifacts[stackName]
	if !ok {
		return nil, nil
	}

	resourcePath := fmt.Sprintf("/%s/%s/Resource", stackName, resourceID)
	metadataList, ok := artifact.Metadata[resourcePath]
	if !ok {
		return nil, nil
	}

	for _, metadata := range metadataList {
		if metadata.Type == "aws:cdk:logicalId" {
			logicalID, ok := metadata.Data.(string)
			if !ok {
				return nil, fmt.Errorf("Invalid data type (metadata.data is not string)")
			}
			return &logicalID, nil
		}
	}

	return nil, nil
}
