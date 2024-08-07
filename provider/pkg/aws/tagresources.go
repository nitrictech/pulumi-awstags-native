package aws

import (
	awsArn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/jyecusch/pulumi-awstags-native/provider/pkg/mutex"
	p "github.com/pulumi/pulumi-go-provider"
)

var tagClients map[string]*resourcegroupstaggingapi.ResourceGroupsTaggingAPI

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

	release, err := mutex.BorrowTags(input.ResourceARNList, getTagKeys(input.Tags))
	if err != nil {
		return "", state, err
	}

	if preview {
		release(true)
		return name, state, nil
	}
	addTags(input.ResourceARNList, input.Tags)

	release(true)

	return name, state, nil
}

func (TagResources) Delete(ctx p.Context, name string, state TagResourcesState, preview bool) error {
	release, err := mutex.BorrowTags(state.ResourceARNList, getTagKeys(state.Tags))
	if err != nil {
		return err
	}

	if preview {
		release(false)
		return nil
	}

	removeTags(state.ResourceARNList, getTagKeys(state.Tags))

	release(false)

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

	releaseRemoved, err := mutex.BorrowTags(removedArns, getTagKeys(old.Tags))
	if err != nil {
		return err
	}
	releaseKept, err := mutex.BorrowTags(keptArns, removedTagKeys)
	if err != nil {
		return err
	}
	releaseDesired, err := mutex.BorrowTags(new.ResourceARNList, getTagKeys(new.Tags))
	if err != nil {
		return err
	}

	if preview {
		releaseRemoved(false)
		releaseKept(false)
		releaseDesired(true)
		return nil
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
	if err := addTags(new.ResourceARNList, new.Tags); err != nil {
		return err
	}

	releaseRemoved(false)
	releaseKept(false)
	releaseDesired(true)

	return nil
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
