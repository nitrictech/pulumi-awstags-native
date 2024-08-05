import * as pulumi from "@pulumi/pulumi";
import * as awstags from "@pulumi/awstags";

const myRandomResource = new awstags.Random("myRandomResource", {length: 24});
export const output = {
    value: myRandomResource.result,
};
