import pulumi
import pulumi_awstags as awstags

my_random_resource = awstags.Random("myRandomResource", length=24)
pulumi.export("output", {
    "value": my_random_resource.result,
})
