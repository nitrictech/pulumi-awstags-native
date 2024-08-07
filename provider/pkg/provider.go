// Copyright 2016-2023, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"fmt"

	"github.com/nitrictech/pulumi-awstags-native/provider/pkg/aws"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/middleware/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
)

// Version is initialized by the Go linker to contain the semver of this build.
var Version string

const Name string = "awstags"

func Provider() p.Provider {
	// We tell the provider what resources it needs to support.
	return infer.Provider(infer.Options{
		Resources: []infer.InferredResource{
			infer.Resource[aws.ResourceTag, aws.ResourceTagArgs, aws.ResourceTagState](),
		},
		ModuleMap: map[tokens.ModuleName]tokens.ModuleName{
			"provider": "index",
		},
		Metadata: schema.Metadata{
			Description: "The AWS tags provider enables you to manage tags on already deployed or imported AWS resources.",
			DisplayName: "Aws Tags",
			Keywords: []string{
				"pulumi",
				"awstags",
				"kind/native",
			},
			Homepage:          "https://github.com/nitrictech/pulumi-awstags",
			Repository:        "https://github.com/nitrictech/pulumi-awstags",
			Publisher:         "Nitric",
			LogoURL:           "",
			License:           "MIT",
			PluginDownloadURL: fmt.Sprintf("https://github.com/nitrictech/pulumi-awstags/releases/download/v%s/pulumi-awstags-v%s.tgz", Version, Version),
			LanguageMap: map[string]any{
				"nodejs": map[string]any{
					"packageName":        "@nitric/pulumi-awstags",
					"packageDescription": "A pulumi provider that manages awstags resources",
					"dependencies": map[string]string{
						"@pulumi/pulumi": "^3.0.0",
					},
				},
				"go": map[string]any{
					"generateResourceContainerTypes": true,
					"importBasePath":                 "github.com/nitrictech/pulumi-awstags/sdk/v3/go/awstags",
				},
			},
		},
	})
}
