package mapfs

import (
	"strings"
)

var initFiles = map[string]string{
	"init.luau":        "ModuleScript",
	"init.lua":         "ModuleScript",
	"init.server.luau": "Script",
	"init.server.lua":  "Script",
	"init.client.luau": "LocalScript",
	"init.client.lua":  "LocalScript",
}

var leafSuffixes = []struct {
	suffix    string
	className string
}{
	{".server.luau", "Script"},
	{".server.lua", "Script"},
	{".client.luau", "LocalScript"},
	{".client.lua", "LocalScript"},
	{".luau", "ModuleScript"},
	{".lua", "ModuleScript"},
}

const forbiddenChars = `:*?"<>|/\`

func ClassifySourceFile(fileName, parentDirName string) (Classified, error) {
	if class, ok := initFiles[fileName]; ok {
		if err := validateStem(parentDirName); err != nil {
			return Classified{}, err
		}
		return Classified{InstanceName: parentDirName, ClassName: class, Init: true}, nil
	}
	for _, spec := range leafSuffixes {
		if strings.HasSuffix(fileName, spec.suffix) {
			stem := strings.TrimSuffix(fileName, spec.suffix)
			if err := validateStem(stem); err != nil {
				return Classified{}, err
			}
			return Classified{InstanceName: stem, ClassName: spec.className, Init: false}, nil
		}
	}
	return Classified{}, errNotSource
}

var errNotSource = &Error{Code: "", Message: "not a source file"}

func validateStem(stem string) error {
	if stem == "" || stem == "." || stem == ".." {
		return projectInvalid("forbidden file stem")
	}
	if strings.ContainsAny(stem, forbiddenChars) {
		return projectInvalid("forbidden file stem")
	}
	return nil
}

func InverseDiskRelative(className, instanceName string, mappedChildren, existingLua bool) string {
	if className == "Folder" {
		return instanceName + "/"
	}
	ext := ".luau"
	if existingLua {
		ext = ".lua"
	}
	if mappedChildren {
		switch className {
		case "Script":
			return instanceName + "/init.server" + ext
		case "LocalScript":
			return instanceName + "/init.client" + ext
		default:
			return instanceName + "/init" + ext
		}
	}
	switch className {
	case "Script":
		return instanceName + ".server" + ext
	case "LocalScript":
		return instanceName + ".client" + ext
	default:
		return instanceName + ext
	}
}

func normalizeSource(source string) string {
	source = strings.TrimPrefix(source, "\uFEFF")
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	return source
}

func stripBOM(source string) string {
	return strings.TrimPrefix(source, "\uFEFF")
}

func splitNodeID(nodeID string) (rootID, rest string, err error) {
	i := strings.IndexByte(nodeID, '/')
	if i <= 0 || i == len(nodeID)-1 {
		return "", "", notMapped("nodeId outside configured roots")
	}
	return nodeID[:i], nodeID[i+1:], nil
}
