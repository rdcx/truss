package gengokit

import (
	"bytes"
	"io"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/TuneLab/go-truss/deftree"

	templateFileAssets "github.com/TuneLab/go-truss/gengokit/template"
)

func TestTemplatePathToActual(t *testing.T) {

	pathToWants := map[string]string{
		"NAME-service/":                "package-service/",
		"NAME-service/test.gotemplate": "package-service/test.go",
		"NAME-service/NAME-server":     "package-service/package-server",
	}

	for path, want := range pathToWants {
		if got := templatePathToActual(path, "package"); got != want {
			t.Errorf("\n`%v` got\n`%v` wanted", got, want)
		}
	}
}

func TestNewTemplateExecutor(t *testing.T) {
	const def = `
		syntax = "proto3";

		// General package
		package general;

		import "google/api/annotations.proto";

		// RequestMessage is so foo
		message RequestMessage {
		  string input = 1;
		}

		// ResponseMessage is so bar
		message ResponseMessage {
		  string output = 1;
		}

		// ProtoService is a service
		service ProtoService {
		  // ProtoMethod is simple. Like a gopher.
		  rpc ProtoMethod (RequestMessage) returns (ResponseMessage) {
			// No {} in path and no body, everything is in the query
			option (google.api.http) = {
			  get: "/route"
			};
		  }
		}
	`
	dt, err := deftree.NewFromString(def)
	if err != nil {
		t.Error(err)
	}

	const goImportPath = "github.com/TuneLab/go-truss/gengokit"

	te, err := newTemplateExecutor(dt, goImportPath)

	if err != nil {
		t.Error(err)
	}

	if got, want := te.ImportPath, goImportPath+"/general-service"; got != want {
		t.Errorf("\n`%v` was ImportPath\n`%v` was wanted", got, want)
	}

	if got, want := te.PackageName, dt.GetName(); got != want {
		t.Errorf("\n`%v` was PackageName\n`%v` was wanted", got, want)
	}
}

func TestGetProtoService(t *testing.T) {
	const def = `
		syntax = "proto3";

		// General package
		package general;

		import "google/api/annotations.proto";

		// RequestMessage is so foo
		message RequestMessage {
		  string input = 1;
		}

		// ResponseMessage is so bar
		message ResponseMessage {
		  string output = 1;
		}

		// ProtoService is a service
		service ProtoService {
		  // ProtoMethod is simple. Like a gopher.
		  rpc ProtoMethod (RequestMessage) returns (ResponseMessage) {
			// No {} in path and no body, everything is in the query
			option (google.api.http) = {
			  get: "/route"
			};
		  }
		}
	`
	dt, err := deftree.NewFromString(def)
	if err != nil {
		t.Error(err)
	}

	svc, err := getProtoService(dt)
	if err != nil {
		t.Error(err)
	}

	if got, want := svc.GetName(), "ProtoService"; got != want {
		t.Errorf("\n`%v` was service name\n`%v` was wanted", got, want)
	}

	if got, want := svc.Methods[0].GetName(), "ProtoMethod"; got != want {
		t.Errorf("\n`%v` was rpc in service\n`%v` was wanted", got, want)
	}
}

func TestApplyTemplateFromPath(t *testing.T) {
	const def = `
		syntax = "proto3";

		// General package
		package general;

		import "google/api/annotations.proto";

		// RequestMessage is so foo
		message RequestMessage {
		  string input = 1;
		}

		// ResponseMessage is so bar
		message ResponseMessage {
		  string output = 1;
		}

		// ProtoService is a service
		service ProtoService {
		  // ProtoMethod is simple. Like a gopher.
		  rpc ProtoMethod (RequestMessage) returns (ResponseMessage) {
			// No {} in path and no body, everything is in the query 
			option (google.api.http) = { 
				get: "/route"
			};
		  }
		}
	`

	const goImportPath = "github.com/TuneLab/go-truss/gengokit"

	dt, err := deftree.NewFromString(def)
	if err != nil {
		t.Error(err)
	}

	te, err := newTemplateExecutor(dt, goImportPath)
	if err != nil {
		t.Error(err)
	}

	end, err := applyTemplateFromPath("NAME-service/generated/endpoints.gotemplate", te)
	if err != nil {
		t.Error(err)
	}

	endCode, err := ioutil.ReadAll(end)
	if err != nil {
		t.Error(err)
	}

	_, err = formatCode(endCode)
	if err != nil {
		t.Error(err)
	}

}

func TestUpdateServerMethods(t *testing.T) {
	const def = `
		syntax = "proto3";

		// General package
		package general;

		import "google/api/annotations.proto";

		// RequestMessage is so foo
		message RequestMessage {
		  string input = 1;
		}

		// ResponseMessage is so bar
		message ResponseMessage {
		  string output = 1;
		}

		// ProtoService is a service
		service ProtoService {
		  // ProtoMethod is simple. Like a gopher.
		  rpc ProtoMethod (RequestMessage) returns (ResponseMessage) {
			// No {} in path and no body, everything is in the query 
			option (google.api.http) = { 
				get: "/route"
			};
		  }
		}
	`

	const goImportPath = "github.com/TuneLab/go-truss/gengokit"

	dt, err := deftree.NewFromString(def)
	if err != nil {
		t.Error(err)
	}

	te, err := newTemplateExecutor(dt, goImportPath)
	if err != nil {
		t.Error(err)
	}

	// apply server_handler.go template
	sh, err := applyTemplateFromPath("NAME-service/handlers/server/server_handler.gotemplate", te)
	if err != nil {
		t.Error(err)
	}

	// read the code off the io.Reader
	shCode, err := ioutil.ReadAll(sh)
	if err != nil {
		t.Error(err)
	}

	// format the code
	shCode, err = formatCode(shCode)
	if err != nil {
		t.Error(err)
	}

	// updateServerMethods with the same templateExecutor
	same, err := updateServerMethods(bytes.NewReader(shCode), te)
	if err != nil {
		t.Error(err)
	}

	// read the new code off the io.Reader
	sameCode, err := ioutil.ReadAll(same)
	if err != nil {
		t.Error(err)
	}

	// format that new code
	sameCode, err = formatCode(sameCode)
	if err != nil {
		t.Error(err)
	}

	// make sure the code is the same
	if bytes.Compare(shCode, sameCode) != 0 {
		t.Fatalf("\n__BEFORE__\n\n%v\n\n__AFTER\n\n%v\n\nCode before and after updating differs",
			string(shCode), string(sameCode))
	}
}

