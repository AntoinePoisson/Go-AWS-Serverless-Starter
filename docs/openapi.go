// Package docs is the OpenAPI general info, nothing else. Nobody imports it.
// Run `make docs` after touching an annotation.
//
// Lives outside both functions because `swag init -g` takes exactly one file.
package docs

//	@title			Go AWS Serverless Starter API
//	@version		1.0
//	@license.name	MIT
//	@license.url	https://opensource.org/license/mit
//	@description	Item resource served by two Lambda functions behind a single HTTP API.
//	@description
//	@description	The /items routes are authenticated with the X-Api-Key header and are
//	@description	served by the `api` function. /health and /public/... are open and are
//	@description	served by the `public` function. Both sit behind the same API Gateway,
//	@description	which is why the public read is namespaced under /public.
//	@description
//	@description				Errors carry a JSON body of the shape {"code": "...", "message": "..."}.
//
//	@externalDocs.url			https://github.com/AntoinePoisson/go-aws-serverless-starter
//	@externalDocs.description	Source repository and README
//
//	@servers.url				http://localhost:8080
//	@servers.description		Local api function
//
//	@servers.url				http://localhost:8081
//	@servers.description		Local public function
//
//	@servers.url				https://{apiId}.execute-api.{region}.amazonaws.com
//	@servers.description		Deployed stage, both functions behind one API Gateway
//	@servers.variables.default	apiId	xxxxxxxxxx
//	@servers.variables.default	region	eu-west-1
//
// Tags have to come before @securityDefinitions. Its description swallows
// everything after it and they just vanish, no warning.
//
//	@tag.name					items
//	@tag.description			Authenticated CRUD over the item resource.
//	@tag.name					public
//	@tag.description			Unauthenticated reads, namespaced under /public.
//	@tag.name					health
//	@tag.description			Liveness, stage and version.
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-Api-Key
//	@description				Shared secret. Deployed stages read it from SSM Parameter Store.
//
// Version above is the contract, not a release. Release Please never
// touches this file. Bump it when you break a consumer.
//
// Dont start a prose line with @. swaggo treats it as an annotation
// and folds it into the block above.
