using System.Collections.Generic;
using System.Linq;
using Pulumi;
using Awstags = Pulumi.Awstags;

return await Deployment.RunAsync(() => 
{
    var myRandomResource = new Awstags.Random("myRandomResource", new()
    {
        Length = 24,
    });

    return new Dictionary<string, object?>
    {
        ["output"] = 
        {
            { "value", myRandomResource.Result },
        },
    };
});

