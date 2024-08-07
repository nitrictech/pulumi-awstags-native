package main

import (
	"github.com/nitrictech/pulumi-awstags-native/sdk/v3/go/awstags/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		_, err := aws.NewResourceTag(ctx, "myResourceTag", &aws.ResourceTagArgs{
			ResourceARN: pulumi.String("arn:aws:s3:::myBucket"),
			Tag: &aws.TagArgs{
				Key:   pulumi.String("myTagKey"),
				Value: pulumi.String("myTagValue"),
			},
		})

		if err != nil {
			return err
		}
		return nil
	})
}
