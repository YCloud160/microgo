/*
 *
 * Copyright 2022 tars authors.
 *
 * Licensed under the Apache License, version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// protoc-gen-go-fifi is a plugin for the Google protocol buffer compiler to
// generate Go code. Install it by building this program and making it
// accessible within your PATH with the name:
//
//	protoc-gen-go-fifi
//
// The 'go-fifi' suffix becomes part of the argument for the protocol compiler,
// such that it can be invoked as:
//
//	protoc --go-fifi_out=. path/to/file.proto
//
// This generates Go service definitions for the protocol buffer defined by
// file.proto.  With that input, the output will be written to:
//
//	path/to/file_tars.pb.go
package main

import (
	"flag"
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	var (
		flags flag.FlagSet
		//plugins     = flag.String("plugins", "", "list of plugins to enable (supported values: grpc)")
		showVersion = flag.Bool("version", false, "print the version and exit")
	)
	if *showVersion {
		fmt.Printf("protoc-gen-go-microgo %v\n", version)
		return
	}

	protogen.Options{
		ParamFunc: func(name, value string) error {
			return flags.Set(name, value)
		},
	}.Run(func(gen *protogen.Plugin) error {
		microgo := true
		acmgo := true
		//microgo := false
		//acmgo := false
		//for _, plugin := range strings.Split(*plugins, ",") {
		//	switch plugin {
		//	case "microgo":
		//		microgo = true
		//	case "acmgo":
		//		acmgo = true
		//	case "":
		//	default:
		//		return fmt.Errorf("protoc-gen-go: unknown plugin %q", plugin)
		//	}
		//}
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			if microgo {
				GenerateMicroGoFile(gen, f)
			}
			if acmgo {
				GenerateAcmGoFile(gen, f)
			}
		}
		return nil
	})
}