func TestAllTemplates(t *testing.T) {
	const goImportPath = "github.com/TuneLab/go-truss/gengokit"

	const def = `
		syntax = "proto3";

		// General package
		package general;

		import "google/api/annotations.proto";

		// RequestMessage is so foo
		message RequestMessage {
		  string input = 1;
		}

		// ResponseMessage is so bar
		message ResponseMessage {
		  string output = 1;
		}

		// ProtoService is a service
		service ProtoService {
		  // ProtoMethod is simple. Like a gopher.
		  rpc ProtoMethod (RequestMessage) returns (ResponseMessage) {
			// No {} in path and no body, everything is in the query 
			option (google.api.http) = { 
				get: "/route"
			};
		  }
		}
	`

	const def2 = `
		syntax = "proto3";

		// General package
		package general;

		import "google/api/annotations.proto";

		// RequestMessage is so foo
		message RequestMessage {
		  string input = 1;
		}

		// ResponseMessage is so bar
		message ResponseMessage {
		  string output = 1;
		}

		// ProtoService is a service
		service ProtoService {
		  // ProtoMethod is simple. Like a gopher.
		  rpc ProtoMethod (RequestMessage) returns (ResponseMessage) {
			// No {} in path and no body, everything is in the query 
			option (google.api.http) = { 
				get: "/route"
			};
		  }
		  // ProtoMethodAgain is simple. Like a gopher again.
		  rpc ProtoMethodAgain (RequestMessage) returns (ResponseMessage) {
			// No {} in path and no body, everything is in the query 
			option (google.api.http) = { 
				get: "/route2"
			};
		  }
		}
	`

	dt, err := deftree.NewFromString(def)
	if err != nil {
		t.Error(err)
	}

	te, err := newTemplateExecutor(dt, goImportPath)
	if err != nil {
		t.Error(err)
	}

	dt2, err := deftree.NewFromString(def2)
	if err != nil {
		t.Error(err)
	}

	te2, err := newTemplateExecutor(dt2, goImportPath)
	if err != nil {
		t.Error(err)
	}

	for _, templFP := range templateFileAssets.AssetNames() {
		// skip the partial templates
		if filepath.Ext(templFP) != ".gotemplate" {
			continue
		}
		prevGenMap := make(map[string]io.Reader)

		firstCode, err := testGenerateResponseFile(templFP, te, prevGenMap)
		if err != nil {
			t.Fatalf("%v failed to format on first generation\n\nERROR:\n\n%v\n\nCODE:\n\n%v", templFP, err.Error(), string(firstCode))
		}

		// store the file to act to pass back to testGenerateResponseFile for second generation
		prevGenMap[templatePathToActual(templFP, te.PackageName)] = bytes.NewReader(firstCode)

		secondCode, err := testGenerateResponseFile(templFP, te, prevGenMap)
		if err != nil {
			t.Fatalf("%v failed to format on second identical generation\n\nERROR:\n\n%v\n\nCODE:\n\n%v", templFP, err.Error(), string(secondCode))
		}

		if bytes.Compare(firstCode, secondCode) != 0 {
			t.Fatalf("\n__BEFORE__\n\n%v\n\n__AFTER\n\n%v\n\nCode differs after being regenerated with same definition file",
				string(firstCode), string(secondCode))
		}

		// store the file to act to pass back to testGenerateResponseFile for third generation
		prevGenMap[templatePathToActual(templFP, te.PackageName)] = bytes.NewReader(secondCode)

		// pass in templateExecutor created from def2
		addRPCCode, err := testGenerateResponseFile(templFP, te2, prevGenMap)
		if err != nil {
			t.Fatalf("%v failed to format on third generation with 1 rpc added\n\nERROR:\n\n%v\n\nCODE:\n\n%v", templFP, err.Error(), string(addRPCCode))
		}

		// store the file to act to pass back to testGenerateResponseFile for forth generation
		prevGenMap[templatePathToActual(templFP, te.PackageName)] = bytes.NewReader(addRPCCode)

		// pass in templateExecutor create from def1
		_, err = testGenerateResponseFile(templFP, te, prevGenMap)
		if err != nil {
			t.Fatalf("%v failed to format on forth generation with 1 rpc removed\n\nERROR:\n\n%v\n\nCODE:\n\n%v", templFP, err.Error(), string(addRPCCode))
		}
	}
}

func testGenerateResponseFile(templFP string, te *templateExecutor, prevGenMap map[string]io.Reader) ([]byte, error) {
	// apply server_handler.go template
	code, err := generateResponseFile(templFP, te, prevGenMap)
	if err != nil {
		return nil, err
	}

	// read the code off the io.Reader
	codeBytes, err := ioutil.ReadAll(code)
	if err != nil {
		return nil, err
	}

	// format the code
	codeBytes, err = formatCode(codeBytes)
	if err != nil {
		return nil, err
	}

	return codeBytes, nil
}
