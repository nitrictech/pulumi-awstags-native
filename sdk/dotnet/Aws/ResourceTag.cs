// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.Awstags.Aws
{
    [AwstagsResourceType("awstags:aws:ResourceTag")]
    public partial class ResourceTag : global::Pulumi.CustomResource
    {
        [Output("resourceARN")]
        public Output<string> ResourceARN { get; private set; } = null!;

        [Output("tag")]
        public Output<Outputs.Tag> Tag { get; private set; } = null!;


        /// <summary>
        /// Create a ResourceTag resource with the given unique name, arguments, and options.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resource</param>
        /// <param name="args">The arguments used to populate this resource's properties</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public ResourceTag(string name, ResourceTagArgs args, CustomResourceOptions? options = null)
            : base("awstags:aws:ResourceTag", name, args ?? new ResourceTagArgs(), MakeResourceOptions(options, ""))
        {
        }

        private ResourceTag(string name, Input<string> id, CustomResourceOptions? options = null)
            : base("awstags:aws:ResourceTag", name, null, MakeResourceOptions(options, id))
        {
        }

        private static CustomResourceOptions MakeResourceOptions(CustomResourceOptions? options, Input<string>? id)
        {
            var defaultOptions = new CustomResourceOptions
            {
                Version = Utilities.Version,
                PluginDownloadURL = "https://github.com/nitrictech/pulumi-awstags-native/releases/download/v0.0.1-alpha.1723004377+3996998c.dirty/pulumi-awstags-v0.0.1-alpha.1723004377+3996998c.dirty.tgz",
            };
            var merged = CustomResourceOptions.Merge(defaultOptions, options);
            // Override the ID if one was specified for consistency with other language SDKs.
            merged.Id = id ?? merged.Id;
            return merged;
        }
        /// <summary>
        /// Get an existing ResourceTag resource's state with the given name, ID, and optional extra
        /// properties used to qualify the lookup.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resulting resource.</param>
        /// <param name="id">The unique provider ID of the resource to lookup.</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public static ResourceTag Get(string name, Input<string> id, CustomResourceOptions? options = null)
        {
            return new ResourceTag(name, id, options);
        }
    }

    public sealed class ResourceTagArgs : global::Pulumi.ResourceArgs
    {
        [Input("resourceARN", required: true)]
        public Input<string> ResourceARN { get; set; } = null!;

        [Input("tag", required: true)]
        public Input<Inputs.TagArgs> Tag { get; set; } = null!;

        public ResourceTagArgs()
        {
        }
        public static new ResourceTagArgs Empty => new ResourceTagArgs();
    }
}
