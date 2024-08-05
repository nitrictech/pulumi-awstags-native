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
	awsArn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
)

// Version is initialized by the Go linker to contain the semver of this build.
var Version string

const Name string = "awstags"

var tagClients map[string]*resourcegroupstaggingapi.ResourceGroupsTaggingAPI

func Provider() p.Provider {
	// We tell the provider what resources it needs to support.
	return infer.Provider(infer.Options{
		Resources: []infer.InferredResource{
			infer.Resource[TagResources, TagResourcesArgs, TagResourcesState](),
		},
		ModuleMap: map[tokens.ModuleName]tokens.ModuleName{
			"provider": "index",
		},
	})
}

// Each resource has a controlling struct.
// Resource behavior is determined by implementing methods on the controlling struct.
// The `Create` method is mandatory, but other methods are optional.
// - Check: Remap inputs before they are typed.
// - Diff: Change how instances of a resource are compared.
// - Update: Mutate a resource in place.
// - Read: Get the state of a resource from the backing provider.
// - Delete: Custom logic when the resource is deleted.
// - Annotate: Describe fields and set defaults for a resource.
// - WireDependencies: Control how outputs and secrets flows through values.
type TagResources struct{}

type TagResourcesArgs struct {
	ResourceARNList []string          `pulumi:"resourceARNList"`
	Tags            map[string]string `pulumi:"tags"`
}

type TagResourcesState struct {
	TagResourcesArgs
}

// All resources must implement Create at a minimum.
func (TagResources) Create(ctx p.Context, name string, input TagResourcesArgs, preview bool) (string, TagResourcesState, error) {
	state := TagResourcesState{TagResourcesArgs: input}
	if preview {
		return name, state, nil
	}

	addTags(input.ResourceARNList, input.Tags)

	return name, state, nil
}

func (TagResources) Delete(ctx p.Context, name string, state TagResourcesState, preview bool) error {
	if preview {
		return nil
	}

	removeTags(state.ResourceARNList, getTagKeys(state.Tags))

	return nil
}

func (TagResources) Update(ctx p.Context, name string, old, new TagResourcesState, preview bool) error {
	// Find tags that need to be removed.
	removedTagKeys := []string{}
	for k := range old.Tags {
		if _, ok := new.Tags[k]; !ok {
			removedTagKeys = append(removedTagKeys, k)
		}
	}

	// Find the difference between the old and new ARNs.
	keptArns := make([]string, 0)
	removedArns := make([]string, 0)
	for _, arn := range old.ResourceARNList {
		if !contains(new.ResourceARNList, arn) {
			removedArns = append(removedArns, arn)
		} else {
			keptArns = append(keptArns, arn)
		}
	}

	// Remove existing tags from removed ARNs.
	if err := removeTags(removedArns, getTagKeys(old.Tags)); err != nil {
		return err
	}

	// Remove removed tags from kept ARNs.
	if err := removeTags(keptArns, removedTagKeys); err != nil {
		return err
	}

	// Add desired tags/values to kept/new ARNs.
	return addTags(new.ResourceARNList, new.Tags)
}

func removeTags(arns []string, tagKeys []string) error {
	// Group ARNs by region so we can make a single call to each region.
	arnsByRegion := groupArnsByRegion(arns)

	for region, arns := range arnsByRegion {
		tagClient, err := getTaggingClient(region)
		if err != nil {
			return err
		}

		_, err = tagClient.UntagResources(&resourcegroupstaggingapi.UntagResourcesInput{
			ResourceARNList: aws.StringSlice(arns),
			TagKeys:         aws.StringSlice(tagKeys),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func addTags(arns []string, tags map[string]string) error {
	// Group ARNs by region so we can make a single call to each region.
	arnsByRegion := groupArnsByRegion(arns)

	for region, arns := range arnsByRegion {
		tagClient, err := getTaggingClient(region)
		if err != nil {
			return err
		}

		_, err = tagClient.TagResources(&resourcegroupstaggingapi.TagResourcesInput{
			ResourceARNList: aws.StringSlice(arns),
			Tags:            aws.StringMap(tags),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func groupArnsByRegion(arns []string) map[string][]string {
	arnsByRegion := make(map[string][]string)
	for _, arnString := range arns {
		arn, err := awsArn.Parse(arnString)
		if err != nil {
			return nil
		}

		arnsByRegion[arn.Region] = append(arnsByRegion[arn.Region], arn.String())
	}

	return arnsByRegion
}

func getTagKeys(tags map[string]string) []string {
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}

	return keys
}

func getTaggingClient(region string) (*resourcegroupstaggingapi.ResourceGroupsTaggingAPI, error) {
	if tagClients == nil {
		tagClients = make(map[string]*resourcegroupstaggingapi.ResourceGroupsTaggingAPI)
	}

	if client, ok := tagClients[region]; ok {
		return client, nil
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	}))

	tagClients[region] = resourcegroupstaggingapi.New(sess)

	return tagClients[region], nil
}

func contains(arr []string, target string) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}

	return false
}
