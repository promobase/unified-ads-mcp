This is repo for unified-ads-mcp server. It's a monorepo of golang MCP server implementations of Google Ads, Facebook Business API, TikTok Business API.
It exposes api for LLMs to CRUD on ads, campaigns, manage, and optimze them. On high level, we incorporate the specs from the source of truth, for FB it's JSON api specs, wrapping the underlying graph api. For tiktok, they have openapi yaml schemas, for google ads, it's protocol buffer definitions. We implement custom codegen scripts that transforms these endpoints to MCP tools, at the same time handles auth.

We'll start with facebook business api, then expand to tiktok business api & google ads as complexity scales.

## Resources
1. [mcp-go](https://raw.githubusercontent.com/mark3labs/mcp-go/refs/heads/main/README.md) this includes the readme for how to create Model Context Protocal Servers in golang.
2. [facebook-business-marketing-api-overview](https://developers.facebook.com/docs/marketing-apis/overview) Marketing api is subset of facebook business sdk, it's dedicated for ads mgmt, insights, campaigns etc. We prioritize supporthing these first.
3. [api-reference-v23.0](https://developers.facebook.com/docs/marketing-api/reference/v23.0) It's wrapping the underlying graph api, where we have nodes & edges. The JSON api specs are stored in `internal/facebook/api_specs/specs` dir, which has directory of .JSON specs. each one defines "apis" and "fields", apis are the endpoints of different methods, and fields are the object of the entity, e.g. we have Ad.json, AdAccount.json, etc. Fields mapping could be referring to other api specs, so we need to handle them properly. NOTE that there's a special `enum_types.json` in the specs dir, it contains all the ENUM definitions. Which might be used in the field def across repo, so handle it first.
4. [complete example of using mcp-go](https://raw.githubusercontent.com/mark3labs/mcp-go/refs/heads/main/examples/everything/main.go)
5. repo of the coden sdk path`/Users/ruizeli/dev/promobase/facebook-business-sdk-codegen` this is the repo of the codegen. it does not have src code, but it has the specs(we already have it), as well as templates for python, node, for reference.


## ----- Current Task -----
we're in a new repo and start implementing mcp servers for facebook business api. refer to the api specs, i wanna implement codegen that generates MCP tools for each of the adobjects in the spec. in a type safe manner. 

## notes
1. we use makefile to handle codegen, build, run. make sure your changes are validated.
2. when using 3P libraries, always look up the docs/example code on using it, e.g. mcp-go docs can be done using `go doc ..`, etc. 
