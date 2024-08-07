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
type ResourceTag struct{}

type tag struct {
	Key   string
	Value string
}

type ResourceTagArgs struct {
	ResourceARN string `pulumi:"resourceARN"`
	Tag         tag    `pulumi:"tag"`
}

type ResourceTagState struct {
	ResourceTagArgs
}

// All resources must implement Create at a minimum.
func (ResourceTag) Create(ctx p.Context, name string, input ResourceTagArgs, preview bool) (string, ResourceTagState, error) {
	state := ResourceTagState{ResourceTagArgs: input}

	release, err := mutex.BorrowTag(input.ResourceARN, input.Tag.Key)
	if err != nil {
		return "", state, err
	}

	if preview {
		release(true)
		return name, state, nil
	}
	addTag(input.ResourceARN, input.Tag)

	release(true)

	return name, state, nil
}

func (ResourceTag) Delete(ctx p.Context, name string, state ResourceTagState, preview bool) error {
	release, err := mutex.BorrowTag(state.ResourceARN, state.Tag.Key)
	if err != nil {
		// A write operation has already been registered for the tag on the ARN. So deletion isn't needed, the write operation will handle it.
		return nil
	}

	if preview {
		release(false)
		return nil
	}

	removeTag(state.ResourceARN, state.Tag.Key)

	release(false)

	return nil
}

func (ResourceTag) Update(ctx p.Context, name string, old, new ResourceTagState, preview bool) error {
	if new.ResourceARN != old.ResourceARN || new.Tag.Key != old.Tag.Key {
		release, err := mutex.BorrowTag(old.ResourceARN, old.Tag.Key)
		// Remove can be skipped if a write operation has already been registered for the tag on the ARN.
		if err == nil {
			removeTag(old.ResourceARN, old.Tag.Key)
			release(false)
		}
	}

	release, err := mutex.BorrowTag(new.ResourceARN, new.Tag.Key)
	if err != nil {
		return err
	}

	if preview {
		release(true)
		return nil
	}

	addTag(new.ResourceARN, new.Tag)

	release(true)

	return nil
}

func removeTag(arn string, tagKey string) error {
	region, err := getRegion(arn)
	if err != nil {
		return err
	}

	tagClient, err := getTaggingClient(region)
	if err != nil {
		return err
	}

	_, err = tagClient.UntagResources(&resourcegroupstaggingapi.UntagResourcesInput{
		ResourceARNList: aws.StringSlice([]string{arn}),
		TagKeys:         aws.StringSlice([]string{tagKey}),
	})
	if err != nil {
		return err
	}

	return nil
}

func addTag(arn string, tag tag) error {
	// Group ARNs by region so we can make a single call to each region.
	region, err := getRegion(arn)
	if err != nil {
		return err
	}

	tagClient, err := getTaggingClient(region)
	if err != nil {
		return err
	}

	_, err = tagClient.TagResources(&resourcegroupstaggingapi.TagResourcesInput{
		ResourceARNList: aws.StringSlice([]string{arn}),
		Tags:            aws.StringMap(map[string]string{tag.Key: tag.Value}),
	})
	if err != nil {
		return err
	}

	return nil
}

func getRegion(arnString string) (string, error) {
	arn, err := awsArn.Parse(arnString)
	if err != nil {
		return "", err
	}

	return arn.Region, nil
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
