package main

import (
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"strings"
)

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	acmgoPackage = protogen.GoImportPath("github.com/luban-cloud/utils/acm")
	jsonPackage  = protogen.GoImportPath("encoding/json")
)

// GenerateAcmGoFile generates a _acmgo.pb.go file containing acmgo service definitions.
func GenerateAcmGoFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	//if !strings.HasSuffix(file.GeneratedFilenamePrefix, "acm.int") && !strings.HasSuffix(file.GeneratedFilenamePrefix, "acm.ext") {
	//	return nil
	//}
	isGenerate := checkGenerateAcm(file)
	if isGenerate == false {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_acmgo.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// versions:")
	g.P("// - protoc             ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	t := acmgo{}
	t.Init(g)
	t.Generate(file)
	return g
}

// acmgo is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for tars rpc support.
type acmgo struct {
	gen *protogen.GeneratedFile
}

// Name returns the name of this plugin
func (t *acmgo) Name() string {
	return "acmgo"
}

// Init initializes the plugin.
func (t *acmgo) Init(gen *protogen.GeneratedFile) {
	t.gen = gen
}

// GenerateImports generates the import declaration for this file.
func (t *acmgo) GenerateImports(file *protogen.File) {
	t.gen.QualifiedGoIdent(acmgoPackage.Ident("RegisterIAcmData"))
	t.gen.QualifiedGoIdent(jsonPackage.Ident("Marshal"))
}

// P forwards to g.gen.P.
func (t *acmgo) P(args ...interface{}) { t.gen.P(args...) }

// Generate generates code for the services in the given file.
func (t *acmgo) Generate(file *protogen.File) {
	t.GenerateImports(file)

	t.P()

	t.generateAcm(file)
}

func checkGenerateAcm(file *protogen.File) bool {
	for _, message := range file.Proto.MessageType {
		if len(message.ReservedName) == 4 && message.ReservedName[0] == "acm" {
			return true
		}
	}
	return false
}

// generateAcm
func (t *acmgo) generateAcm(file *protogen.File) {
	i := strings.LastIndex(file.GeneratedFilenamePrefix, ".")
	nameSuffix := file.GeneratedFilenamePrefix
	if i >= 0 {
		nameSuffix = file.GeneratedFilenamePrefix[i+1:]
	}
	nameSuffix = upperFirstLatter(nameSuffix)

	var datas []string
	var servers = make(map[string][]string, 0)
	for _, message := range file.Proto.MessageType {
		if len(message.ReservedName) == 4 && message.ReservedName[0] == "acm" {
			name := *message.Name
			dataId := message.ReservedName[2]
			dataIdKey := name + "Key"
			dataName := message.ReservedName[3]
			datas = append(datas, dataIdKey, dataId, dataName)
			srvs := strings.Split(message.ReservedName[1], ",")
			for _, srv := range srvs {
				srv = upperFirstLatter(srv)
				if _, ok := servers[srv]; ok {
					servers[srv] = append(servers[srv], name)
				} else {
					servers[srv] = []string{name}
				}
			}
			t.generateAcmData(name, dataIdKey, dataName)
		}
	}
	t.P()
	if len(datas) >= 2 {
		t.P("const (")
		for i := 2; i < len(datas); i = i + 3 {
			t.P(fmt.Sprintf("%s = \"%s\"", datas[i-2], datas[i-1]))
		}
		t.P(")")
		t.P()
		t.P("var Acm", nameSuffix, "Map = map[string]string{")
		for i := 2; i < len(datas); i = i + 3 {
			t.P(fmt.Sprintf("%s: \"%s\",", datas[i-2], datas[i]))
		}
		t.P("}")
		t.P()
	}
	if len(servers) > 0 {
		for srv, messages := range servers {
			t.P(fmt.Sprintf("var %s%sIAcmDatas = []acm.IAcmData {", srv, nameSuffix))
			for _, message := range messages {
				t.P(fmt.Sprintf("&%s{},", message))
			}
			t.P("}")
			t.P()
		}
	}
}

func (t *acmgo) generateAcmData(name, dataIdKey, dataName string) {
	t.P(fmt.Sprintf("func (*%s) DataId() string {", name))
	t.P("return ", dataIdKey)
	t.P("}")
	t.P()

	t.P(fmt.Sprintf("func (*%s) Name() string {", name))
	t.P(fmt.Sprintf("return \"%s\"", dataName))
	t.P("}")
	t.P()

	t.P(fmt.Sprintf("func (x *%s) Marshal() (string, error) {", name))
	t.P("bs, err := json.Marshal(x)")
	t.P("if err != nil {")
	t.P("return \"\", err")
	t.P("}")
	t.P("return string(bs), nil")
	t.P("}")
	t.P()

	t.P(fmt.Sprintf("func (x *%s) Unmarshal(content string) (interface{}, error) {", name))
	t.P(fmt.Sprintf("data := &%s{}", name))
	t.P("if err := json.Unmarshal([]byte(content), data); err != nil {")
	t.P("return data, err")
	t.P("}")
	t.P("return data, nil")
	t.P("}")
	t.P()

	t.P(fmt.Sprintf("func Push%sToACM(data *%s) error {", name, name))
	t.P("return acm.Push(data)")
	t.P("}")
	t.P()

	t.P(fmt.Sprintf(`func Get%sFromACM() (*%s, error) {
		val, err := acm.Pull(&%s{})	
		if err != nil {
			return nil, err
		}
		if data, ok := val.(*%s); ok {
			return data, nil
		}
		return &%s{}, nil
		}`, name, name, name, name, name))
	t.P()
}
